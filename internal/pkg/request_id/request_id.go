package request_id

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// RequestID is a hex string with random content. It doesn't correspond UUID format based on RFC 4122
// because it is overkill for such small project. Also it can be easily modified to match RFC 4122
type RequestID string

func (r RequestID) ToString() string {
	return string(r)
}

const requestIDLength = 8

// New creates a new request id
func New() RequestID {
	data := make([]byte, requestIDLength/2)
	rand.Read(data) // nolint:errcheck
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

// ToContext
func ToContext(ctx context.Context, reqID RequestID) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, reqID)
}
