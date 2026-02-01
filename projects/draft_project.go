package projects

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DraftProject struct {
	Alias      string
	Dir        *string
	Compose    *string
	Migrations *string
}

func (draft DraftProject) unwrapDir() (*string, error) {
	var dir string
	var err error
	if draft.Dir == nil || *draft.Dir == "" {
		dir, err = os.Getwd()
	}

	if err != nil {
		return nil, err
	}

	if dir != "" {
		return &dir, nil
	} else {
		dir = *draft.Dir
	}

	info, err := os.Stat(dir)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("could not find the directory: %s", dir), err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("found file instead of directory at: %s", dir)
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return &dir, nil
}

func (draft DraftProject) ToProject() (*Project, error) {
	dir, err := draft.unwrapDir()
	if err != nil {
		return nil, err
	}

	return &Project{
		Alias:      draft.Alias,
		Dir:        *dir,
		Compose:    draft.Compose,
		Migrations: draft.Migrations,
	}, nil
}
