#!/bin/bash

set -e

MODULE_NAME=$(go list -m 2>/dev/null)
if [ -z "$MODULE_NAME" ]; then
    echo "Error: This directory is not a Go module."
    exit 1
fi

# ─────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────

# Read consecutive comment lines immediately above a declaration.
# Stops at any blank line or non-comment line.
_annotation_block() {
    local file="$1"
    local decl_line="$2"
    local result=""
    local i=$(( decl_line - 1 ))
    while [ "$i" -ge 1 ]; do
        local ln
        ln=$(sed -n "${i}p" "$file")
        case "$ln" in
            ""|"	") break ;;
            "//"*)  result="$ln
$result" ;;
            *)      break ;;
        esac
        i=$(( i - 1 ))
    done
    echo "$result"
}

has_ignore()      { _annotation_block "$1" "$2" | grep -q "@wire:ignore"; }
has_wire_struct() { _annotation_block "$1" "$2" | grep -q "@wire:struct"; }
has_default()     { _annotation_block "$1" "$2" | grep -qE '@wire:set\([^)]*,default[^)]*\)'; }

get_set_name() {
    _annotation_block "$1" "$2" \
        | grep -oE '@wire:set\([^)]+\)' \
        | grep -oE 'name=[^,)]+' \
        | sed 's/name=//' \
        | head -1
}

path_to_alias() { echo "$1" | tr '/' '_'; }

# ─────────────────────────────────────────────
# Process one di/ folder
# ─────────────────────────────────────────────

process_di_folder() {
    local di_dir="$1"

    local PARENT_DIR
    PARENT_DIR=$(dirname "$di_dir")
    [ "$PARENT_DIR" = "." ] && PARENT_DIR=""

    local OUTPUT_FILE="$di_dir/di_gen.go"
    local PARENT_BASE_NAME
    PARENT_BASE_NAME=$(basename "${PARENT_DIR:-$MODULE_NAME}")
    local DEFAULT_SET_NAME
    DEFAULT_SET_NAME="$(tr '[:lower:]' '[:upper:]' <<< ${PARENT_BASE_NAME:0:1})${PARENT_BASE_NAME:1}Set"

    local NESTED_DI
    NESTED_DI=$(find "${PARENT_DIR:-.}" -type d -name "di" \
        -not -path "$di_dir" \
        -not -path "$di_dir/*" \
        -not -path "*/vendor/*" | sed 's|^\./||' | sort -u)

    local EXCLUDE_ARGS="-not -path \"$di_dir/*\""
    local nd nd_parent
    for nd in $NESTED_DI; do
        nd_parent=$(dirname "$nd")
        EXCLUDE_ARGS="$EXCLUDE_ARGS -not -path \"$nd_parent/*\""
    done

    local GO_FILES
    GO_FILES=$(eval "find \"${PARENT_DIR:-.}\" -type f -name \"*.go\" $EXCLUDE_ARGS" 2>/dev/null || true)

    # Temp files
    # TMP_DEFAULT : "alias|expr"
    # TMP_NAMED   : "SetName|alias|expr"
    # TMP_IMPORTS : "alias|import_path"
    local TMP_DEFAULT TMP_NAMED TMP_IMPORTS
    TMP_DEFAULT=$(mktemp); TMP_NAMED=$(mktemp); TMP_IMPORTS=$(mktemp)

    local file pkg_dir alias import_path
    for file in $GO_FILES; do
        pkg_dir=$(dirname "$file" | sed 's|^\./||')
        [ "$pkg_dir" = "." ] && pkg_dir=""

        if [ -z "$pkg_dir" ]; then
            alias="app_root"; import_path="$MODULE_NAME"
        else
            alias=$(path_to_alias "$pkg_dir")
            import_path="$MODULE_NAME/$pkg_dir"
        fi
        echo "$alias|$import_path" >> "$TMP_IMPORTS"

        # ── Function providers: func ProvideXxx ──────────────────────────
        local func_line set_name func_name
        while IFS=: read -r func_line rest; do
            func_name=$(echo "$rest" | grep -oE 'Provide[a-zA-Z0-9]+')
            [ -z "$func_name" ] && continue
            has_ignore "$file" "$func_line" && continue

            set_name=$(get_set_name "$file" "$func_line")
            if [ -z "$set_name" ]; then
                echo "$alias|$alias.$func_name" >> "$TMP_DEFAULT"
            else
                echo "$set_name|$alias|$alias.$func_name" >> "$TMP_NAMED"
                has_default "$file" "$func_line" \
                    && echo "$alias|$alias.$func_name" >> "$TMP_DEFAULT"
            fi
        done < <(grep -nE "^func Provide[a-zA-Z0-9]+" "$file" 2>/dev/null || true)

        # ── Struct providers: @wire:struct ───────────────────────────────
        local struct_line struct_name wire_expr
        while IFS=: read -r struct_line rest; do
            struct_name=$(echo "$rest" | grep -oE 'type [A-Z][a-zA-Z0-9]+' | awk '{print $2}')
            [ -z "$struct_name" ] && continue
            has_wire_struct "$file" "$struct_line" || continue

            wire_expr="wire.Struct(new($alias.$struct_name), \"*\")"
            set_name=$(get_set_name "$file" "$struct_line")
            if [ -z "$set_name" ]; then
                echo "$alias|$wire_expr" >> "$TMP_DEFAULT"
            else
                echo "$set_name|$alias|$wire_expr" >> "$TMP_NAMED"
                has_default "$file" "$struct_line" \
                    && echo "$alias|$wire_expr" >> "$TMP_DEFAULT"
            fi
        done < <(grep -nE "^type [A-Z][a-zA-Z0-9]+ struct" "$file" 2>/dev/null || true)
    done

    # Skip if nothing to generate
    local has_content=0
    [ -s "$TMP_DEFAULT" ] && has_content=1
    [ -s "$TMP_NAMED" ]   && has_content=1
    [ -n "$NESTED_DI" ]   && has_content=1
    if [ "$has_content" -eq 0 ]; then
        rm -f "$TMP_DEFAULT" "$TMP_NAMED" "$TMP_IMPORTS"
        return 0
    fi

    echo "Generating: $OUTPUT_FILE"

    # Collect only aliases that are actually used
    local TMP_USED_ALIASES
    TMP_USED_ALIASES=$(mktemp)
    awk -F'|' '{print $1}' "$TMP_DEFAULT" >> "$TMP_USED_ALIASES"
    awk -F'|' '{print $2}' "$TMP_NAMED"  >> "$TMP_USED_ALIASES"
    sort -u "$TMP_USED_ALIASES" -o "$TMP_USED_ALIASES"

    {
        echo "// Code generated by gen_wire.sh; DO NOT EDIT."
        echo "package di"
        echo ""
        echo "import ("
        echo "    \"github.com/google/wire\""

        for nd in $NESTED_DI; do
            [ -z "$nd" ] && continue
            local nd_alias
            nd_alias=$(path_to_alias "$nd" | sed 's/_di$//')
            echo "    ${nd_alias}_di \"$MODULE_NAME/$nd\""
        done

        while IFS= read -r used_alias; do
            [ -z "$used_alias" ] && continue
            local imp
            imp=$(grep "^$used_alias|" "$TMP_IMPORTS" | head -1 | cut -d'|' -f2)
            [ -z "$imp" ] && continue
            echo "    $used_alias \"$imp\""
        done < "$TMP_USED_ALIASES"

        echo ")"
        echo ""

        # ── Default set ───────────────────────────────────────────────────
        echo "var $DEFAULT_SET_NAME = wire.NewSet("

        for nd in $NESTED_DI; do
            [ -z "$nd" ] && continue
            local nd_alias nd_parent nd_set
            nd_alias=$(path_to_alias "$nd" | sed 's/_di$//')
            nd_parent=$(basename "$(dirname "$nd")")
            nd_set="$(tr '[:lower:]' '[:upper:]' <<< ${nd_parent:0:1})${nd_parent:1}Set"
            echo "    ${nd_alias}_di.$nd_set,"
        done

        while IFS='|' read -r _ expr; do
            echo "    $expr,"
        done < "$TMP_DEFAULT"

        echo ")"

        # ── Named sets ────────────────────────────────────────────────────
        local all_set_names
        all_set_names=$(awk -F'|' '{print $1}' "$TMP_NAMED" | sort -u)

        for set_name in $all_set_names; do
            echo ""
            echo "var $set_name = wire.NewSet("
            grep "^$set_name|" "$TMP_NAMED" | while IFS='|' read -r _ _ expr; do
                echo "    $expr,"
            done
            echo ")"
        done

    } > "$OUTPUT_FILE"

    gofmt -w "$OUTPUT_FILE" 2>/dev/null || echo "Warning: gofmt failed on $OUTPUT_FILE"
    rm -f "$TMP_DEFAULT" "$TMP_NAMED" "$TMP_IMPORTS" "$TMP_USED_ALIASES"
}

# ─────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────

DI_FOLDERS=$(find . -type d -name "di" \
    -not -path "*/vendor/*" \
    -not -path "./cmd/*/di" \
    | sed 's|^\./||' | sort -r)

for di_dir in $DI_FOLDERS; do
    process_di_folder "${di_dir#./}"
done

# ── Run wire in injector directories ─────────────────────────────────────
if command -v wire &> /dev/null; then
    INJECTOR_DIRS=$(grep -rlE "//(go:build|\+build) wireinject" . \
        --include="*.go" --exclude-dir=vendor 2>/dev/null \
        | xargs -I {} dirname {} | sort -u || true)
    for dir in $INJECTOR_DIRS; do
        echo "Running wire in: $dir"
        (cd "$dir" && wire) || exit 1
    done
else
    echo "Warning: wire binary not found. Skipping wire generation."
    echo "Install with: go install github.com/google/wire/cmd/wire@latest"
fi
