package state

import "fmt"

// Stage represents a workflow stage.
type Stage string

const (
	StageDefinition  Stage = "definition"
	StageTask        Stage = "task"
	StageRun         Stage = "run"
	StageVerify      Stage = "verify"
	StageCheckup     Stage = "checkup"
	StageDone        Stage = "done"
	StageHalted      Stage = "halted"
)

// ValidTransitions maps current stage -> allowed next stages.
var ValidTransitions = map[Stage][]Stage{
	StageDefinition:  {StageTask},
	StageTask:        {StageRun, StageHalted},
	StageRun:         {StageVerify, StageHalted},
	StageVerify:      {StageCheckup, StageRun, StageHalted},
	StageCheckup:     {StageDone, StageHalted},
	StageDone:        {},
	StageHalted:      {StageRun, StageTask}, // resume paths
}

// CanTransition checks if a stage change is valid.
func CanTransition(from, to Stage) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// DefaultNext returns the default next stage for a given stage.
func DefaultNext(from Stage) (Stage, error) {
	switch from {
	case StageDefinition:
		return StageRun, nil
	case StageTask:
		return StageRun, nil
	case StageRun:
		return StageVerify, nil
	case StageVerify:
		return StageCheckup, nil
	case StageCheckup:
		return StageDone, nil
	default:
		return "", fmt.Errorf("no default next stage from %s", from)
	}
}
