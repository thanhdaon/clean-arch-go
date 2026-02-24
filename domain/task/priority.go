package task

import "errors"

type Priority int

const (
	PriorityZero Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return ""
	}
}

func PriorityFromString(s string) (Priority, error) {
	switch s {
	case "low":
		return PriorityLow, nil
	case "medium":
		return PriorityMedium, nil
	case "high":
		return PriorityHigh, nil
	case "urgent":
		return PriorityUrgent, nil
	default:
		return PriorityZero, errors.New("unknown priority: " + s)
	}
}

func (p Priority) IsZero() bool {
	return p == PriorityZero
}
