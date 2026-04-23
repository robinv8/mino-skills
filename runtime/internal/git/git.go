package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo wraps git operations for a given repository root.
type Repo struct {
	Root string
}

// IsDirty returns true if the working tree has uncommitted changes.
func (r *Repo) IsDirty() (bool, error) {
	out, err := exec.Command("git", "-C", r.Root, "status", "--short").Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// Stage adds files to the index. If paths is empty, stages all tracked changes.
func (r *Repo) Stage(paths ...string) error {
	if len(paths) == 0 {
		// Stage all tracked changes only (safer than -A which might pick up untracked)
		out, err := exec.Command("git", "-C", r.Root, "add", "-u").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git add -u: %s", string(out))
		}
		return nil
	}

	// Stage specific files
	args := append([]string{"-C", r.Root, "add"}, paths...)
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add %v: %s", paths, string(out))
	}
	return nil
}

// HasStagedChanges returns true if the index has staged content.
func (r *Repo) HasStagedChanges() (bool, error) {
	// --cached --quiet exits 1 if there are differences
	err := exec.Command("git", "-C", r.Root, "diff", "--cached", "--quiet").Run()
	if err != nil {
		// Exit error means there ARE staged changes
		return true, nil
	}
	return false, nil
}

// ErrNothingToCommit indicates the index is empty after staging.
var ErrNothingToCommit = fmt.Errorf("nothing to commit")

// Commit creates a commit with the given message.
// Returns ErrNothingToCommit if the index is empty.
func (r *Repo) Commit(msg string) (string, error) {
	hasStaged, err := r.HasStagedChanges()
	if err != nil {
		return "", err
	}
	if !hasStaged {
		return "", ErrNothingToCommit
	}

	out, err := exec.Command("git", "-C", r.Root, "commit", "-m", msg).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git commit: %s", string(out))
	}

	// Extract commit hash
	hashOut, err := exec.Command("git", "-C", r.Root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(hashOut)), nil
}

// Push pushes the current branch to origin.
func (r *Repo) Push() error {
	out, err := exec.Command("git", "-C", r.Root, "push").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push: %s", string(out))
	}
	return nil
}

// CurrentBranch returns the current git branch name.
func (r *Repo) CurrentBranch() (string, error) {
	out, err := exec.Command("git", "-C", r.Root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// EnsureMinoIgnored adds .mino/ to .gitignore if not already present.
func (r *Repo) EnsureMinoIgnored() error {
	gitignore := filepath.Join(r.Root, ".gitignore")
	data, err := os.ReadFile(gitignore)
	if err != nil {
		// No .gitignore yet — create one
		return os.WriteFile(gitignore, []byte(".mino/\n"), 0644)
	}
	if strings.Contains(string(data), ".mino/") {
		return nil
	}
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n.mino/\n")
	return err
}
