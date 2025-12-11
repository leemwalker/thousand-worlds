package state

import "time"

// AddLog adds a decision log entry, keeping only the last 50 entries
func (e *LivingEntityState) AddLog(action, reason string) {
	log := DecisionLog{
		Timestamp: time.Now().Unix(),
		Action:    action,
		Reason:    reason,
	}

	e.Logs = append(e.Logs, log)

	// Keep last 50
	if len(e.Logs) > 50 {
		e.Logs = e.Logs[len(e.Logs)-50:]
	}
}
