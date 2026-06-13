package dotnet

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/andrew-malikov/workspace/projects"
)

func DiscoverTests(root string, config projects.TestDiscoveryConfig) (projects.TestConfig, error) {
	projectPaths, err := listProjects(root)
	if err != nil {
		return projects.TestConfig{}, err
	}

	result := projects.TestConfig{}
	for _, kind := range projects.AllTestKinds {
		target := config.Target(kind)
		patterns, err := compilePatterns(kind, target.ProjectPatterns)
		if err != nil {
			return projects.TestConfig{}, err
		}

		if len(patterns) == 0 {
			continue
		}

		matches := matchProjects(projectPaths, patterns)
		if len(matches) > 1 {
			return projects.TestConfig{}, fmt.Errorf("found multiple %s test projects in %s: %s", kind, root, strings.Join(matches, ", "))
		}

		if len(matches) == 1 {
			result.SetTarget(kind, projects.TestTarget{
				Project: matches[0],
				Filter:  target.Filter,
			})
		}
	}

	return result, nil
}

func listProjects(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".csproj" && ext != ".fsproj" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
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

func compilePatterns(kind projects.TestKind, rawPatterns []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, pattern := range rawPatterns {
		if strings.TrimSpace(pattern) == "" {
			continue
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid %s test project regex %q: %w", kind, pattern, err)
		}

		patterns = append(patterns, re)
	}

	return patterns, nil
}

func matchProjects(projectPaths []string, patterns []*regexp.Regexp) []string {
	matches := make([]string, 0, 1)
	for _, projectPath := range projectPaths {
		for _, pattern := range patterns {
			if pattern.MatchString(projectPath) {
				matches = append(matches, projectPath)
				break
			}
		}
	}

	return matches
}
