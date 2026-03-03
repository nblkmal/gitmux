package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Repository struct {
	Path   string
	Name   string
	Branch string
	Status RepoStatus
	Active bool
}

type RepoStatus struct {
	IsClean   bool
	IsDirty   bool
	Ahead     int
	Behind    int
	Untracked int
}

func Discover(scanPaths []string, exclude []string) ([]Repository, error) {
	var repos []Repository
	excludeSet := make(map[string]bool)
	for _, e := range exclude {
		excludeSet[e] = true
	}

	for _, scanPath := range scanPaths {
		err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Check if this is a git directory (must check before exclude!)
			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)
				repos = append(repos, Repository{
					Path: repoPath,
					Name: filepath.Base(repoPath),
				})
				return filepath.SkipDir
			}

			// Skip excluded directories (but not .git - we handle that above)
			if info.Name() != ".git" && excludeSet[info.Name()] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error scanning %s: %w", scanPath, err)
		}
	}

	// Sort by name
	for i := 0; i < len(repos)-1; i++ {
		for j := i + 1; j < len(repos); j++ {
			if strings.ToLower(repos[i].Name) > strings.ToLower(repos[j].Name) {
				repos[i], repos[j] = repos[j], repos[i]
			}
		}
	}

	return repos, nil
}
