package cfg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flr "github.com/andrew-malikov/workspace/failure"
	projects "github.com/andrew-malikov/workspace/projects"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Projects map[string]projects.Project
}

func emptyConfig() Config {
	return Config{
		Projects: map[string]projects.Project{},
	}
}

func (config Config) ResolveProjectByDir(dir string) *projects.Project {
	for _, project := range config.Projects {
		if strings.HasPrefix(project.Dir, dir) {
			return &project
		}
	}

	return nil
}

type AddProjectResult struct {
	DoesComposeExist  bool
	DoMigrationsExist bool
}

func (config *Config) AddProject(project projects.Project) (*AddProjectResult, error) {
	for _, existingProject := range config.Projects {
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

	if len(config.Projects) == 0 {
		config.Projects = map[string]projects.Project{
			project.Alias: project,
		}
	} else {
		config.Projects[project.Alias] = project
	}

	return &result, nil
}

type NoProjectFound struct {
	Alsdir string
}

func (config *Config) RemoveProject(alsdir string) (*projects.Project, *flr.Failure) {
	var foundProject *projects.Project
	for _, project := range config.Projects {
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

	delete(config.Projects, foundProject.Alias)

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

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(*configPath)
	if err != nil {
		config := emptyConfig()
		err = SaveConfig(config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	}

	var config *Config
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func SaveConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	content, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(*configPath), CONFIG_PERMISSION)
	if err != nil {
		return err
	}
	return os.WriteFile(*configPath, content, CONFIG_PERMISSION)
}
