package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nabilakmal/gitmux/internal/config"
	"github.com/nabilakmal/gitmux/internal/git"
	"github.com/nabilakmal/gitmux/internal/repo"
)

var (
	titleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Bold(true)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)
	branchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	dirtyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("31"))
	cleanStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
)

type Model struct {
	repos          []repo.Repository
	filteredRepos  []repo.Repository
	list           list.Model
	gitService     *git.Service
	spinner        spinner.Model
	loading        bool
	loadingMessage string
	showHelp       bool
	showCommands   bool
	commandInput   textinput.Model
	showBranch     bool
	branchInput    textinput.Model
	filterInput    textinput.Model
	showFilter     bool
	showInactive   bool
	statusMessage  string
	errMessage     string
	scanPath       string
}

func NewModel(repos []repo.Repository, scanPath string) *Model {
	gitSvc := git.New()

	// Create list with empty items initially
	items := make([]list.Item, 0)
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Git Repositories"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))

	// Command input
	cmdInput := textinput.New()
	cmdInput.Placeholder = "Enter git command (e.g., status, log --oneline -5)"
	cmdInput.Prompt = "> "

	// Branch input
	branchInput := textinput.New()
	branchInput.Placeholder = "Enter branch name"
	branchInput.Prompt = "> "

	// Filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Filter repos..."
	filterInput.Prompt = "/ "

	m := &Model{
		repos:         repos,
		filteredRepos: repos,
		list:         l,
		gitService:   gitSvc,
		spinner:      s,
		commandInput: cmdInput,
		branchInput:  branchInput,
		filterInput:  filterInput,
		scanPath:    scanPath,
		loading:     true,
		loadingMessage: "Loading repository status...",
	}

	return m
}

// Init starts the background loading of repo statuses
func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadRepoStatuses)
}

// loadRepoStatuses loads git status for all repos in the background
func (m *Model) loadRepoStatuses() tea.Msg {
	for i := range m.repos {
		m.loadingMessage = fmt.Sprintf("Loading %s...", m.repos[i].Name)
		_ = m.gitService.GetStatus(&m.repos[i])
	}

	// Filter and update list items
	m.applyActiveFilter()

	m.loading = false
	m.loadingMessage = ""

	return nil
}

// applyActiveFilter filters repos based on Active status and showInactive flag
func (m *Model) applyActiveFilter() {
	m.filteredRepos = nil
	for _, r := range m.repos {
		if m.showInactive || r.Active {
			m.filteredRepos = append(m.filteredRepos, r)
		}
	}

	// Update list items
	items := make([]list.Item, len(m.filteredRepos))
	for i, r := range m.filteredRepos {
		items[i] = repoItem{repo: r}
	}
	m.list.SetItems(items)
}

type repoItem struct {
	repo repo.Repository
}

func (r repoItem) Title() string {
	status := ""
	if !r.repo.Active {
		status = "[inactive] "
	} else if r.repo.Status.IsDirty {
		status = "(dirty) "
	} else {
		status = "(clean) "
	}
	return status + r.repo.Name
}

func (r repoItem) Description() string {
	return "branch: " + r.repo.Branch + " | " + r.repo.Path
}

func (r repoItem) FilterValue() string {
	return r.repo.Name
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle command mode
		if m.showCommands {
			switch msg.String() {
			case "enter":
				m.showCommands = false
				if m.commandInput.Value() != "" {
					m.runCustomCommand()
				}
				m.commandInput.Reset()
				return m, nil
			case "esc":
				m.showCommands = false
				m.commandInput.Reset()
				return m, nil
			}
			var cmd tea.Cmd
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd
		}

		// Handle branch input mode
		if m.showBranch {
			switch msg.String() {
			case "enter":
				m.showBranch = false
				if m.branchInput.Value() != "" {
					m.checkoutBranch()
				}
				m.branchInput.Reset()
				return m, nil
			case "esc":
				m.showBranch = false
				m.branchInput.Reset()
				return m, nil
			}
			var cmd tea.Cmd
			m.branchInput, cmd = m.branchInput.Update(msg)
			return m, cmd
		}

		// Handle filter mode
		if m.showFilter {
			switch msg.String() {
			case "enter", "esc":
				m.showFilter = false
				m.filterInput.Reset()
				return m, nil
			}
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}

		// Main key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r", "ctrl+r":
			return m, m.refresh
		case "f":
			return m, m.fetchAll
		case "p":
			return m, m.pullCurrent
		case "P":
			return m, m.pullAll
		case "c":
			m.showBranch = true
			m.branchInput.Focus()
			return m, nil
		case ":":
			m.showCommands = true
			m.commandInput.Focus()
			return m, nil
		case "/":
			m.showFilter = true
			m.filterInput.Focus()
			return m, nil
		case "a":
			// Toggle showing inactive repos
			m.showInactive = !m.showInactive
			m.applyActiveFilter()
		case " ":
			// Toggle active state for selected repo
			m.toggleActive()
		case "?":
			m.showHelp = !m.showHelp
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 3)
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render(" gitmux "))
	s.WriteString(helpStyle.Render(" - Multi-repo Git Manager\n"))
	s.WriteString(strings.Repeat("─", 50))
	s.WriteString("\n")

	// Scan path
	if m.scanPath != "" {
		s.WriteString(helpStyle.Render("Scanning: "))
		s.WriteString(cleanStyle.Render(m.scanPath))
		s.WriteString("\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n")
	}

	// Selected repo info
	idx := m.list.Index()
	if idx >= 0 && idx < len(m.filteredRepos) {
		r := m.filteredRepos[idx]
		if r.Status.IsDirty {
			s.WriteString(dirtyStyle.Render("● dirty  "))
		} else {
			s.WriteString(cleanStyle.Render("● clean  "))
		}
		s.WriteString(branchStyle.Render("branch: " + r.Branch + "  "))
		s.WriteString(helpStyle.Render(r.Path + "\n"))
	}

	if m.showFilter {
		s.WriteString(m.filterInput.View())
		s.WriteString("\n")
	}

	// Main content
	if m.loading {
		s.WriteString(m.spinner.View())
		s.WriteString(" ")
		s.WriteString(m.loadingMessage)
		s.WriteString("\n")
	} else {
		s.WriteString(m.list.View())
	}

	s.WriteString("\n")

	// Status bar
	if m.errMessage != "" {
		s.WriteString(errorStyle.Render("Error: " + m.errMessage + "\n"))
		m.errMessage = ""
	}
	if m.statusMessage != "" {
		s.WriteString(successStyle.Render(m.statusMessage + "\n"))
		m.statusMessage = ""
	}

	// Command input
	if m.showCommands {
		s.WriteString(titleStyle.Render(" Custom Command "))
		s.WriteString(helpStyle.Render("(esc to cancel)\n"))
		s.WriteString(m.commandInput.View())
		s.WriteString("\n")
	}

	// Branch input
	if m.showBranch {
		s.WriteString(titleStyle.Render(" Checkout Branch "))
		s.WriteString(helpStyle.Render("(esc to cancel)\n"))
		s.WriteString(m.branchInput.View())
		s.WriteString("\n")
	}

	// Help bar
	if m.showHelp {
		s.WriteString(helpStyle.Render(`
Commands:
  j/k, ↑/↓    Navigate
  Space       Toggle active/inactive
  a           Show/hide inactive repos
  f           Fetch all repos
  p           Pull current repo
  P           Pull all repos
  c           Checkout branch
  :           Run custom git command
  /           Filter repos
  r, Ctrl+R   Refresh
  ?           Toggle help
  q           Quit
`))
	} else {
		// Show hint about inactive repos
		inactiveHint := ""
		if m.showInactive {
			inactiveHint = " [showing inactive]"
		}
		s.WriteString(helpStyle.Render(" Press ? for help" + inactiveHint))
	}

	return s.String()
}

// Actions
func (m *Model) refresh() tea.Msg {
	m.loading = true
	m.loadingMessage = "Refreshing repositories..."

	for i := range m.repos {
		_ = m.gitService.GetStatus(&m.repos[i])
	}

	m.applyFilter()

	m.loading = false
	return nil
}

func (m *Model) fetchAll() tea.Msg {
	m.loading = true
	m.loadingMessage = "Fetching all repositories..."

	fetched := 0
	failed := 0
	for i := range m.repos {
		if err := m.gitService.Fetch(&m.repos[i]); err != nil {
			failed++
		} else {
			fetched++
		}
	}

	m.loading = false
	if failed > 0 {
		m.statusMessage = fmt.Sprintf("✓ Fetched %d repos, %d failed", fetched, failed)
	} else {
		m.statusMessage = fmt.Sprintf("✓ Successfully fetched %d repositories", fetched)
	}
	return nil
}

func (m *Model) pullCurrent() tea.Msg {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.filteredRepos) {
		m.errMessage = "No repository selected"
		return nil
	}

	m.loading = true
	m.loadingMessage = "Pulling " + m.filteredRepos[idx].Name + "..."

	repoPath := m.filteredRepos[idx].Path
	var success bool

	// Find and update the repo in the original slice
	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			if err := m.gitService.Pull(&m.repos[i]); err != nil {
				m.errMessage = fmt.Sprintf("✗ Failed: %v", err)
			} else {
				m.statusMessage = fmt.Sprintf("✓ Pulled %s", m.repos[i].Name)
			}
			_ = m.gitService.GetStatus(&m.repos[i])
			success = true
			break
		}
	}

	// Update filtered repos and list
	if success {
		m.applyActiveFilter()
	}

	m.loading = false
	return nil
}

func (m *Model) pullAll() tea.Msg {
	m.loading = true
	m.loadingMessage = "Pulling all repositories..."

	pulled := 0
	failed := 0
	for i := range m.repos {
		if err := m.gitService.Pull(&m.repos[i]); err != nil {
			failed++
		} else {
			pulled++
		}
		// Refresh status after pull
		_ = m.gitService.GetStatus(&m.repos[i])
	}

	// Update list
	m.applyActiveFilter()

	m.loading = false
	if failed > 0 {
		m.statusMessage = fmt.Sprintf("✓ Pulled %d repos, %d failed", pulled, failed)
	} else {
		m.statusMessage = fmt.Sprintf("✓ Successfully pulled %d repositories", pulled)
	}
	return nil
}

func (m *Model) checkoutBranch() {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.filteredRepos) {
		m.errMessage = "No repository selected"
		return
	}

	branchName := m.branchInput.Value()
	if branchName == "" {
		m.errMessage = "Please enter a branch name"
		return
	}

	m.loading = true
	m.loadingMessage = "Checking out " + branchName + "..."

	repoPath := m.filteredRepos[idx].Path

	// Find and update the repo in the original slice
	var success bool
	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			if err := m.gitService.Checkout(&m.repos[i], branchName); err != nil {
				m.errMessage = fmt.Sprintf("Failed: %v", err)
			} else {
				m.statusMessage = fmt.Sprintf("✓ Switched to branch '%s' in %s", m.repos[i].Branch, m.repos[i].Name)
				_ = m.gitService.GetStatus(&m.repos[i])
				success = true
			}
			break
		}
	}

	// Update filtered repos and list
	if success {
		m.applyActiveFilter()
	}

	m.loading = false
}

func (m *Model) runCustomCommand() {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.filteredRepos) {
		m.errMessage = "No repository selected"
		return
	}

	cmdStr := m.commandInput.Value()
	m.loading = true
	m.loadingMessage = "Running: " + cmdStr

	// Use shell to run the command
	r := &m.filteredRepos[idx]
	parts := strings.Fields(cmdStr)
	args := append([]string{"-C", r.Path}, parts...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()

	m.loading = false

	if err != nil {
		m.errMessage = fmt.Sprintf("Command failed: %v\n%s", err, output)
	} else {
		m.statusMessage = fmt.Sprintf("Output:\n%s", output)
	}
}

func (m *Model) applyFilter() {
	filter := m.filterInput.Value()
	if filter == "" {
		m.filteredRepos = m.repos
	} else {
		filterLower := strings.ToLower(filter)
		m.filteredRepos = nil
		for _, r := range m.repos {
			if strings.Contains(strings.ToLower(r.Name), filterLower) ||
				strings.Contains(strings.ToLower(r.Path), filterLower) {
				m.filteredRepos = append(m.filteredRepos, r)
			}
		}
	}

	// Also filter by active status
	if !m.showInactive {
		activeRepos := m.filteredRepos
		m.filteredRepos = nil
		for _, r := range activeRepos {
			if r.Active {
				m.filteredRepos = append(m.filteredRepos, r)
			}
		}
	}

	// Update list
	items := make([]list.Item, len(m.filteredRepos))
	for i, r := range m.filteredRepos {
		items[i] = repoItem{repo: r}
	}
	m.list.SetItems(items)
}

// toggleActive toggles the active state for the selected repo
func (m *Model) toggleActive() {
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.filteredRepos) {
		return
	}

	// Find the repo in the original repos slice
	repoPath := m.filteredRepos[idx].Path
	for i := range m.repos {
		if m.repos[i].Path == repoPath {
			m.repos[i].Active = !m.repos[i].Active
			// Save the state
			_ = config.SetRepoState(repoPath, m.repos[i].Active)
			m.statusMessage = fmt.Sprintf("%s marked as %s", m.repos[i].Name, map[bool]string{true: "active", false: "inactive"}[m.repos[i].Active])
			break
		}
	}

	// Reapply filter
	m.applyFilter()
}
