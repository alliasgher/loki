package middleware

import (
	"context"
	"sync/atomic"
	"time"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	internalcontext "github.com/aws/aws-sdk-go-v2/internal/context"
	"github.com/aws/smithy-go/middleware"
)

// AddTimeOffsetMiddleware is deprecated.
//
// Deprecated: handled in retry loop.
type AddTimeOffsetMiddleware struct {
	Offset *atomic.Int64
}

// ID the identifier for AddTimeOffsetMiddleware
func (m *AddTimeOffsetMiddleware) ID() string { return "AddTimeOffsetMiddleware" }

// HandleBuild sets the attempt skew from the client offset.
func (m AddTimeOffsetMiddleware) HandleBuild(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
	out middleware.BuildOutput, metadata middleware.Metadata, err error,
) {
	if m.Offset != nil {
		ctx = internalcontext.SetAttemptSkewContext(ctx, time.Duration(m.Offset.Load()))
	}
	return next.HandleBuild(ctx, in)
}

// HandleDeserialize stores the attempt skew on the client offset.
func (m *AddTimeOffsetMiddleware) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	out, metadata, err = next.HandleDeserialize(ctx, in)
	if m.Offset != nil {
		if v, ok := awsmiddleware.GetAttemptSkew(metadata); ok {
			m.Offset.Store(v.Nanoseconds())
		}
	}
	return out, metadata, err
}
