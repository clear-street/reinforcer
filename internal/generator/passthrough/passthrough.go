package passthrough

import (
	"github.com/clear-street/reinforcer/internal/generator/method"
	"github.com/dave/jennifer/jen"
)

// PassThrough is a code generator that injects no middleware and acts a simple fall through call to the delegate
type PassThrough struct {
	method         *method.Method
	structName     string
	structTypeArgs []jen.Code
	receiverName   string
}

// NewPassThrough is a ctor for PassThrough
func NewPassThrough(method *method.Method, structName string, structTypeArgs []jen.Code, receiverName string) *PassThrough {
	return &PassThrough{
		method:         method,
		structName:     structName,
		structTypeArgs: structTypeArgs,
		receiverName:   receiverName,
	}
}

// Statement generates the jen.Statement for this retryable method
func (p *PassThrough) Statement() (*jen.Statement, error) {
	methodArgParams := p.method.ParametersNameAndType
	params := p.method.Parameters()
	delegateCall := jen.Id(p.receiverName).Dot("delegate").Dot(p.method.Name).Call(params...)

	var block []jen.Code
	if len(p.method.ReturnTypes) > 0 {
		block = append(block, jen.Return(delegateCall))
	} else {
		block = append(block, delegateCall)
	}

	return jen.Func().Params(jen.Id(p.receiverName).Op("*").Id(p.structName).Types(p.structTypeArgs...)).Id(p.method.Name).Call(methodArgParams...).Block(
		block...,
	), nil
}
