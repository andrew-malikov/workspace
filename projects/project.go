package projects

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Alias      string
	Dir        string
	Compose    *string
	Migrations *string
}

func (project Project) EnsureValidDir() error {
	info, err := os.Stat(project.Dir)
	if err != nil {
		return errors.Join(fmt.Errorf("Could not find the directory: %s", project.Dir), err)
	}

	if !info.IsDir() {
		return fmt.Errorf("Found file instead of directory at: %s", project.Dir)
	}

	return nil
}

func (project Project) DoesComposeExist() bool {
	if project.Compose == nil {
		return false
	}

	info, err := os.Stat(filepath.Join(project.Dir, *project.Compose))
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

func (project Project) DoMigrationsExist() bool {
	if project.Migrations == nil {
		return false
	}

	info, err := os.Stat(filepath.Join(project.Dir, *project.Migrations))
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}

func (project *Project) ResetCompose() {
	project.Compose = nil
}

func (project *Project) ResetMigrations() {
	project.Migrations = nil
}
