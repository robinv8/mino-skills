package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Config holds daemon configuration.
type Config struct {
	RepoRoot   string
	Addr       string
	PIDFile    string
	LogFile    string
	AutoRestart bool
}

// DefaultConfig returns config for the given repo.
func DefaultConfig(repoRoot string) Config {
	return Config{
		RepoRoot:    repoRoot,
		Addr:        ":8765",
		PIDFile:     filepath.Join(repoRoot, ".mino", "daemon.pid"),
		LogFile:     filepath.Join(repoRoot, ".mino", "daemon.log"),
		AutoRestart: true,
	}
}

// IsRunning checks if the daemon process is alive.
func IsRunning(cfg Config) (bool, int, error) {
	data, err := os.ReadFile(cfg.PIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, 0, err
	}

	// Check if process exists
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, nil
	}
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		return false, 0, nil
	}

	return true, pid, nil
}

// Start forks the current binary into background.
func Start(cfg Config) error {
	running, pid, _ := IsRunning(cfg)
	if running {
		return fmt.Errorf("daemon already running (PID %d)", pid)
	}

	// Fork into background
	cmd := exec.Command(os.Args[0], "serve", cfg.Addr)
	cmd.Dir = cfg.RepoRoot
	cmd.Env = os.Environ()

	// Redirect output to log file
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start detached
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	// Write PID file
	if err := os.WriteFile(cfg.PIDFile, []byte(fmt.Sprintf("%d\n", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}

	fmt.Printf("Daemon started (PID %d)\n", cmd.Process.Pid)
	fmt.Printf("Log: %s\n", cfg.LogFile)
	return nil
}

// Stop terminates the daemon process.
func Stop(cfg Config) error {
	running, pid, err := IsRunning(cfg)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("daemon not running")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal daemon: %w", err)
	}

	// Wait for process to exit
	for i := 0; i < 30; i++ {
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Remove PID file
	os.Remove(cfg.PIDFile)
	fmt.Println("Daemon stopped")
	return nil
}

// Status returns daemon status.
func Status(cfg Config) string {
	running, pid, _ := IsRunning(cfg)
	if running {
		return fmt.Sprintf("running (PID %d)", pid)
	}
	return "stopped"
}

// InstallSystemd generates a systemd service file.
func InstallSystemd(cfg Config, binaryPath string) error {
	service := fmt.Sprintf(`[Unit]
Description=Mino Runtime Daemon
After=network.target

[Service]
Type=simple
ExecStart=%s serve %s
WorkingDirectory=%s
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`, binaryPath, cfg.Addr, cfg.RepoRoot)

	path := filepath.Join(os.Getenv("HOME"), ".config", "systemd", "user", "mino-daemon.service")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(service), 0644); err != nil {
		return err
	}

	fmt.Printf("Systemd service written to: %s\n", path)
	fmt.Println("Enable: systemctl --user enable mino-daemon")
	fmt.Println("Start:  systemctl --user start mino-daemon")
	return nil
}

// InstallLaunchd generates a macOS launchd plist.
func InstallLaunchd(cfg Config, binaryPath string) error {
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>me.rnode.mino-daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>serve</string>
        <string>%s</string>
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>
`, binaryPath, cfg.Addr, cfg.RepoRoot, cfg.LogFile, cfg.LogFile)

	path := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", "me.rnode.mino-daemon.plist")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		return err
	}

	fmt.Printf("Launchd plist written to: %s\n", path)
	fmt.Println("Load:   launchctl load ~/Library/LaunchAgents/me.rnode.mino-daemon.plist")
	fmt.Println("Start:  launchctl start me.rnode.mino-daemon")
	return nil
}
