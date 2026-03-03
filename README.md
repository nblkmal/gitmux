# gitmux - Multi-Repository Git Manager

A terminal UI for managing multiple git repositories at once. Browse, fetch, pull, and checkout branches across all your repos easily.

## Features

- **Auto-discovery** - Automatically finds all git repositories in a directory
- **Active/Inactive filtering** - Mark repos as inactive to hide them from the list
- **Git operations** - Fetch, pull, checkout branches across multiple repos
- **Custom commands** - Run any git command on selected repos
- **Persistent state** - Remembers last scanned directory and active/inactive states

## Installation

### Option 1: Build from source

```bash
# Clone or navigate to the project
cd ~/Documents/runcloud/gitmux

# Build the binary
go build -o gitmux ./cmd/gitmux

# Move to your PATH (optional)
cp gitmux /usr/local/bin/
# or
cp gitmux ~/bin/  # and add ~/bin to your PATH
```

### Option 2: Run directly

```bash
cd ~/Documents/runcloud/gitmux
./gitmux
```

## Usage

### Running gitmux

Simply run the binary from any directory. It will:

1. Use the last scanned directory (if available)
2. Fall back to current working directory
3. Scan for all git repositories

```bash
gitmux
# or
./gitmux  # if in the same directory
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up/down |
| `↓` / `↑` | Navigate up/down |
| `Space` | Toggle active/inactive for selected repo |
| `a` | Show/hide inactive repos |
| `f` | Fetch all repos |
| `p` | Pull current repo |
| `P` | Pull all repos |
| `c` | Checkout branch |
| `:` | Run custom git command |
| `/` | Filter repos by name |
| `r` | Refresh all repo statuses |
| `?` | Toggle help |
| `q` | Quit |

### Examples

#### Fetch all repos
```
Press: f
```

#### Pull current repo
```
1. Navigate to the repo
2. Press: p
```

#### Pull all repos
```
Press: P
```

#### Checkout a branch
```
1. Navigate to the repo
2. Press: c
3. Enter branch name (e.g., main, develop, feature/xyz)
4. Press: Enter
```

#### Run custom git command
```
1. Navigate to the repo
2. Press: :
3. Enter git command (e.g., status, log --oneline -5, branch -a)
4. Press: Enter
```

#### Filter repos
```
1. Press: /
2. Enter search term
3. Press: Enter to apply
4. Press: Esc to clear filter
```

#### Hide a repo from the list
```
1. Navigate to the repo
2. Press: Space
   - Repo is marked as [inactive]
   - It will be hidden from the list
```

#### Show inactive repos
```
Press: a
```

## Configuration

### Config File (Optional)

Create `~/.gitmux.yaml` or `./config.yaml`:

```yaml
scan_paths:
  - ~/Documents/runcloud
  - ~/projects
exclude:
  - node_modules
  - vendor
  - target
  - dist
```

### State File

gitmux automatically saves state to `~/.gitmux-state`:

- Last scanned directory
- Active/inactive state for each repo

## Requirements

- Go 1.21+
- Terminal with ANSI color support

## Troubleshooting

### "Found 0 repos"
- Make sure the directory contains git repositories
- Check that `.git` folders exist in your projects

### TUI won't start
- Make sure you're running in a terminal
- Some SSH environments may not support TUI

### Checkout fails
- Ensure you have the correct branch name
- Use `: branch -a` to see all available branches
- Make sure you don't have uncommitted changes (or stash them first)
