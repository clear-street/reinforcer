package passthrough_test

import (
	"bytes"
	"go/token"
	"go/types"
	"testing"

	"github.com/clear-street/reinforcer/internal/generator/method"
	"github.com/clear-street/reinforcer/internal/generator/passthrough"
	rtypes "github.com/clear-street/reinforcer/internal/types"
	"github.com/stretchr/testify/require"
)

func TestPassThrough_Statement(t *testing.T) {
	ctxVar := types.NewVar(token.NoPos, nil, "ctx", rtypes.ContextType)

	tests := []struct {
		name       string
		methodName string
		signature  *types.Signature
		want       string
		wantErr    bool
	}{
		{
			name:       "Function arguments and returns",
			methodName: "MyFunction",
			signature: types.NewSignatureType(nil, nil, nil, types.NewTuple(
				ctxVar,
				types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
			), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false),
			want: `func (r *resilient) MyFunction(ctx context.Context, arg1 string) {
	return r.delegate.MyFunction(ctx, arg1)
}`,
			wantErr: false,
		},
		{
			name:       "Function passes arguments",
			methodName: "MyFunction",
			signature: types.NewSignatureType(nil, nil, nil, types.NewTuple(
				ctxVar,
				types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
			), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false),
			want: `func (r *resilient) MyFunction(ctx context.Context, arg1 string) {
	return r.delegate.MyFunction(ctx, arg1)
}`,
			wantErr: false,
		},
		{
			name:       "Function no args no return",
			methodName: "MyFunction",
			signature:  types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(), false),
			want: `func (r *resilient) MyFunction() {
	r.delegate.MyFunction()
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := method.ParseMethod(tt.methodName, tt.signature)
			require.NoError(t, err)
			ret := passthrough.NewPassThrough(m, "resilient", "r")
			buf := &bytes.Buffer{}
			s, err := ret.Statement()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NotNil(t, s)
				require.NoError(t, err)
				renderErr := s.Render(buf)
				require.NoError(t, renderErr)

				got := buf.String()
				require.Equal(t, tt.want, got)
			}
		})
	}
}
