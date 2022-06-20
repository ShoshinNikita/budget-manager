package reqid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// RequestID is a random hex string. It doesn't correspond to UUID RFC 4122 format because it would be
// an overkill for such small project. Also it can be easily modified to match RFC 4122
type RequestID string

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
func (reqID RequestID) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, reqID)
}
