package git

import (
	"fmt"

	"github.com/nabilakmal/gitmux/internal/repo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) GetStatus(r *repo.Repository) error {
	rPath := r.Path

	// Open the repository
	gitRepo, err := git.PlainOpen(rPath)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	// Get current branch
	head, err := gitRepo.Head()
	if err != nil {
		return fmt.Errorf("failed to get head: %w", err)
	}

	r.Branch = head.Name().Short()

	// Get worktree status
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	r.Status.IsClean = status.IsClean()
	r.Status.IsDirty = !status.IsClean()

	// Count untracked files
	untracked := 0
	for _, st := range status {
		if st.Staging == git.Untracked {
			untracked++
		}
	}
	r.Status.Untracked = untracked

	return nil
}

func (s *Service) Fetch(r *repo.Repository) error {
	gitRepo, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	err = gitRepo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch failed: %w", err)
	}

	return nil
}

func (s *Service) Pull(r *repo.Repository) error {
	gitRepo, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	worktree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Force:      true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

func (s *Service) Checkout(r *repo.Repository, branchName string) error {
	gitRepo, err := git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	worktree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// First fetch to get latest remote branches
	err = gitRepo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		// Continue anyway, fetch might fail but checkout might still work
	}

	// First try local branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		Force:  true,
	})
	if err == nil {
		r.Branch = branchName
		return nil
	}

	// Try remote branch - create local tracking branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/remotes/origin/" + branchName),
		Force:  true,
	})
	if err == nil {
		r.Branch = branchName
		return nil
	}

	// Try to create new branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		Create: true,
		Force:  true,
	})
	if err == nil {
		r.Branch = branchName
		return nil
	}

	return fmt.Errorf("branch '%s' not found. Use ': branch -a' to see all branches", branchName)
}

func (s *Service) GetBranches(r *repo.Repository) ([]string, error) {
	gitRepo, err := git.PlainOpen(r.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	branches, err := gitRepo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	var branchList []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchList = append(branchList, ref.Name().Short())
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	// Also get remote branches
	remote, err := gitRepo.Remote("origin")
	if err == nil {
		refs, err := remote.List(&git.ListOptions{})
		if err == nil {
			for _, ref := range refs {
				name := ref.Name().Short()
				if !contains(branchList, name) {
					branchList = append(branchList, name)
				}
			}
		}
	}

	return branchList, nil
}

func (s *Service) RunCommand(r *repo.Repository, cmd string) (string, error) {
	// Simple implementation - just return the command for now
	// Could use exec.Command for actual execution
	return fmt.Sprintf("Would run: git -C %s %s", r.Path, cmd), nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var NoErrAlreadyUpToDate = git.NoErrAlreadyUpToDate
