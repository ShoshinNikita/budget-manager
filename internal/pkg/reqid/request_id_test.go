package reqid

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	reqID := New()

	// Insert and extract request id
	ctx := ToContext(context.Background(), reqID)
	reqIDFromCtx := FromContext(ctx)
	require.Equal(t, reqID, reqIDFromCtx)

	// Extract from empty context
	_ = FromContext(context.Background())
}
