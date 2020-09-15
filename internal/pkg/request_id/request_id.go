package reqid

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/sirupsen/logrus"
)

// RequestID is a random hex string. It doesn't correspond to UUID RFC 4122 format because it would be
// an overkill for such small project. Also it can be easily modified to match RFC 4122
type RequestID string

func (r RequestID) ToString() string {
	return string(r)
}

const requestIDLength = 8

// New creates a new request id
func New() RequestID {
	data := make([]byte, requestIDLength/2)
	rand.Read(data) //nolint:errcheck
	return RequestID(hex.EncodeToString(data))
}

type requestIDContextKey struct{}

// FromContext extracts request id from context.
// If request id doesn't exist, it generates a new one
func FromContext(ctx context.Context) RequestID {
	if reqID, ok := ctx.Value(requestIDContextKey{}).(RequestID); ok {
		return reqID
	}
	return New()
}

// ToContext returns a context based on passed one with injected request id
func ToContext(ctx context.Context, reqID RequestID) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, reqID)
}

const loggerFieldKey = "request_id"

// FromContextToLogger extracts request id from context and returns logger with added field
func FromContextToLogger(ctx context.Context, log logrus.FieldLogger) *logrus.Entry {
	reqID := FromContext(ctx)
	return log.WithField(loggerFieldKey, reqID)
}
