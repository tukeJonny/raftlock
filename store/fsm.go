package store

import (
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

var (
	ErrUnmarshalLockEvt = errors.New("Failed to unmarshal lock event")
	ErrUnknownOperation = errors.New("Unknown operation")
)

type fsm Store

func (f *fsm) Apply(l *raft.Log) interface{} {
	var evt lockEvent
	if err := json.Unmarshal(l.Data, &evt); err != nil {
		panic(ErrUnmarshalLockEvt)
	}

	switch newLockOp(evt.op) {
	case opAcquire:
		return f.applyAcquire(evt.id)
	case opRelease:
		return f.applyRelease(evt.id)
	default:
		panic(ErrUnknownOperation)
	}
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	clone := make(map[string]string)
	f.m.Range(func(k, v interface{}) bool {
		var (
			key   = k.(string)
			value = k.(string)
		)
		clone[key] = value

		return true
	})
	return &snapshot{store: clone}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	store := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&store); err != nil {
		return err
	}

	next := sync.Map{}
	for k, v := range store {
		next.Store(k, v)
	}
	f.m = next

	return nil
}

func (f *fsm) applyAcquire(id string) interface{} {
	f.m.Store(id, id)
	return nil
}

func (f *fsm) applyRelease(id string) interface{} {
	f.m.Delete(id)
	return nil
}
