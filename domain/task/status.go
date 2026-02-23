package task

import "errors"

var (
	StatusUnknow     = Status{}
	StatusTodo       = Status{"todo"}
	StatusPending    = Status{"pending"}
	StatusInProgress = Status{"inprogress"}
	StatusCompleted  = Status{"completed"}
)

type Status struct {
	slug string
}

func (s Status) String() string {
	return s.slug
}

func (s Status) IsZero() bool {
	return s.slug == ""
}

func StatusFromString(s string) (Status, error) {
	switch s {
	case StatusTodo.slug:
		return StatusTodo, nil
	case StatusPending.slug:
		return StatusPending, nil
	case StatusInProgress.slug:
		return StatusInProgress, nil
	case StatusCompleted.slug:
		return StatusCompleted, nil
	default:
		return StatusUnknow, errors.New("unknow status: " + s)
	}
}

var validTransitions = map[Status][]Status{
	StatusTodo:       {StatusPending, StatusInProgress},
	StatusPending:    {StatusTodo, StatusInProgress},
	StatusInProgress: {StatusCompleted, StatusPending},
	StatusCompleted:  {},
}

func (s Status) CanTransitionTo(target Status) bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	for _, t := range allowed {
		if t == target {
			return true
		}
	}
	return false
}
