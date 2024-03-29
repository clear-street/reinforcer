package generator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/clear-street/reinforcer/internal/generator"
	"github.com/clear-street/reinforcer/internal/loader"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"
)

type input struct {
	interfaceName string
	code          string
}

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name                  string
		ignoreNoReturnMethods bool
		inputs                map[string]input
		outCode               *generator.Generated
		wantErr               bool
	}{
		{
			name:                  "Using aliased import",
			ignoreNoReturnMethods: false,
			inputs: map[string]input{
				"my_service.go": {
					interfaceName: "Service",
					code: `package fake

import goctx "context"

type Service interface {
	A(ctx goctx.Context) error
	B(ctx goctx.Context, fn func(myArg string) (myBool bool)) (func() bool, error)
}
`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import "context"

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	A string
	B string
}{
	A: "A",
	B: "B",
}

type targetService interface {
	A(ctx context.Context) error
	B(ctx context.Context, arg1 func(string) bool) (func() bool, error)
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) A(ctx context.Context) error {
	var nonRetryableErr error
	err := g.run(ctx, GeneratedServiceMethods.A, func(ctx context.Context) error {
		var err error
		err = g.delegate.A(ctx)
		if g.errorPredicate(GeneratedServiceMethods.A, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
func (g *GeneratedService) B(ctx context.Context, arg1 func(string) bool) (func() bool, error) {
	var nonRetryableErr error
	var r0 func() bool
	err := g.run(ctx, GeneratedServiceMethods.B, func(ctx context.Context) error {
		var err error
		r0, err = g.delegate.B(ctx, arg1)
		if g.errorPredicate(GeneratedServiceMethods.B, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}
`,
					},
				},
			},
		},
		{
			name:                  "Complex",
			ignoreNoReturnMethods: false,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "Service",
					code: `package fake

import "context"

type User struct {
	Name string
}

type Service interface {
	A()
	B(ctx context.Context)
	C(ctx context.Context, param1 int, param2 *int32, param3 *User)
	GetUserID(ctx context.Context, userID string) (string, error)
	GetUserID2(ctx context.Context, userID *string) (*User, error)
	HasVariadic(ctx context.Context, fields ...string) error
}`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	unresilient "github.com/clear-street/fake/unresilient"
)

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	A           string
	B           string
	C           string
	GetUserID   string
	GetUserID2  string
	HasVariadic string
}{
	A:           "A",
	B:           "B",
	C:           "C",
	GetUserID:   "GetUserID",
	GetUserID2:  "GetUserID2",
	HasVariadic: "HasVariadic",
}

type targetService interface {
	A()
	B(ctx context.Context)
	C(ctx context.Context, arg1 int, arg2 *int32, arg3 *unresilient.User)
	GetUserID(ctx context.Context, arg1 string) (string, error)
	GetUserID2(ctx context.Context, arg1 *string) (*unresilient.User, error)
	HasVariadic(ctx context.Context, arg1 ...string) error
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) A() {
	err := g.run(context.Background(), GeneratedServiceMethods.A, func(_ context.Context) error {
		g.delegate.A()
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func (g *GeneratedService) B(ctx context.Context) {
	err := g.run(ctx, GeneratedServiceMethods.B, func(ctx context.Context) error {
		g.delegate.B(ctx)
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func (g *GeneratedService) C(ctx context.Context, arg1 int, arg2 *int32, arg3 *unresilient.User) {
	err := g.run(ctx, GeneratedServiceMethods.C, func(ctx context.Context) error {
		g.delegate.C(ctx, arg1, arg2, arg3)
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func (g *GeneratedService) GetUserID(ctx context.Context, arg1 string) (string, error) {
	var nonRetryableErr error
	var r0 string
	err := g.run(ctx, GeneratedServiceMethods.GetUserID, func(ctx context.Context) error {
		var err error
		r0, err = g.delegate.GetUserID(ctx, arg1)
		if g.errorPredicate(GeneratedServiceMethods.GetUserID, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}
func (g *GeneratedService) GetUserID2(ctx context.Context, arg1 *string) (*unresilient.User, error) {
	var nonRetryableErr error
	var r0 *unresilient.User
	err := g.run(ctx, GeneratedServiceMethods.GetUserID2, func(ctx context.Context) error {
		var err error
		r0, err = g.delegate.GetUserID2(ctx, arg1)
		if g.errorPredicate(GeneratedServiceMethods.GetUserID2, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}
func (g *GeneratedService) HasVariadic(ctx context.Context, arg1 ...string) error {
	var nonRetryableErr error
	err := g.run(ctx, GeneratedServiceMethods.HasVariadic, func(ctx context.Context) error {
		var err error
		err = g.delegate.HasVariadic(ctx, arg1...)
		if g.errorPredicate(GeneratedServiceMethods.HasVariadic, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
`,
					},
				},
			},
		},
		{
			name:                  "Ignore No Return Methods",
			ignoreNoReturnMethods: true,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "Service",
					code: `package fake

import "context"

type User struct {
	Name string
}

type Service interface {
	A()
	B(ctx context.Context, userID string) (string, error)
}`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import "context"

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	A string
	B string
}{
	A: "A",
	B: "B",
}

type targetService interface {
	A()
	B(ctx context.Context, arg1 string) (string, error)
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) A() {
	g.delegate.A()
}
func (g *GeneratedService) B(ctx context.Context, arg1 string) (string, error) {
	var nonRetryableErr error
	var r0 string
	err := g.run(ctx, GeneratedServiceMethods.B, func(ctx context.Context) error {
		var err error
		r0, err = g.delegate.B(ctx, arg1)
		if g.errorPredicate(GeneratedServiceMethods.B, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}
`,
					},
				},
			},
		},
		{
			name:                  "Local type resolved properly in generated code",
			ignoreNoReturnMethods: true,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "Service",
					code: `package fake

type T struct {
	Name string
}

type Service interface {
	SaveUser(user *T) error
}`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	unresilient "github.com/clear-street/fake/unresilient"
)

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	SaveUser string
}{
	SaveUser: "SaveUser",
}

type targetService interface {
	SaveUser(arg0 *unresilient.T) error
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) SaveUser(arg0 *unresilient.T) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.SaveUser, func(_ context.Context) error {
		var err error
		err = g.delegate.SaveUser(arg0)
		if g.errorPredicate(GeneratedServiceMethods.SaveUser, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
`,
					},
				},
			},
		},
		{
			name:                  "Channels",
			ignoreNoReturnMethods: true,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "Service",
					code: `package fake

type Service interface {
	ReceiveDir(myChan <- chan error) error
	SendDir(myChan chan <- error) error
	SendReceiveDir(myChan chan error) error
}`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import "context"

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	ReceiveDir     string
	SendDir        string
	SendReceiveDir string
}{
	ReceiveDir:     "ReceiveDir",
	SendDir:        "SendDir",
	SendReceiveDir: "SendReceiveDir",
}

type targetService interface {
	ReceiveDir(arg0 <-chan error) error
	SendDir(arg0 chan<- error) error
	SendReceiveDir(arg0 chan error) error
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) ReceiveDir(arg0 <-chan error) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.ReceiveDir, func(_ context.Context) error {
		var err error
		err = g.delegate.ReceiveDir(arg0)
		if g.errorPredicate(GeneratedServiceMethods.ReceiveDir, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
func (g *GeneratedService) SendDir(arg0 chan<- error) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.SendDir, func(_ context.Context) error {
		var err error
		err = g.delegate.SendDir(arg0)
		if g.errorPredicate(GeneratedServiceMethods.SendDir, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
func (g *GeneratedService) SendReceiveDir(arg0 chan error) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.SendReceiveDir, func(_ context.Context) error {
		var err error
		err = g.delegate.SendReceiveDir(arg0)
		if g.errorPredicate(GeneratedServiceMethods.SendReceiveDir, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
`,
					},
				},
			},
		},
		{
			name:                  "Unexported",
			ignoreNoReturnMethods: true,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "service",
					code: `package fake

type service interface {
	SayHello(name string) error
}
`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import "context"

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	SayHello string
}{
	SayHello: "SayHello",
}

type targetService interface {
	SayHello(arg0 string) error
}
type GeneratedService struct {
	*base
	delegate targetService
}

func NewGeneratedService(delegate targetService, runnerFactory runnerFactory, options ...Option) *GeneratedService {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService) SayHello(arg0 string) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.SayHello, func(_ context.Context) error {
		var err error
		err = g.delegate.SayHello(arg0)
		if g.errorPredicate(GeneratedServiceMethods.SayHello, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
`,
					},
				},
			},
		},
		{
			name:                  "Generic Type Parameters",
			ignoreNoReturnMethods: true,
			inputs: map[string]input{
				"users_service.go": {
					interfaceName: "Service",
					code: `package fake

type Service[T any] interface {
	SayHello(name T) error
	DoNothing()
}
`,
				},
			},
			outCode: &generator.Generated{
				Common: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import (
	"context"
	goresilience "github.com/slok/goresilience"
)

type base struct {
	errorPredicate func(string, error) bool
	runnerFactory  runnerFactory
}
type runnerFactory interface {
	GetRunner(name string) goresilience.Runner
}

var RetryAllErrors = func(_ string, _ error) bool {
	return true
}

type Option func(*base)

func WithRetryableErrorPredicate(fn func(string, error) bool) Option {
	return func(o *base) {
		o.errorPredicate = fn
	}
}
func (b *base) run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	return b.runnerFactory.GetRunner(name).Run(ctx, fn)
}
`,
				Files: []*generator.GeneratedFile{
					{
						TypeName: "GeneratedService",
						Contents: `// Code generated by reinforcer, DO NOT EDIT.

package resilient

import "context"

// GeneratedServiceMethods are the methods in GeneratedService
var GeneratedServiceMethods = struct {
	DoNothing string
	SayHello  string
}{
	DoNothing: "DoNothing",
	SayHello:  "SayHello",
}

type targetService[T any] interface {
	DoNothing()
	SayHello(arg0 T) error
}
type GeneratedService[T any] struct {
	*base
	delegate targetService[T]
}

func NewGeneratedService[T any](delegate targetService[T], runnerFactory runnerFactory, options ...Option) *GeneratedService[T] {
	if delegate == nil {
		panic("provided nil delegate")
	}
	if runnerFactory == nil {
		panic("provided nil runner factory")
	}
	c := &GeneratedService[T]{
		base: &base{
			errorPredicate: RetryAllErrors,
			runnerFactory:  runnerFactory,
		},
		delegate: delegate,
	}
	for _, o := range options {
		o(c.base)
	}
	return c
}
func (g *GeneratedService[T]) DoNothing() {
	g.delegate.DoNothing()
}
func (g *GeneratedService[T]) SayHello(arg0 T) error {
	var nonRetryableErr error
	err := g.run(context.Background(), GeneratedServiceMethods.SayHello, func(_ context.Context) error {
		var err error
		err = g.delegate.SayHello(arg0)
		if g.errorPredicate(GeneratedServiceMethods.SayHello, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}
`,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ifaces := loadInterface(t, tt.inputs)
			got, err := generator.Generate(generator.Config{
				OutPkg:                "resilient",
				Files:                 ifaces,
				IgnoreNoReturnMethods: tt.ignoreNoReturnMethods,
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				require.Equal(t, tt.outCode.Common, got.Common)

				for idx, genFile := range got.Files {
					require.Equal(t, tt.outCode.Files[idx].TypeName, genFile.TypeName)
					require.Equal(t, tt.outCode.Files[idx].Contents, genFile.Contents, "Contents don't match. Got:\n%s", genFile.Contents)
				}
				require.Equal(t, tt.outCode, got)
			}
		})
	}
}

func loadInterface(t *testing.T, filesCode map[string]input) []*generator.FileConfig {
	pkg := "github.com/clear-street/fake/unresilient"
	m := map[string]interface{}{}
	for fileName, in := range filesCode {
		m[fileName] = in.code
	}

	mods := []packagestest.Module{
		{
			Name:  pkg,
			Files: m,
		},
	}
	exported := packagestest.Export(t, packagestest.GOPATH, mods)
	defer exported.Cleanup()

	l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
		exported.Config.Mode = cfg.Mode
		return packages.Load(exported.Config, patterns...)
	})

	var loadedTypes []*generator.FileConfig
	for _, in := range filesCode {
		svc, err := l.LoadOne(pkg, in.interfaceName, loader.PackageLoadMode)
		require.NoError(t, err)
		// cannot use cases.Title as it will lowercase MyService to Myservice
		if len(in.interfaceName) > 0 {
			in.interfaceName = strings.ToUpper(string(in.interfaceName[0])) + in.interfaceName[1:]
		}
		loadedTypes = append(loadedTypes, generator.NewFileConfig(in.interfaceName,
			fmt.Sprintf("Generated%s", in.interfaceName),
			svc.TypeParams,
			svc.TypeArgs,
			svc.Methods,
		))
	}
	return loadedTypes
}
