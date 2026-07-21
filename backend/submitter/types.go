package submitter

import "shared/golog"

var subLog = golog.New("Submitter")

type SubmitError struct {
	Msg         string
	RateLimited bool
}

func (e *SubmitError) Error() string {
	return e.Msg
}

type SeenEntry struct {
	Status    FlagResultStatus
	TeamId    int
	Submitter string
	Timestamp int64
}

type FlagEntry struct {
	Value     string
	Status    FlagResultStatus
	TeamId    int
	Submitter string
	Timestamp int64
}

type FlagResultStatus int

const (
	StatusUnknown FlagResultStatus = iota
	StatusPending
	StatusAccepted
	StatusDuplicate
	StatusTooOld
)

type FlagSummary struct {
	Total    int
	Accepted int
	Pending  int
	Rejected int
}

type TopAttackerEntry struct {
	Username string `json:"username"`
	Accepted int    `json:"accepted"`
	Total    int    `json:"total"`
}

type TimelinePoint struct {
	Timestamp int64 `json:"timestamp"`
	Accepted  int   `json:"accepted"`
	Rejected  int   `json:"rejected"`
}

type FlagResult struct {
	Value  string
	Status FlagResultStatus
}

type Submitter interface {
	Submit(flags []Flag) ([]FlagResult, *SubmitError)
}

type Flag struct {
	Value     string
	TeamId    int
	Submitter string
	Service   string
}

func (s FlagResultStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusAccepted:
		return "accepted"
	case StatusDuplicate:
		return "duplicate"
	case StatusTooOld:
		return "too_old"
	default:
		return "unknown"
	}
}
