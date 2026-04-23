package event

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ValidTypes is the allowlist of event types.
var ValidTypes = map[string]bool{
	"task_published":   true,
	"task_adopted":     true,
	"task_re_adopted":  true,
	"task_advanced":    true,
	"run_completed":    true,
	"run_failed":       true,
	"verify_passed":    true,
	"verify_failed":    true,
	"checkup_done":     true,
	"loop_started":     true,
	"loop_halted":      true,
	"loop_resumed":     true,
	"loop_completed":   true,
	"loop_cancelled":   true,
}

// Record writes a new event file under .mino/events/issue-{N}/.
// It validates the event type, auto-assigns sequence number, and persists.
func Record(repoRoot string, issue int, eventType string, fields map[string]string) error {
	if !ValidTypes[eventType] {
		return fmt.Errorf("invalid event type: %q (not in allowlist)", eventType)
	}

	eventsDir := filepath.Join(repoRoot, ".mino", "events", fmt.Sprintf("issue-%d", issue))
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return fmt.Errorf("mkdir events: %w", err)
	}

	seq := nextSequence(eventsDir)
	path := filepath.Join(eventsDir, fmt.Sprintf("%04d-%s.yml", seq, eventType))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s at %s.\n\n", eventType, time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("```yaml\niron_tree:\n")
	sb.WriteString("  version: 1\n")
	sb.WriteString(fmt.Sprintf("  event_type: %s\n", eventType))
	sb.WriteString(fmt.Sprintf("  sequence: %d\n", seq))
	sb.WriteString(fmt.Sprintf("  timestamp: %s\n", time.Now().UTC().Format(time.RFC3339)))
	for k, v := range fields {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}
	sb.WriteString("```\n")

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	return nil
}

func nextSequence(dir string) int {
	entries, _ := os.ReadDir(dir)
	return len(entries) + 1
}
