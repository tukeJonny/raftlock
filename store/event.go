package store

type lockOp string

const (
	opAcquire lockOp = "acquire"
	opRelease lockOp = "release"
)

func newLockOp(op string) lockOp {
	switch op {
	case "acquire":
		return opAcquire
	case "release":
		return opRelease
	default:
		return ""
	}
}

type lockEvent struct {
	op string `json:"op,omitempty"`
	id string `json:"id,omitempty"`
}
