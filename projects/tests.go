package projects

import "strings"

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
