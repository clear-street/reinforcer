package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/clear-street/reinforcer/internal/generator/method"
	"github.com/clear-street/reinforcer/internal/generator/noret"
	"github.com/clear-street/reinforcer/internal/generator/passthrough"
	"github.com/clear-street/reinforcer/internal/generator/retryable"
	"github.com/dave/jennifer/jen"
	"github.com/rs/zerolog/log"
)

var fileHeader = "Code generated by reinforcer, DO NOT EDIT."

// FileConfig holds the code generation configuration for a specific type
type FileConfig struct {
	// srcTypeName is the source type that we want to generate code for
	srcTypeName string
	// outTypeName is the desired output type name
	outTypeName string
	// typeParams is the list of generic type parameters
	typeParams []jen.Code
	// typeArgs is the list of generic type arguments
	typeArgs []jen.Code
	// methods that should be in the generated type
	methods []*method.Method
}

// NewFileConfig creates a new instance of the FileConfig which holds code generation configuration
func NewFileConfig(srcTypeName, outTypeName string, typeParams []jen.Code, typeArgs []jen.Code, methods []*method.Method) *FileConfig {
	// cannot use cases.Title as it will lowercase MyService to Myservice
	if len(srcTypeName) > 0 {
		srcTypeName = strings.ToUpper(string(srcTypeName[0])) + srcTypeName[1:]
	}
	if len(outTypeName) > 0 {
		outTypeName = strings.ToUpper(string(outTypeName[0])) + outTypeName[1:]
	}
	return &FileConfig{
		srcTypeName: srcTypeName,
		outTypeName: outTypeName,
		typeParams:  typeParams,
		typeArgs:    typeArgs,
		methods:     methods,
	}
}

func (f *FileConfig) targetName() string {
	return "target" + f.srcTypeName
}

func (f *FileConfig) receiverName() string {
	return strings.ToLower(f.outTypeName[0:1])
}

// Config holds the code generation configuration for all of desired types
type Config struct {
	// OutPkg holds the name of the output package
	OutPkg string
	// Files holds the code generation configuration for every file being processed
	Files []*FileConfig
	// IgnoreNoReturnMethods determines whether methods that don't return anything should be wrapped in the middleware or not.
	IgnoreNoReturnMethods bool
}

// GeneratedFile contains the code generation output for a specific type
type GeneratedFile struct {
	// TypeName is the type's name that has been generated, note that this is the output version not the source
	TypeName string
	// Contents is the golang code that was generated
	Contents string
}

type statement interface {
	Statement() (*jen.Statement, error)
}

// Generated contains the code generation out for all the processed types
type Generated struct {
	// Common is the golang code that is shared across all generated types
	Common string
	// Files is the golang code that was generated for every type that was processed
	Files []*GeneratedFile
}

// Generate processes the given configuration and performs reinforcer's code generation
func Generate(cfg Config) (*Generated, error) {
	if len(cfg.Files) == 0 {
		return nil, fmt.Errorf("must provide at least one file for generation")
	}

	c, err := generateCommon(cfg.OutPkg)
	if err != nil {
		return nil, err
	}

	gen := &Generated{
		Common: c,
	}

	for _, fileConfig := range cfg.Files {
		methods := fileConfig.methods
		s, err := generateFile(cfg.OutPkg, cfg.IgnoreNoReturnMethods, fileConfig, methods)
		if err != nil {
			return nil, err
		}
		gen.Files = append(gen.Files, &GeneratedFile{
			TypeName: fileConfig.outTypeName,
			Contents: s,
		})
	}

	return gen, nil
}

// generateFile generates the proxy code for the given interface, the interface must have at least one method returning an
// error as those are the only ones wrapped in the middleware
func generateFile(outPkg string, ignoreNoReturnMethods bool, fileCfg *FileConfig, methods []*method.Method) (string, error) {
	f := jen.NewFile(outPkg)
	f.HeaderComment(fileHeader)

	// Compile-time constants
	var fields []jen.Code
	var constantAssign []jen.Code
	for _, m := range methods {
		fields = append(fields, jen.Id(m.Name).Id("string"))
		constantAssign = append(constantAssign, jen.Id(m.Name).Op(":").Lit(m.Name).Op(","))
	}

	constObjName := fmt.Sprintf("%sMethods", fileCfg.outTypeName)
	log.Debug().Msgf("Adding constants for type %s", fileCfg.outTypeName)
	f.Add(jen.Comment(fmt.Sprintf("%s are the methods in %s", constObjName, fileCfg.outTypeName)))
	f.Add(
		jen.Var().Id(constObjName).Op("=").Struct(
			fields...,
		).Block(
			constantAssign...,
		),
	)

	// Declare the target interface we are proxying
	var declMethods []jen.Code
	for _, meth := range methods {
		declMethods = append(declMethods, jen.Id(meth.Name).Params(meth.ParametersNameAndType...).Params(meth.ReturnTypes...))
	}
	f.Add(jen.Type().Id(fileCfg.targetName()).Types(fileCfg.typeParams...).Interface(
		declMethods...,
	))

	// Declare the proxy implementation
	f.Add(jen.Type().Id(fileCfg.outTypeName).Types(fileCfg.typeParams...).Struct(
		jen.Op("*").Id("base"),
		jen.Id("delegate").Id(fileCfg.targetName()).Types(fileCfg.typeArgs...),
	))

	// Declare the ctor
	f.Add(jen.Func().Id("New"+fileCfg.outTypeName).Types(fileCfg.typeParams...).Params(
		jen.Id("delegate").Id(fileCfg.targetName()).Types(fileCfg.typeArgs...),
		jen.Id("runnerFactory").Id("runnerFactory"),
		jen.Id("options").Op("...").Id("Option"),
	).Op("*").Id(fileCfg.outTypeName).Types(fileCfg.typeArgs...).Block(
		// if delegate == nil
		jen.If(jen.Id("delegate").Op("==").Nil().Block(
			// panic("...")
			jen.Panic(jen.Lit("provided nil delegate")),
		)),
		// if runnerFactory == nil
		jen.If(jen.Id("runnerFactory").Op("==").Nil().Block(
			// panic("...")
			jen.Panic(jen.Lit("provided nil runner factory")),
		)),
		// c:= &OutTypeName{...}
		jen.Id("c").Op(":=").Add(jen.Op("&").Id(fileCfg.outTypeName).Types(fileCfg.typeArgs...).Values(jen.Dict{
			// embed the base struct
			jen.Id("base"): jen.Op("&").Id("base").Values(jen.Dict{
				jen.Id("errorPredicate"): jen.Id("RetryAllErrors"),
				jen.Id("runnerFactory"):  jen.Id("runnerFactory"),
			}),
			jen.Id("delegate"): jen.Id("delegate"),
		})),
		// for _, o := range options {...}
		jen.For(jen.Id("_").Op(",").Id("o").Op(":=").Range().Id("options")).Block(
			jen.Id("o").Call(jen.Id("c").Dot("base")),
		),
		jen.Return(jen.Id("c")),
	))

	// Declare all of our proxy methods
	for _, mm := range methods {
		if mm.ReturnsError {
			r := retryable.NewRetryable(mm, fileCfg.outTypeName, fileCfg.typeArgs, fileCfg.receiverName())
			s, err := r.Statement()
			if err != nil {
				return "", err
			}
			f.Add(s)
		} else {
			var p statement
			if ignoreNoReturnMethods {
				p = passthrough.NewPassThrough(mm, fileCfg.outTypeName, fileCfg.typeArgs, fileCfg.receiverName())
			} else {
				p = noret.NewNoReturn(mm, fileCfg.outTypeName, fileCfg.typeArgs, fileCfg.receiverName())
			}
			s, err := p.Statement()
			if err != nil {
				return "", err
			}
			f.Add(s)
		}
	}
	return renderToString(f)
}

func generateCommon(outPkg string) (string, error) {
	f := jen.NewFile(outPkg)
	f.HeaderComment(fileHeader)

	// Declare base impl that will be used to hold the common fields
	f.Add(jen.Type().Id("base").Struct(
		jen.Id("errorPredicate").Add(jen.Func().Params(jen.Id("string"), jen.Id("error")).Params(jen.Bool())),
		jen.Id("runnerFactory").Id("runnerFactory"),
	))

	// Declares the runner's factory
	f.Add(jen.Type().Id("runnerFactory").Interface(
		jen.Id("GetRunner").Params(jen.Id("name").Id("string")).Qual("github.com/slok/goresilience", "Runner"),
	))

	// Declare the RetryAllErrors predicate that enables the middleware on all errors received from proxy call
	f.Add(jen.Var().Id("RetryAllErrors").Op("=").Func().Params(jen.Id("_").Id("string"), jen.Id("_").Id("error")).Params(jen.Id("bool")).Block(
		jen.Return(jen.Lit(true)),
	))

	// Declare the Option type that allows to configure the service
	f.Add(jen.Type().Id("Option").Func().Params(jen.Op("*").Id("base")))

	// Declare the WithRetryableErrorPredicate Option which configures the predicate to determine which errors should be retried
	f.Add(jen.Func().Id("WithRetryableErrorPredicate").Params(jen.Id("fn").Id("func").Params(jen.Id("string"), jen.Id("error")).Params(jen.Bool())).Params(jen.Id("Option")).Block(
		jen.Return(jen.Func().Params(jen.Id("o").Op("*").Id("base")).Block(
			jen.Id("o").Dot("errorPredicate").Op("=").Id("fn"),
		)),
	))

	// Declare our runner helper
	f.Add(jen.Func().Params(jen.Id("b").Op("*").Id("base")).Id("run").Params(
		jen.Id("ctx").Qual("context", "Context"),
		jen.Id("name").Id("string"),
		jen.Id("fn").Func().Params(jen.Id("ctx").Qual("context", "Context")).Id("error"),
	).Id("error").Block(
		jen.Return(jen.Id("b").Dot("runnerFactory").Dot("GetRunner").Call(jen.Id("name")).Dot("Run").Call(jen.Id("ctx"), jen.Id("fn"))),
	))
	return renderToString(f)
}

func renderToString(f *jen.File) (string, error) {
	b := &bytes.Buffer{}
	if err := f.Render(b); err != nil {
		return "", err
	}
	return b.String(), nil
}
