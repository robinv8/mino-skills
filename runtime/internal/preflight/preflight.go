package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result collects all findings.
type Result struct {
	Errors   []string
	Warnings []string
}

func (r *Result) Ok() bool {
	return len(r.Errors) == 0
}

func (r *Result) String() string {
	var sb strings.Builder
	for _, e := range r.Errors {
		sb.WriteString("ERROR: " + e + "\n")
	}
	for _, w := range r.Warnings {
		sb.WriteString("WARN:  " + w + "\n")
	}
	return sb.String()
}

// Run executes all pre-flight checks for the given repo root.
func Run(repoRoot string) *Result {
	r := &Result{}

	// 1. Required tools
	for _, tool := range []string{"git", "gh"} {
		if _, err := exec.LookPath(tool); err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("required tool not found: %s", tool))
		}
	}

	// 2. Git working tree
	if out, err := exec.Command("git", "-C", repoRoot, "status", "--short").Output(); err == nil {
		if len(strings.TrimSpace(string(out))) > 0 {
			r.Warnings = append(r.Warnings, "git working tree is dirty")
		}
	} else {
		r.Errors = append(r.Errors, fmt.Sprintf("git status failed: %v", err))
	}

	// 3. .mino/ directory structure
	for _, sub := range []string{"briefs", "events"} {
		p := filepath.Join(repoRoot, ".mino", sub)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			r.Warnings = append(r.Warnings, fmt.Sprintf(".mino/%s does not exist — will be created", sub))
		}
	}

	// 4. gh auth
	if out, err := exec.Command("gh", "auth", "status").CombinedOutput(); err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("gh not authenticated: %s", strings.TrimSpace(string(out))))
	}

	return r
}
