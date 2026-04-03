package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnaryOpString(t *testing.T) {
	for _, tc := range []struct {
		name string
		op   UnaryOp
		want string
	}{
		{name: "not", op: UnaryOpNot, want: "NOT"},
		{name: "abs", op: UnaryOpAbs, want: "ABS"},
		{name: "cast_float", op: UnaryOpCastFloat, want: "CAST_FLOAT"},
		{name: "cast_bytes", op: UnaryOpCastBytes, want: "CAST_BYTES"},
		{name: "cast_duration", op: UnaryOpCastDuration, want: "CAST_DURATION"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, tc.op.String())
		})
	}
}
