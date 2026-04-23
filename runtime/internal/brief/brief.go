package brief

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robinv8/mino-runtime/pkg/schema"
)

// Brief represents a parsed local brief file.
type Brief struct {
	TaskKey       string
	IssueNumber   int
	SpecRevision  string
	CurrentStage  string
	NextStage     string
	AttemptCount  int
	MaxRetryCount int
	Raw           string
}

// Load reads a brief file from .mino/briefs/issue-{N}.md.
// Validates schema before returning.
func Load(repoRoot string, issueNumber int) (*Brief, error) {
	path := filepath.Join(repoRoot, ".mino", "briefs", fmt.Sprintf("issue-%d.md", issueNumber))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read brief: %w", err)
	}
	if errs := schema.ValidateBrief(string(data)); len(errs) > 0 {
		return nil, fmt.Errorf("schema validation failed: %v", errs)
	}
	return Parse(string(data)), nil
}

// Parse extracts structured fields from raw markdown.
func Parse(raw string) *Brief {
	b := &Brief{Raw: raw}

	// Task Key
	if m := re(`Task Key:\s*(.+)`).FindStringSubmatch(raw); len(m) > 1 {
		b.TaskKey = strings.TrimSpace(m[1])
	}
	// Issue Number
	if m := re(`Issue Number:\s*(\d+)`).FindStringSubmatch(raw); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &b.IssueNumber)
	}
	// Spec Revision
	if m := re(`Spec Revision:\s*([a-f0-9]+)`).FindStringSubmatch(raw); len(m) > 1 {
		b.SpecRevision = m[1]
	}
	// Current Stage
	if m := re(`Current Stage:\s*(\w+)`).FindStringSubmatch(raw); len(m) > 1 {
		b.CurrentStage = m[1]
	}
	// Next Stage
	if m := re(`Next Stage:\s*(\w+)`).FindStringSubmatch(raw); len(m) > 1 {
		b.NextStage = m[1]
	}
	// Attempt Count
	if m := re(`Attempt Count:\s*(\d+)`).FindStringSubmatch(raw); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &b.AttemptCount)
	}
	// Max Retry Count
	if m := re(`Max Retry Count:\s*(\d+)`).FindStringSubmatch(raw); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &b.MaxRetryCount)
	}

	return b
}

// Save writes the brief back to disk.
// Validates schema before writing to prevent persisting malformed state.
func (b *Brief) Save(repoRoot string) error {
	if errs := schema.ValidateBrief(b.Raw); len(errs) > 0 {
		return fmt.Errorf("schema validation failed before save: %v", errs)
	}
	path := filepath.Join(repoRoot, ".mino", "briefs", fmt.Sprintf("issue-%d.md", b.IssueNumber))
	return os.WriteFile(path, []byte(b.Raw), 0644)
}

// Patch updates a field in the raw markdown and returns a new Brief.
func (b *Brief) Patch(field, oldVal, newVal string) *Brief {
	pattern := fmt.Sprintf("%s: %s", regexp.QuoteMeta(field), regexp.QuoteMeta(oldVal))
	re := regexp.MustCompile(pattern)
	b.Raw = re.ReplaceAllString(b.Raw, fmt.Sprintf("%s: %s", field, newVal))
	return Parse(b.Raw)
}

func re(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
