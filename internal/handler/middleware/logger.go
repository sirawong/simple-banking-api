package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sirawong/simple-banking-api/pkg/logger"
)

const (
	headerRequestID = "X-Request-ID"
	maxFieldLen     = 200
	maxBodyLog      = 4 * 1024 // 4 KB
)

var sensitiveKeys = map[string]bool{
	"password":     true,
	"passwordhash": true,
	"accesstoken":  true,
	"refreshtoken": true,
	"token":        true,
}

func truncate(s string) string {
	if len(s) <= maxFieldLen {
		return s
	}
	return s[:maxFieldLen] + "…"
}

func maskBody(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return truncate(string(raw))
	}
	for k := range m {
		if sensitiveKeys[strings.ToLower(k)] {
			m[k] = "[REDACTED]"
		}
	}
	masked, _ := json.Marshal(m)
	return string(masked)
}

type bodyWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	if w.buf.Len() < maxBodyLog {
		remaining := maxBodyLog - w.buf.Len()
		if len(b) > remaining {
			w.buf.Write(b[:remaining])
		} else {
			w.buf.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		requestID := c.GetHeader(headerRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header(headerRequestID, requestID)

		var reqBody string
		if c.Request.Body != nil {
			raw, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyLog))
			c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			reqBody = maskBody(raw)
		}

		bw := &bodyWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = bw

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		fields := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", truncate(c.Request.URL.Path),
			"query", truncate(c.Request.URL.RawQuery),
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"req_body", reqBody,
		}

		if status >= 400 {
			fields = append(fields, "res_body", maskBody(bw.buf.Bytes()))
		}

		switch {
		case status >= 500:
			log.Error("request", nil, fields...)
		case status >= 400:
			log.Warn("request", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}
