package workspaces

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/andrew-malikov/workspace/dotnet"
	flr "github.com/andrew-malikov/workspace/failure"
	projects "github.com/andrew-malikov/workspace/projects"

	"github.com/BurntSushi/toml"
)

type Workspace struct {
	Projects map[string]projects.Project  `toml:"projects"`
	Git      GitConfig                    `toml:"git"`
	Test     projects.TestDiscoveryConfig `toml:"test"`
}

type GitConfig struct {
	Clear GitClearConfig `toml:"clear"`
}

type GitClearConfig struct {
	Ownership GitClearOwnershipConfig `toml:"ownership"`
}

type GitClearOwnershipConfig struct {
	LookbackCommits int                           `toml:"lookback_commits"`
	Include         GitClearOwnershipFilterConfig `toml:"include"`
	Exclude         GitClearOwnershipFilterConfig `toml:"exclude"`
}

type GitClearOwnershipFilterConfig struct {
	AuthorEmails    []string `toml:"author_emails"`
	AuthorNames     []string `toml:"author_names"`
	MessagePatterns []string `toml:"message_patterns"`
}

func emptyWorkspace() Workspace {
	return Workspace{
		Projects: map[string]projects.Project{},
		Git:      GitConfig{},
	}
}

func (workspace *Workspace) Normalize() {
	if workspace.Projects == nil {
		workspace.Projects = map[string]projects.Project{}
	}

	if workspace.Git.Clear.Ownership.LookbackCommits <= 0 {
		workspace.Git.Clear.Ownership.LookbackCommits = 3
	}
}

func (workspace Workspace) Validate() error {
	if err := validateMessagePatterns(workspace.Git.Clear.Ownership.Include.MessagePatterns, "git.clear.ownership.include.message_patterns"); err != nil {
		return err
	}

	if err := validateMessagePatterns(workspace.Git.Clear.Ownership.Exclude.MessagePatterns, "git.clear.ownership.exclude.message_patterns"); err != nil {
		return err
	}

	for _, kind := range projects.AllTestKinds {
		target := workspace.Test.Target(kind)
		if err := validatePatterns(target.ProjectPatterns, fmt.Sprintf("test.%s.project_patterns", kind)); err != nil {
			return err
		}
	}

	return nil
}

func validateMessagePatterns(patterns []string, field string) error {
	return validatePatterns(patterns, field)
}

func validatePatterns(patterns []string, field string) error {
	for _, pattern := range patterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid %s regex %q: %w", field, pattern, err)
		}
	}

	return nil
}

func (workspace Workspace) ResolveProjectByDir(dir string) *projects.Project {
	cleanDir := filepath.Clean(dir)
	var matchedProject *projects.Project
	for _, project := range workspace.Projects {
		cleanProjectDir := filepath.Clean(project.Dir)
		if cleanDir == cleanProjectDir || strings.HasPrefix(cleanDir, cleanProjectDir+string(os.PathSeparator)) {
			current := project
			if matchedProject == nil || len(current.Dir) > len(matchedProject.Dir) {
				matchedProject = &current
			}
		}
	}

	return matchedProject
}

type AddProjectResult struct {
	DoesComposeExist    bool
	DoMigrationsExist   bool
	HasUnitTests        bool
	HasIntegrationTests bool
	HasComponentTests   bool
}

func (workspace *Workspace) AddProject(project projects.Project) (*AddProjectResult, error) {
	for _, existingProject := range workspace.Projects {
		if existingProject.Alias == project.Alias {
			return nil, fmt.Errorf("a project with such alias %s already exist at %s", project.Alias, existingProject.Dir)
		}
		if strings.HasPrefix(existingProject.Dir, project.Dir) {
			return nil, fmt.Errorf("a project is already tracked under %s with alias %s", existingProject.Dir, existingProject.Alias)
		}
	}

	result := AddProjectResult{
		DoesComposeExist:  project.DoesComposeExist(),
		DoMigrationsExist: project.DoMigrationsExist(),
	}

	if !result.DoesComposeExist {
		project.ResetCompose()
	}

	if !result.DoMigrationsExist {
		project.ResetMigrations()
	}

	if err := workspace.ScaffoldProjectTests(&project); err != nil {
		return nil, err
	}

	result.HasUnitTests = project.Test.Unit.IsConfigured()
	result.HasIntegrationTests = project.Test.Integration.IsConfigured()
	result.HasComponentTests = project.Test.Component.IsConfigured()

	if len(workspace.Projects) == 0 {
		workspace.Projects = map[string]projects.Project{
			project.Alias: project,
		}
	} else {
		workspace.Projects[project.Alias] = project
	}

	return &result, nil
}

func (workspace Workspace) ScaffoldProjectTests(project *projects.Project) error {
	discoveredTests, err := dotnet.DiscoverTests(project.Dir, workspace.Test)
	if err != nil {
		return err
	}

	project.Test = discoveredTests
	return nil
}

type NoProjectFound struct {
	Alsdir string
}

func (workspace *Workspace) RemoveProject(alsdir string) (*projects.Project, *flr.Failure) {
	var foundProject *projects.Project
	for _, project := range workspace.Projects {
		if project.Alias == alsdir || strings.HasPrefix(alsdir, project.Dir) {
			foundProject = &project
			break
		}
	}

	if foundProject == nil {
		return nil, flr.OfCtx(
			flr.Context{
				// todo: definitely MUST be a constant
				Type: "NO_PROJECT_FOUND",
				Details: &NoProjectFound{
					Alsdir: alsdir,
				},
			},
		)
	}

	delete(workspace.Projects, foundProject.Alias)

	return foundProject, nil
}

const DEFAULT_UNIX_CONFIG_DIR = ".config"

var CONFIG_PERMISSION = os.FileMode(0777)

func getConfigPath() (*string, error) {
	osType := runtime.GOOS

	var dir string
	switch osType {
	case "darwin":
		fallthrough
	case "linux":
		dir = os.Getenv("XDG_CONFIG_HOME")
		if dir == "" {
			dir = DEFAULT_UNIX_CONFIG_DIR
		}
		break

	default:
		return nil, fmt.Errorf("The OS isn't supported yet: %s\n", osType)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, dir, "ws/config.toml")
	return &path, nil
}

func LoadWorkspace() (*Workspace, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(*configPath)
	if err != nil {
		workspace := emptyWorkspace()
		workspace.Normalize()
		err = SaveWorkspace(workspace)
		if err != nil {
			return nil, err
		}
		return &workspace, nil
	}

	var workspace *Workspace
	err = toml.Unmarshal(content, &workspace)
	if err != nil {
		return nil, err
	}

	workspace.Normalize()
	if err := workspace.Validate(); err != nil {
		return nil, err
	}

	return workspace, nil
}

func SaveWorkspace(workspace Workspace) error {
	workspace.Normalize()
	if err := workspace.Validate(); err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	content, err := toml.Marshal(workspace)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(*configPath), CONFIG_PERMISSION)
	if err != nil {
		return err
	}
	return os.WriteFile(*configPath, content, CONFIG_PERMISSION)
}
