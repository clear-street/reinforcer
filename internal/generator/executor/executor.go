//go:generate mockery --all

package executor

import (
	"fmt"

	"github.com/clear-street/reinforcer/internal/generator"
	"github.com/clear-street/reinforcer/internal/loader"
	"github.com/pkg/errors"
)

// ErrNoTargetableTypesFound indicates that no types that could be targeted for code generation were discovered
var ErrNoTargetableTypesFound = fmt.Errorf("no targetable types were discovered")

// Loader describes the loader component
type Loader interface {
	LoadAll(path string, mode loader.LoadMode) (map[string]*loader.Result, error)
	LoadMatched(path string, expressions []string, mode loader.LoadMode) (map[string]*loader.Result, error)
}

// Parameters are the input parameters for the executor
type Parameters struct {
	// Sources are the paths to the packages that are eligible for targeting
	Sources []string
	// SourcePackages are the packages that are eligible for targeting (e.g. github.com/clear-street/somelib)
	SourcePackages []string
	// Targets contains the target types to search for, these are expressions that may contain RegEx
	Targets []string
	// TargetsAll enables targeting of every exported interface type
	TargetsAll bool
	// OutPkg the package name for the output code
	OutPkg string
	// IgnoreNoReturnMethods disables proxying of methods that don't return anything
	IgnoreNoReturnMethods bool
}

// Executor is a utility service to orchestrate code generation
type Executor struct {
	loader Loader
}

// New creates an instance of the executor with the given type loader
func New(l Loader) *Executor {
	return &Executor{loader: l}
}

// Execute orchestrates code generation sourced from multiple files/targets
func (e *Executor) Execute(settings *Parameters) (*generator.Generated, error) {
	discoveredTypes := make(map[string]struct{})

	var cfg []*generator.FileConfig
	var err error

	for _, sourcePkg := range settings.SourcePackages {
		var match map[string]*loader.Result
		if settings.TargetsAll {
			match, err = e.loader.LoadAll(sourcePkg, loader.PackageLoadMode)
		} else {
			match, err = e.loader.LoadMatched(sourcePkg, settings.Targets, loader.PackageLoadMode)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load from pkg=%s", sourcePkg)
		}

		configs, err := createFileConfigs(discoveredTypes, match)
		if err != nil {
			return nil, err
		}

		cfg = append(cfg, configs...)
	}

	for _, source := range settings.Sources {
		var match map[string]*loader.Result
		if settings.TargetsAll {
			match, err = e.loader.LoadAll(source, loader.FileLoadMode)
		} else {
			match, err = e.loader.LoadMatched(source, settings.Targets, loader.FileLoadMode)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load from file=%s", source)
		}
		configs, err := createFileConfigs(discoveredTypes, match)
		if err != nil {
			return nil, err
		}
		cfg = append(cfg, configs...)
	}

	if len(cfg) == 0 {
		return nil, ErrNoTargetableTypesFound
	}

	code, err := generator.Generate(generator.Config{
		OutPkg:                settings.OutPkg,
		IgnoreNoReturnMethods: settings.IgnoreNoReturnMethods,
		Files:                 cfg,
	})
	if err != nil {
		return nil, err
	}
	return code, nil
}

func createFileConfigs(discoveredSet map[string]struct{}, match map[string]*loader.Result) ([]*generator.FileConfig, error) {
	var cfg []*generator.FileConfig
	for typName, res := range match {
		// Check types aren't repeated before adding them to the generator's config
		if _, ok := discoveredSet[typName]; ok {
			return nil, errors.Errorf("multiple types with same name discovered with name %s", typName)
		}
		discoveredSet[typName] = struct{}{}
		cfg = append(cfg, generator.NewFileConfig(typName, typName, res.TypeParams, res.TypeArgs, res.Methods))
	}
	return cfg, nil
}
