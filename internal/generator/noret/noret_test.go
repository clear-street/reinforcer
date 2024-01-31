package noret_test

import (
	"bytes"
	"go/token"
	"go/types"
	"testing"

	"github.com/clear-street/reinforcer/internal/generator/method"
	"github.com/clear-street/reinforcer/internal/generator/noret"
	rtypes "github.com/clear-street/reinforcer/internal/types"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

func TestNoReturn_Statement(t *testing.T) {
	ctxVar := types.NewVar(token.NoPos, nil, "ctx", rtypes.ContextType)

	tests := []struct {
		name           string
		methodName     string
		structTypeArgs []jen.Code
		signature      *types.Signature
		want           string
		wantErr        bool
	}{
		{
			name:       "MyFunction()",
			methodName: "MyFunction",
			signature:  types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(), false),
			want: `func (r *Resilient) MyFunction() {
	err := r.run(context.Background(), ResilientMethods.MyFunction, func(_ context.Context) error {
		r.delegate.MyFunction()
		return nil
	})
	if err != nil {
		panic(err)
	}
}`,
			wantErr: false,
		},
		{
			name:       "MyFunction(ctx context.Context, arg1 string)",
			methodName: "MyFunction",
			signature: types.NewSignatureType(nil, nil, nil, types.NewTuple(
				ctxVar,
				types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
			), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false),
			want: `func (r *Resilient) MyFunction(ctx context.Context, arg1 string) {
	err := r.run(ctx, ResilientMethods.MyFunction, func(ctx context.Context) error {
		r.delegate.MyFunction(ctx, arg1)
		return nil
	})
	if err != nil {
		panic(err)
	}
}`,
			wantErr: false,
		},
		{
			name:           "struct type args",
			methodName:     "MyFunction",
			structTypeArgs: []jen.Code{jen.Id("T")},
			signature:      types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(), false),
			want: `func (r *Resilient[T]) MyFunction() {
	err := r.run(context.Background(), ResilientMethods.MyFunction, func(_ context.Context) error {
		r.delegate.MyFunction()
		return nil
	})
	if err != nil {
		panic(err)
	}
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := method.ParseMethod(tt.methodName, tt.signature)
			require.NoError(t, err)
			ret := noret.NewNoReturn(m, "Resilient", tt.structTypeArgs, "r")
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
