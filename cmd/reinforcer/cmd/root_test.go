package cmd_test

import (
	"bytes"
	"testing"

	"github.com/clear-street/reinforcer/cmd/reinforcer/cmd"
	"github.com/clear-street/reinforcer/cmd/reinforcer/cmd/mocks"
	"github.com/clear-street/reinforcer/internal/generator"
	"github.com/clear-street/reinforcer/internal/generator/executor"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	gen := &generator.Generated{}

	t.Run("Provide Targets", func(t *testing.T) {
		exec := &mocks.Executor{}
		exec.On("Execute", &executor.Parameters{
			Sources:               []string{"/path/to/target.go"},
			SourcePackages:        []string{},
			Targets:               []string{"Client", "SomeOtherClient"},
			TargetsAll:            false,
			OutPkg:                "reinforced",
			IgnoreNoReturnMethods: false,
		}).Return(gen, nil)
		writ := &mocks.Writer{}
		writ.On("Write", "./reinforced", gen).Return(nil)

		b := bytes.NewBufferString("")
		c := cmd.NewRootCmd(exec, writ)
		c.SetOut(b)
		c.SetArgs([]string{"--src=/path/to/target.go", "--target=Client", "--target=SomeOtherClient", "--outputdir=./reinforced"})
		require.NoError(t, c.Execute())
	})

	t.Run("Source packages", func(t *testing.T) {
		exec := &mocks.Executor{}
		exec.On("Execute", &executor.Parameters{
			Sources:               []string{},
			SourcePackages:        []string{"github.com/clear-street/somelib"},
			Targets:               []string{"Client", "SomeOtherClient"},
			TargetsAll:            false,
			OutPkg:                "reinforced",
			IgnoreNoReturnMethods: false,
		}).Return(gen, nil)
		writ := &mocks.Writer{}
		writ.On("Write", "./reinforced", gen).Return(nil)

		b := bytes.NewBufferString("")
		c := cmd.NewRootCmd(exec, writ)
		c.SetOut(b)
		c.SetArgs([]string{"--srcpkg=github.com/clear-street/somelib", "--target=Client", "--target=SomeOtherClient", "--outputdir=./reinforced"})
		require.NoError(t, c.Execute())
	})

	t.Run("Target All", func(t *testing.T) {
		exec := &mocks.Executor{}
		exec.On("Execute", &executor.Parameters{
			Sources:               []string{"/path/to/target.go"},
			SourcePackages:        []string{},
			Targets:               []string{},
			TargetsAll:            true,
			OutPkg:                "reinforced",
			IgnoreNoReturnMethods: false,
		}).Return(gen, nil)
		writ := &mocks.Writer{}
		writ.On("Write", "./reinforced", gen).Return(nil)

		b := bytes.NewBufferString("")
		c := cmd.NewRootCmd(exec, writ)
		c.SetOut(b)
		c.SetArgs([]string{"--src=/path/to/target.go", "--targetall", "--outputdir=./reinforced"})
		require.NoError(t, c.Execute())
	})

	t.Run("Ignore No Return Methods", func(t *testing.T) {
		exec := &mocks.Executor{}
		exec.On("Execute", &executor.Parameters{
			Sources:               []string{"/path/to/target.go"},
			SourcePackages:        []string{},
			Targets:               []string{"Client", "SomeOtherClient"},
			TargetsAll:            false,
			OutPkg:                "reinforced",
			IgnoreNoReturnMethods: true,
		}).Return(gen, nil)
		writ := &mocks.Writer{}
		writ.On("Write", "./reinforced", gen).Return(nil)

		b := bytes.NewBufferString("")
		c := cmd.NewRootCmd(exec, writ)
		c.SetOut(b)
		c.SetArgs([]string{"--src=/path/to/target.go", "--target=Client", "--target=SomeOtherClient", "--outputdir=./reinforced", "--ignorenoret"})
		require.NoError(t, c.Execute())
	})

	t.Run("No targets found", func(t *testing.T) {
		exec := &mocks.Executor{}
		exec.On("Execute", &executor.Parameters{
			Sources:               []string{"/path/to/target.go"},
			SourcePackages:        []string{},
			Targets:               []string{},
			TargetsAll:            true,
			OutPkg:                "reinforced",
			IgnoreNoReturnMethods: false,
		}).Return(nil, executor.ErrNoTargetableTypesFound)
		writ := &mocks.Writer{}
		writ.On("Write", "./reinforced", gen).Return(nil)

		b := bytes.NewBufferString("")
		c := cmd.NewRootCmd(exec, writ)
		c.SetOut(b)
		c.SetArgs([]string{"--src=/path/to/target.go", "--targetall", "--outputdir=./reinforced"})
		require.EqualError(t, c.Execute(), "failed to generate code; error=no targetable types were discovered")
	})
}
