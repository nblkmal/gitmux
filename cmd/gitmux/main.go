package main

import (
	"fmt"
	"os"

	"github.com/nabilakmal/gitmux/internal/config"
	"github.com/nabilakmal/gitmux/internal/repo"
	"github.com/nabilakmal/gitmux/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Try to get last scanned path
	scanPath := ""
	lastPath, err := config.GetLastPath()
	if err == nil && lastPath != "" {
		// Verify the path exists
		if _, err := os.Stat(lastPath); err == nil {
			scanPath = lastPath
		}
	}

	// Fall back to current working directory
	if scanPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		scanPath = cwd
	}

	// Save this path for next time
	_ = config.SaveLastPath(scanPath)

	// Load configuration (for exclude patterns)
	cfg, err := config.Load()
	exclude := []string{"node_modules", "vendor", ".git", "target", "dist", "build"}
	if err == nil && len(cfg.Exclude) > 0 {
		exclude = cfg.Exclude
	}

	// Discover repositories
	repos, err := repo.Discover([]string{scanPath}, exclude)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering repos: %v\n", err)
		os.Exit(1)
	}

	// Load repo active states
	repoStates, _ := config.GetRepoStates()
	for i := range repos {
		// Default to active if not set, otherwise use saved state
		if savedState, exists := repoStates[repos[i].Path]; exists {
			repos[i].Active = savedState
		} else {
			repos[i].Active = true // Default to active
		}
	}

	// Start TUI
	p := tea.NewProgram(tui.NewModel(repos, scanPath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
