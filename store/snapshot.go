package store

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type snapshot struct {
	store map[string]string
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	write := func() error {
		b, err := json.Marshal(s.store)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}
	persist := func() (err error) {
		err = write()
		if err != nil {
			sink.Cancel()
		}
		return err
	}

	return persist()
}

func (s *snapshot) Release() {}
