package method

import (
	"fmt"
	"go/types"

	rtypes "github.com/clear-street/reinforcer/internal/types"
	"github.com/dave/jennifer/jen"
)

const (
	ctxVarName = "ctx"
)

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

			paramType, err := rtypes.ToType(param.Type(), isVariadic && i == lastIndex)
			if err != nil {
				return nil, fmt.Errorf("failed to convert type=%v; error=%w", param.Type(), err)
			}
			m.ParametersNameAndType = append(m.ParametersNameAndType, jen.Id(paramName).Add(paramType))
			m.ParameterNames = append(m.ParameterNames, paramName)
		}
	}
	for i := 0; i < signature.Results().Len(); i++ {
		res := signature.Results().At(i)
		resType, err := rtypes.ToType(res.Type(), false)
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
