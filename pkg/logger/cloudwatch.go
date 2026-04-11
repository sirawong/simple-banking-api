package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

// cloudWatchHandler is a slog.Handler that ships log records to AWS CloudWatch Logs.
//
// Each Handle call sends a single PutLogEvents request. For high-throughput
// services consider batching via a channel, but for most banking API use cases
// this is sufficient and keeps the implementation simple.
type cloudWatchHandler struct {
	client   *cloudwatchlogs.Client
	group    string
	stream   string
	minLevel slog.Level
	preAttrs []slog.Attr // attrs collected via WithAttrs
	groups   []string    // group names collected via WithGroup
	mu       sync.Mutex
}

func newCloudWatchHandler(ctx context.Context, opts Options, level slog.Level) (*cloudWatchHandler, error) {
	if opts.CloudWatchGroup == "" {
		return nil, fmt.Errorf("CLOUDWATCH_LOG_GROUP must be set when LOG_OUTPUT=cloudwatch")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(opts.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := cloudwatchlogs.NewFromConfig(awsCfg)

	// Ensure log group exists
	if _, err = client.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(opts.CloudWatchGroup),
	}); err != nil {
		var already *types.ResourceAlreadyExistsException
		if !errors.As(err, &already) {
			return nil, fmt.Errorf("create log group: %w", err)
		}
	}

	// Resolve stream name (default: hostname-YYYY-MM-DD)
	stream := opts.CloudWatchStream
	if stream == "" {
		hostname, _ := os.Hostname()
		stream = hostname + "-" + time.Now().UTC().Format("2006-01-02")
	}

	// Ensure log stream exists
	if _, err = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(opts.CloudWatchGroup),
		LogStreamName: aws.String(stream),
	}); err != nil {
		var already *types.ResourceAlreadyExistsException
		if !errors.As(err, &already) {
			return nil, fmt.Errorf("create log stream: %w", err)
		}
	}

	return &cloudWatchHandler{
		client:   client,
		group:    opts.CloudWatchGroup,
		stream:   stream,
		minLevel: level,
	}, nil
}

func (h *cloudWatchHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *cloudWatchHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := map[string]any{
		"time":    r.Time.UTC().Format(time.RFC3339Nano),
		"level":   r.Level.String(),
		"message": r.Message,
	}

	// Attrs collected via WithAttrs (handler-level)
	for _, a := range h.preAttrs {
		flattenAttr(entry, h.groups, a)
	}

	// Attrs on this specific record
	r.Attrs(func(a slog.Attr) bool {
		flattenAttr(entry, h.groups, a)
		return true
	})

	msg, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	ts := r.Time.UnixMilli()
	payload := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(h.group),
		LogStreamName: aws.String(h.stream),
		LogEvents: []types.InputLogEvent{
			{Message: aws.String(string(msg)), Timestamp: aws.Int64(ts)},
		},
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err = h.client.PutLogEvents(ctx, payload)
	return err
}

func (h *cloudWatchHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := h.clone()
	clone.preAttrs = append(clone.preAttrs, attrs...)
	return clone
}

func (h *cloudWatchHandler) WithGroup(name string) slog.Handler {
	clone := h.clone()
	clone.groups = append(clone.groups, name)
	return clone
}

func (h *cloudWatchHandler) clone() *cloudWatchHandler {
	preAttrs := make([]slog.Attr, len(h.preAttrs))
	copy(preAttrs, h.preAttrs)
	groups := make([]string, len(h.groups))
	copy(groups, h.groups)
	return &cloudWatchHandler{
		client:   h.client,
		group:    h.group,
		stream:   h.stream,
		minLevel: h.minLevel,
		preAttrs: preAttrs,
		groups:   groups,
	}
}

// flattenAttr writes a slog.Attr into the flat map, prefixing with any active groups.
func flattenAttr(dst map[string]any, groups []string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}

	key := a.Key
	if len(groups) > 0 {
		for i := len(groups) - 1; i >= 0; i-- {
			key = groups[i] + "." + key
		}
	}

	if a.Value.Kind() == slog.KindGroup {
		for _, sub := range a.Value.Group() {
			flattenAttr(dst, append(groups, a.Key), sub)
		}
		return
	}

	dst[key] = a.Value.Any()
}
