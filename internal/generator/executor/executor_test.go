package executor_test

import (
	"github.com/csueiras/reinforcer/internal/generator/executor"
	"github.com/csueiras/reinforcer/internal/generator/executor/mocks"
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/stretchr/testify/require"
	"go/token"
	"go/types"
	"testing"
)

func TestExecutor_Execute(t *testing.T) {
	l := &mocks.Loader{}
	l.On("LoadMatched", "./testpkg.go", []string{"MyService"}, loader.FileLoadMode).Return(
		map[string]*loader.Result{
			"LockService": &loader.Result{
				Name:          "LockService",
				InterfaceType: createTestInterfaceType(),
				Package:       types.NewPackage("github.com/csueiras/testpkg", "testpkg"),
			},
		}, nil,
	)

	exec := executor.New(l)
	got, err := exec.Execute(&executor.Parameters{
		Sources:               []string{"./testpkg.go"},
		Targets:               []string{"MyService"},
		OutPkg:                "testpkg",
		IgnoreNoReturnMethods: false,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 1, len(got.Files))
	require.Equal(t, "LockService", got.Files[0].TypeName)
}

func createTestInterfaceType() *types.Interface {
	nullary := types.NewSignature(nil, nil, nil, false) // func()
	methods := []*types.Func{
		types.NewFunc(token.NoPos, nil, "Lock", nullary),
		types.NewFunc(token.NoPos, nil, "Unlock", nullary),
	}
	return types.NewInterfaceType(methods, nil).Complete()
}
