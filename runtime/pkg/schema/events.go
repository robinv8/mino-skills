package schema

import "time"

// Event is pushed to all connected WebSocket clients.
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// HaltReason describes why an issue loop halted.
type HaltReason string

const (
	HaltApprovalRequired  HaltReason = "approval-required"
	HaltPendingAcceptance HaltReason = "pending_acceptance"
	HaltFailTerminal      HaltReason = "fail_terminal"
	HaltBlocked           HaltReason = "blocked"
	HaltBudgetExhausted   HaltReason = "budget_exhausted"
)

// CommandResult is the outcome of a posted command.
type CommandResult struct {
	Success      bool   `json:"success"`
	Advanced     string `json:"advanced,omitempty"`
	Error        string `json:"error,omitempty"`
	DryRunOutput string `json:"dry_run_output,omitempty"`
	DurationMs   int    `json:"duration_ms,omitempty"`
}

// IssueEvent is a single persisted event for an issue.
type IssueEvent struct {
	Seq       int                    `json:"seq"`
	Issue     int                    `json:"issue"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// Event type constants used by Runtime and Native Client.
const (
	EventConnected         = "connected"
	EventPing              = "ping"
	EventPong              = "pong"
	EventTaskAdvanced      = "task_advanced"
	EventTaskSkipped       = "task_skipped"
	EventTaskCancelled     = "task_cancelled"
	EventTaskRetried       = "task_retried"
	EventCommandCompleted  = "command_completed"
	EventLoopHalted        = "loop_halted"
	EventError             = "error"
)
