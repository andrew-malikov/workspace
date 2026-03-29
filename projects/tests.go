package projects

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TestKind string

const (
	UnitTestKind        TestKind = "unit"
	IntegrationTestKind TestKind = "integration"
	ComponentTestKind   TestKind = "component"
)

var AllTestKinds = []TestKind{
	UnitTestKind,
	IntegrationTestKind,
	ComponentTestKind,
}

type TestTarget struct {
	Project string `toml:"project"`
	Filter  string `toml:"filter"`
}

func (target TestTarget) IsConfigured() bool {
	return strings.TrimSpace(target.Project) != ""
}

type TestConfig struct {
	Unit        TestTarget `toml:"unit"`
	Integration TestTarget `toml:"integration"`
	Component   TestTarget `toml:"component"`
}

func (config TestConfig) Target(kind TestKind) TestTarget {
	switch kind {
	case UnitTestKind:
		return config.Unit
	case IntegrationTestKind:
		return config.Integration
	case ComponentTestKind:
		return config.Component
	default:
		return TestTarget{}
	}
}

func (config *TestConfig) SetTarget(kind TestKind, target TestTarget) {
	switch kind {
	case UnitTestKind:
		config.Unit = target
	case IntegrationTestKind:
		config.Integration = target
	case ComponentTestKind:
		config.Component = target
	}
}

func (config TestConfig) ConfiguredKinds() []TestKind {
	kinds := make([]TestKind, 0, len(AllTestKinds))
	for _, kind := range AllTestKinds {
		if config.Target(kind).IsConfigured() {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

type TestDiscoveryTarget struct {
	ProjectPatterns []string `toml:"project_patterns"`
	Filter          string   `toml:"filter"`
}

type TestDiscoveryConfig struct {
	Unit        TestDiscoveryTarget `toml:"unit"`
	Integration TestDiscoveryTarget `toml:"integration"`
	Component   TestDiscoveryTarget `toml:"component"`
}

func (config TestDiscoveryConfig) Target(kind TestKind) TestDiscoveryTarget {
	switch kind {
	case UnitTestKind:
		return config.Unit
	case IntegrationTestKind:
		return config.Integration
	case ComponentTestKind:
		return config.Component
	default:
		return TestDiscoveryTarget{}
	}
}

func (project *Project) ApplyDiscoveredTests(config TestDiscoveryConfig) error {
	matches, err := project.DiscoverTests(config)
	if err != nil {
		return err
	}

	project.Test = TestConfig{}
	for _, kind := range AllTestKinds {
		match, ok := matches[kind]
		if !ok {
			continue
		}

		project.Test.SetTarget(kind, TestTarget{
			Project: match,
			Filter:  config.Target(kind).Filter,
		})
	}

	return nil
}

func (project Project) DiscoverTests(config TestDiscoveryConfig) (map[TestKind]string, error) {
	projects, err := project.listCSharpProjects()
	if err != nil {
		return nil, err
	}

	result := map[TestKind]string{}
	for _, kind := range AllTestKinds {
		target := config.Target(kind)
		patterns := make([]*regexp.Regexp, 0, len(target.ProjectPatterns))
		for _, pattern := range target.ProjectPatterns {
			if strings.TrimSpace(pattern) == "" {
				continue
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid %s test project regex %q: %w", kind, pattern, err)
			}
			patterns = append(patterns, re)
		}

		if len(patterns) == 0 {
			continue
		}

		matches := make([]string, 0, 1)
		for _, projectPath := range projects {
			for _, pattern := range patterns {
				if pattern.MatchString(projectPath) {
					matches = append(matches, projectPath)
					break
				}
			}
		}

		if len(matches) > 1 {
			return nil, fmt.Errorf("found multiple %s test projects in %s: %s", kind, project.Dir, strings.Join(matches, ", "))
		}

		if len(matches) == 1 {
			result[kind] = matches[0]
		}
	}

	return result, nil
}

func (project Project) listCSharpProjects() ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(project.Dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".csproj" {
			return nil
		}

		rel, err := filepath.Rel(project.Dir, path)
		if err != nil {
			return err
		}

		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}
