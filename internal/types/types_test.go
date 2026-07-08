package types_test

import (
	"go/token"
	"go/types"
	"testing"

	rtypes "github.com/clear-street/reinforcer/internal/types"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

// Fixed-size arrays have no case in ToType's switch and don't implement its
// local `named` fallback interface, so this is a type ToType can't convert.
func newUnsupportedType() types.Type {
	return types.NewArray(types.Typ[types.String], 3)
}

func TestToType(t *testing.T) {
	fakePkg := types.NewPackage("github.com/clear-street/fake", "fake")

	ifaceSig := types.NewSignatureType(nil, nil, nil,
		types.NewTuple(),
		types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])),
		false,
	)
	ifaceUnderlying := types.NewInterfaceType([]*types.Func{
		types.NewFunc(token.NoPos, fakePkg, "DoTheThing", ifaceSig),
	}, nil)
	ifaceUnderlying.Complete()
	namedIface := types.NewNamed(types.NewTypeName(token.NoPos, fakePkg, "MyInterface", nil), ifaceUnderlying, nil)

	type args struct {
		t        types.Type
		variadic bool
	}
	tests := []struct {
		name string
		args args
		want jen.Code
	}{
		{
			name: "[]string (non-variadic slice)",
			args: args{t: types.NewSlice(types.Typ[types.String])},
			want: jen.Index().Add(jen.Id("string")),
		},
		{
			name: "named interface with a package path",
			args: args{t: namedIface},
			want: jen.Qual("github.com/clear-street/fake", "MyInterface"),
		},
		{
			name: "func(string) with no return values",
			args: args{t: types.NewSignatureType(nil, nil, nil,
				types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])),
				types.NewTuple(),
				false,
			)},
			want: jen.Func().Params(jen.Id("string")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rtypes.ToType(tt.args.t, tt.args.variadic)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestToType_ConversionErrors(t *testing.T) {
	unsupported := newUnsupportedType()

	fakePkg := types.NewPackage("github.com/clear-street/fake", "fake")
	genericOrig := types.NewNamed(types.NewTypeName(token.NoPos, fakePkg, "genericType", nil), types.NewStruct(nil, nil), nil)
	genericWithBadArg, _ := types.Instantiate(types.NewContext(), genericOrig, []types.Type{unsupported}, false)

	tests := []struct {
		name     string
		t        types.Type
		variadic bool
	}{
		{
			name: "*types.Named: generic type-arg conversion failure",
			t:    genericWithBadArg,
		},
		{
			name: "*types.Pointer: elem conversion failure",
			t:    types.NewPointer(unsupported),
		},
		{
			name: "*types.Map: key conversion failure",
			t:    types.NewMap(unsupported, types.Typ[types.String]),
		},
		{
			name: "*types.Map: elem conversion failure",
			t:    types.NewMap(types.Typ[types.String], unsupported),
		},
		{
			name: "*types.Chan: elem conversion failure",
			t:    types.NewChan(types.SendRecv, unsupported),
		},
		{
			name: "*types.Signature: param conversion failure",
			t: types.NewSignatureType(nil, nil, nil,
				types.NewTuple(types.NewVar(token.NoPos, nil, "", unsupported)),
				types.NewTuple(),
				false,
			),
		},
		{
			name: "*types.Signature: return conversion failure",
			t: types.NewSignatureType(nil, nil, nil,
				types.NewTuple(),
				types.NewTuple(types.NewVar(token.NoPos, nil, "", unsupported)),
				false,
			),
		},
		{
			name:     "variadicToType: slice elem conversion failure",
			t:        types.NewSlice(unsupported),
			variadic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rtypes.ToType(tt.t, tt.variadic)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}
