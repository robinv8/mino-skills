package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/robinv8/mino-runtime/internal/brief"
	"github.com/robinv8/mino-runtime/internal/event"
	"github.com/robinv8/mino-runtime/internal/git"
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
  mino run    <issue>          Full run cycle: pre-flight + lock + step + git + release
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

	advanceAndSave(root, issue, b)
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

	// 3. Advance state
	fmt.Println("[run] advancing state...")
	b, err := brief.Load(root, issue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load brief: %v\n", err)
		os.Exit(1)
	}

	from, next := advanceAndSave(root, issue, b)

	// 4. Git: stage + commit if dirty
	repo := &git.Repo{Root: root}
	dirty, _ := repo.IsDirty()
	if dirty {
		fmt.Println("[git] staging changes...")
		if err := repo.Stage(); err != nil {
			fmt.Fprintf(os.Stderr, "git stage: %v\n", err)
			os.Exit(1)
		}

		msg := fmt.Sprintf("[run] #%d: %s (%s → %s)", issue, b.TaskKey, from, next)
		fmt.Println("[git] committing...")
		sha, err := repo.Commit(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "git commit: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[git] committed: %s\n", sha)
	} else {
		fmt.Println("[git] no changes to commit")
	}

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

// advanceAndSave advances the brief state, saves it, records the event.
// Returns (from, next) stages for logging.
func advanceAndSave(root string, issue int, b *brief.Brief) (state.Stage, state.Stage) {
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

	_ = event.Record(root, issue, "task_advanced", map[string]string{
		"from": string(from),
		"to":   string(next),
	})

	fmt.Printf("Issue #%d advanced: %s → %s\n", issue, from, next)
	return from, next
}
