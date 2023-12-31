package method

import (
	"fmt"
	"go/types"

	rtypes "github.com/clear-street/reinforcer/internal/types"
	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
)

const (
	ctxVarName = "ctx"
)

type named interface {
	Name() string
}

// Method holds all of the data for code generation on a specific method signature
type Method struct {
	Name                  string
	HasContext            bool
	ReturnsError          bool
	HasVariadic           bool
	ParameterNames        []string
	ParametersNameAndType []jen.Code
	ReturnTypes           []jen.Code
	ContextParameter      *int
	ReturnErrorIndex      *int
}

// ConstantRef is the reference to the constant for this method's name
func (m *Method) ConstantRef(parentTypeName string) jen.Code {
	constantsStructName := fmt.Sprintf("%sMethods", parentTypeName)
	return jen.Id(constantsStructName).Dot(m.Name)
}

// ContextParam generates the param name and type for a context arg for the given method
func (m *Method) ContextParam() (ctxParamName string, ctxParam jen.Code) {
	ctxParamName = ctxVarName
	if m.HasContext {
		// Passes down the context if one is present in the signature
		ctxParam = jen.Id(ctxVarName)
	} else {
		// Use context.Background() if no context is present in signature
		ctxParam = jen.Qual("context", "Background").Call()
		ctxParamName = "_"
	}
	return
}

// Parameters generates code for parameter names to be used in codegen
func (m *Method) Parameters() []jen.Code {
	var params []jen.Code
	for i, j := 0, len(m.ParameterNames)-1; i < len(m.ParameterNames); i++ {
		if m.HasVariadic && i == j {
			params = append(params, jen.Id(m.ParameterNames[i]).Op("..."))
		} else {
			params = append(params, jen.Id(m.ParameterNames[i]))
		}
	}
	return params
}

// MustParseMethod parses the given types.Signature and generates a Method, if there's an error this method will panic
func MustParseMethod(name string, signature *types.Signature) *Method {
	m, err := ParseMethod(name, signature)
	if err != nil {
		panic(err)
	}
	return m
}

// ParseMethod parses the given types.Signature and generates a Method
func ParseMethod(name string, signature *types.Signature) (*Method, error) {
	m := &Method{
		Name:             name,
		ReturnErrorIndex: nil,
		ContextParameter: nil,
		HasVariadic:      signature.Variadic(),
	}

	isVariadic := signature.Variadic()
	numParams := signature.Params().Len()
	for i, lastIndex := 0, numParams-1; i < numParams; i++ {
		param := signature.Params().At(i)
		if rtypes.IsContextType(param.Type()) {
			m.HasContext = true
			m.ContextParameter = new(int)
			*m.ContextParameter = i
			m.ParametersNameAndType = append(m.ParametersNameAndType, jen.Id(ctxVarName).Add(jen.Qual("context", "Context")))
			m.ParameterNames = append(m.ParameterNames, ctxVarName)
		} else {
			paramName := fmt.Sprintf("arg%d", i)

			paramType, err := toType(param.Type(), isVariadic && i == lastIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to convert type=%v; error=%w", param.Type(), err)
			}
			m.ParametersNameAndType = append(m.ParametersNameAndType, jen.Id(paramName).Add(paramType))
			m.ParameterNames = append(m.ParameterNames, paramName)
		}
	}
	for i := 0; i < signature.Results().Len(); i++ {
		res := signature.Results().At(i)
		resType, err := toType(res.Type(), false)
		if err != nil {
			panic(err)
		}
		if rtypes.IsErrorType(res.Type()) {
			if m.ReturnErrorIndex != nil {
				return nil, fmt.Errorf("multiple errors returned by method signature")
			}
			m.ReturnsError = true
			m.ReturnErrorIndex = new(int)
			*m.ReturnErrorIndex = i
		}
		m.ReturnTypes = append(m.ReturnTypes, resType)
	}
	return m, nil
}

// variadicToType generates the representation for a variadic type "...MyType"
func variadicToType(t types.Type) (jen.Code, error) {
	sliceType, ok := t.(*types.Slice)
	if !ok {
		return nil, fmt.Errorf("expected type to be *types.Slice, got=%T", t)
	}
	sliceElemType, err := toType(sliceType.Elem(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert slice's type; error=%w", err)
	}
	return jen.Op("...").Add(sliceElemType), nil
}

// toType generates the representation for the given type
func toType(t types.Type, variadic bool) (jen.Code, error) {
	if variadic {
		return variadicToType(t)
	}

	switch v := t.(type) {
	case *types.Basic:
		return jen.Id(v.Name()), nil
	case *types.Chan:
		rt, err := toType(v.Elem(), false)
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
		return jen.Qual(
			pkgPath,
			typeName.Name(),
		), nil
	case *types.Pointer:
		rt, err := toType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return jen.Op("*").Add(rt), nil
	case *types.Interface:
		return jen.Id("interface{}"), nil
	case *types.Slice:
		elemType, err := toType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return jen.Index().Add(elemType), nil
	case named:
		return jen.Id(v.Name()), nil
	case *types.Map:
		keyType, err := toType(v.Key(), false)
		if err != nil {
			return nil, err
		}
		elemType, err := toType(v.Elem(), false)
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
			tt, err := toType(paramType, lastIndex == p && fnVariadic)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert type %v", paramType)
			}
			paramTypes = append(paramTypes, tt)
		}

		var returnTypes []jen.Code
		for r := 0; r < v.Results().Len(); r++ {
			returnType := v.Results().At(r).Type()
			tt, err := toType(returnType, false)
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
	default:
		return nil, fmt.Errorf("type not handled: %T", v)
	}
}
