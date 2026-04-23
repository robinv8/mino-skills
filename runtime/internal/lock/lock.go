package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Info represents the content of a lock file.
type Info struct {
	PID      int
	Reason   string
	Acquired time.Time
}

// Path returns the lock file path.
func Path(repoRoot string) string {
	return filepath.Join(repoRoot, ".mino", "run.lock")
}

// Acquire creates the lock file if it does not already exist.
// If the lock is held by a dead process, it steals it.
func Acquire(repoRoot, reason string) error {
	p := Path(repoRoot)

	// Check existing lock
	if info, err := read(p); err == nil {
		if isProcessAlive(info.PID) {
			return fmt.Errorf("lock held by PID %d since %s (%s)",
				info.PID, info.Acquired.Format(time.RFC3339), info.Reason)
		}
		// Stale lock — overwrite
	}

	data := fmt.Sprintf("pid: %d\nreason: %s\nacquired: %s\n",
		os.Getpid(), reason, time.Now().UTC().Format(time.RFC3339))
	return os.WriteFile(p, []byte(data), 0644)
}

// Release removes the lock file.
func Release(repoRoot string) error {
	return os.Remove(Path(repoRoot))
}

// Check returns true if a valid lock exists, plus its info.
func Check(repoRoot string) (bool, *Info, error) {
	info, err := read(Path(repoRoot))
	if err != nil {
		return false, nil, nil // no lock
	}
	alive := isProcessAlive(info.PID)
	return alive, info, nil
}

func read(path string) (*Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	info := &Info{}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "pid: ") {
			fmt.Sscanf(line, "pid: %d", &info.PID)
		}
		if strings.HasPrefix(line, "reason: ") {
			info.Reason = strings.TrimPrefix(line, "reason: ")
		}
		if strings.HasPrefix(line, "acquired: ") {
			s, _ := time.Parse(time.RFC3339, strings.TrimPrefix(line, "acquired: "))
			info.Acquired = s
		}
	}
	return info, nil
}

func isProcessAlive(pid int) bool {
	// signal 0 is the Unix idiom for "check if process exists"
	err := syscall.Kill(pid, 0)
	return err == nil
}
