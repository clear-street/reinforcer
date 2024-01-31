package types

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

type named interface {
	Name() string
}

// ErrType is the types.Type for the error interface
var ErrType types.Type

// ContextType is the types.Type for the context.Context interface
var ContextType *types.Interface

func init() {
	errType := types.NewInterfaceType([]*types.Func{
		types.NewFunc(0, nil, "Error",
			types.NewSignatureType(
				nil,
				nil,
				nil,
				types.NewTuple(),
				types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.String])),
				false,
			),
		),
	}, nil)
	errType.Complete()
	ErrType = types.NewNamed(types.NewTypeName(0, nil, "error", nil), errType, nil)

	// Load the type definition for the Context type
	ctxPkg, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo,
	}, "context")
	if err != nil {
		panic(err)
	}
	ContextType = ctxPkg[0].Types.
		Scope().
		Lookup("Context").
		Type().(*types.Named).
		Underlying().(*types.Interface)
}

// IsErrorType determines if the given type implements the Error interface
func IsErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.Implements(t, ErrType.Underlying().(*types.Interface))
}

// IsContextType determines if the given type is context.Context
func IsContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	if t.String() == "context.Context" {
		return true
	}
	return types.Implements(t, ContextType)
}

// variadicToType generates the representation for a variadic type "...MyType"
func variadicToType(t types.Type) (jen.Code, error) {
	sliceType, ok := t.(*types.Slice)
	if !ok {
		return nil, fmt.Errorf("expected type to be *types.Slice, got=%T", t)
	}
	sliceElemType, err := ToType(sliceType.Elem(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert slice's type; error=%w", err)
	}
	return jen.Op("...").Add(sliceElemType), nil
}

// ToType generates the representation for the given type
func ToType(t types.Type, variadic bool) (jen.Code, error) {
	if variadic {
		return variadicToType(t)
	}

	switch v := t.(type) {
	case *types.Basic:
		return jen.Id(v.Name()), nil
	case *types.Chan:
		rt, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		switch v.Dir() {
		case types.SendRecv:
			return jen.Chan().Add(rt), nil
		case types.RecvOnly:
			return jen.Op("<-").Chan().Add(rt), nil
		default:
			return jen.Chan().Op("<-").Add(rt), nil
		}
	case *types.Named:
		typeName := v.Obj()
		if _, ok := v.Underlying().(*types.Interface); ok {
			if typeName.Pkg() != nil {
				pkgPath := typeName.Pkg().Path()
				return jen.Qual(
					pkgPath,
					typeName.Name(),
				), nil
			}
			return jen.Id(typeName.Name()), nil
		}
		pkgPath := typeName.Pkg().Path()
		var typeArgs []jen.Code
		for p := 0; p < v.TypeArgs().Len(); p++ {
			typeArg := v.TypeArgs().At(p)
			tt, err := ToType(typeArg, false)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert type %v", typeArg)
			}
			typeArgs = append(typeArgs, tt)
		}
		return jen.Qual(
			pkgPath,
			typeName.Name(),
		).Types(typeArgs...), nil
	case *types.Pointer:
		rt, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return jen.Op("*").Add(rt), nil
	case *types.Interface:
		return jen.Id("any"), nil
	case *types.Slice:
		elemType, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return jen.Index().Add(elemType), nil
	case named:
		return jen.Id(v.Name()), nil
	case *types.Map:
		keyType, err := ToType(v.Key(), false)
		if err != nil {
			return nil, err
		}
		elemType, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return jen.Map(keyType).Add(elemType), nil
	case *types.Signature:
		fnVariadic := v.Variadic()
		var paramTypes []jen.Code
		lastIndex := v.Params().Len() - 1
		for p := 0; p < v.Params().Len(); p++ {
			paramType := v.Params().At(p).Type()
			tt, err := ToType(paramType, lastIndex == p && fnVariadic)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert type %v", paramType)
			}
			paramTypes = append(paramTypes, tt)
		}

		var returnTypes []jen.Code
		for r := 0; r < v.Results().Len(); r++ {
			returnType := v.Results().At(r).Type()
			tt, err := ToType(returnType, false)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert type %v", returnType)
			}
			returnTypes = append(returnTypes, tt)
		}
		if len(returnTypes) == 0 {
			return jen.Func().Params(paramTypes...), nil
		}
		if len(returnTypes) > 1 {
			return jen.Func().Params(paramTypes...).Parens(jen.List(returnTypes...)), nil
		}
		return jen.Func().Params(paramTypes...).Add(returnTypes[0]), nil
	case *types.TypeParam:
		return jen.Id(v.Obj().Name()), nil
	default:
		return nil, fmt.Errorf("type not handled: %T", v)
	}
}
