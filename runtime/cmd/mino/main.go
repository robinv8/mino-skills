package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/robinv8/mino-runtime/internal/brief"
	"github.com/robinv8/mino-runtime/internal/lock"
	"github.com/robinv8/mino-runtime/internal/preflight"
	"github.com/robinv8/mino-runtime/internal/state"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "state":
		handleState()
	case "step":
		handleStep()
	case "run":
		handleRun()
	case "check":
		handleCheck()
	case "version":
		fmt.Println("mino 0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`mino — Iron Tree Runtime (Phase 1)

Usage:
  mino state  <issue>          Show current stage of an issue
  mino step   <issue>          Advance to next stage (acquires lock)
  mino run    <issue>          Full run cycle: pre-flight + lock + step + release
  mino check                   Run pre-flight checks for current repo
  mino version                 Show version

Examples:
  mino state issue-23
  mino state 23
  mino run issue-23`)
}

func resolveRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".mino")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return cwd, nil
}

func parseIssue(arg string) (int, error) {
	arg = strings.TrimPrefix(arg, "issue-")
	return strconv.Atoi(arg)
}

func handleState() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: mino state <issue>")
		os.Exit(1)
	}
	issue, err := parseIssue(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid issue: %v\n", err)
		os.Exit(1)
	}

	root, err := resolveRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo: %v\n", err)
		os.Exit(1)
	}

	b, err := brief.Load(root, issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load brief: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Issue #%d (%s)\n", b.IssueNumber, b.TaskKey)
	fmt.Printf("  Stage:        %s\n", b.CurrentStage)
	fmt.Printf("  Next Stage:   %s\n", b.NextStage)
	fmt.Printf("  Attempt:      %d/%d\n", b.AttemptCount, b.MaxRetryCount)
	fmt.Printf("  Spec Rev:     %s\n", b.SpecRevision)
}

func handleStep() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: mino step <issue>")
		os.Exit(1)
	}
	issue, err := parseIssue(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid issue: %v\n", err)
		os.Exit(1)
	}

	root, err := resolveRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo: %v\n", err)
		os.Exit(1)
	}

	// Acquire lock
	if err := lock.Acquire(root, fmt.Sprintf("step issue-%d", issue)); err != nil {
		fmt.Fprintf(os.Stderr, "lock: %v\n", err)
		os.Exit(1)
	}
	defer lock.Release(root)

	b, err := brief.Load(root, issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load brief: %v\n", err)
		os.Exit(1)
	}

	from := state.Stage(b.CurrentStage)
	next, err := state.DefaultNext(from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot advance from %s: %v\n", from, err)
		os.Exit(1)
	}

	updated := b.Patch("Current Stage", b.CurrentStage, string(next))
	nextNext, _ := state.DefaultNext(next)
	if nextNext != "" {
		updated = updated.Patch("Next Stage", b.NextStage, string(nextNext))
	}
	if from == state.StageTask {
		updated = updated.Patch("Attempt Count", fmt.Sprintf("%d", updated.AttemptCount), fmt.Sprintf("%d", updated.AttemptCount+1))
	}

	if err := updated.Save(root); err != nil {
		fmt.Fprintf(os.Stderr, "save brief: %v\n", err)
		os.Exit(1)
	}

	_ = writeEvent(root, issue, "task_advanced", map[string]string{
		"from": string(from),
		"to":   string(next),
	})

	fmt.Printf("Issue #%d advanced: %s → %s\n", issue, from, next)
}

func handleRun() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: mino run <issue>")
		os.Exit(1)
	}
	issue, err := parseIssue(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid issue: %v\n", err)
		os.Exit(1)
	}

	root, err := resolveRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo: %v\n", err)
		os.Exit(1)
	}

	// 1. Pre-flight
	fmt.Println("[preflight] checking environment...")
	res := preflight.Run(root)
	if !res.Ok() {
		fmt.Fprintln(os.Stderr, res.String())
		os.Exit(1)
	}
	if len(res.Warnings) > 0 {
		fmt.Println(res.String())
	}
	fmt.Println("[preflight] ok")

	// 2. Lock
	fmt.Println("[lock] acquiring...")
	if err := lock.Acquire(root, fmt.Sprintf("run issue-%d", issue)); err != nil {
		fmt.Fprintf(os.Stderr, "lock: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		fmt.Println("[lock] releasing...")
		lock.Release(root)
	}()

	// 3. Step
	fmt.Println("[run] advancing state...")
	b, err := brief.Load(root, issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load brief: %v\n", err)
		os.Exit(1)
	}

	from := state.Stage(b.CurrentStage)
	next, err := state.DefaultNext(from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot advance from %s: %v\n", from, err)
		os.Exit(1)
	}

	updated := b.Patch("Current Stage", b.CurrentStage, string(next))
	nextNext, _ := state.DefaultNext(next)
	if nextNext != "" {
		updated = updated.Patch("Next Stage", b.NextStage, string(nextNext))
	}
	if from == state.StageTask {
		updated = updated.Patch("Attempt Count", fmt.Sprintf("%d", updated.AttemptCount), fmt.Sprintf("%d", updated.AttemptCount+1))
	}

	if err := updated.Save(root); err != nil {
		fmt.Fprintf(os.Stderr, "save brief: %v\n", err)
		os.Exit(1)
	}

	_ = writeEvent(root, issue, "run_completed", map[string]string{
		"from": string(from),
		"to":   string(next),
	})

	fmt.Printf("[run] Issue #%d: %s → %s\n", issue, from, next)
	fmt.Println("[run] done")
}

func handleCheck() {
	root, err := resolveRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve repo: %v\n", err)
		os.Exit(1)
	}

	res := preflight.Run(root)
	fmt.Print(res.String())
	if !res.Ok() {
		os.Exit(1)
	}
}

func writeEvent(root string, issue int, eventType string, fields map[string]string) error {
	eventsDir := filepath.Join(root, ".mino", "events", fmt.Sprintf("issue-%d", issue))
	if err := os.MkdirAll(eventsDir, 0755); err != nil {
		return err
	}

	seq := 1
	entries, _ := os.ReadDir(eventsDir)
	seq += len(entries)

	path := filepath.Join(eventsDir, fmt.Sprintf("%04d-%s.yml", seq, eventType))
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s at %s.\n\n", eventType, time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("```yaml\niron_tree:\n")
	sb.WriteString(fmt.Sprintf("  version: 1\n"))
	sb.WriteString(fmt.Sprintf("  event_type: %s\n", eventType))
	sb.WriteString(fmt.Sprintf("  sequence: %d\n", seq))
	sb.WriteString(fmt.Sprintf("  timestamp: %s\n", time.Now().UTC().Format(time.RFC3339)))
	for k, v := range fields {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}
	sb.WriteString("```\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
