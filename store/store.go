package store

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

var (
	ErrNotLeader           = errors.New("This node does not leader")
	ErrLockAlreadyAcquired = errors.New("The lock had already acquired")
	ErrLockNotFound        = errors.New("there are no lock for specified id")
)

type Store struct {
	m sync.Map

	snapshotCount int // as counter
	timeout       time.Time

	RaftDir  string
	RaftBind string

	r   *raft.Raft
	lgr *log.Logger
}

func NewStore() *Store {
	return &Store{
		m:   sync.Map{},
		lgr: log.New(os.Stderr, "[raftlock] ", log.LstdFlags),
	}
}

func (s *Store) Open(isLeader bool, localID string, maxPool, retainSnapshotCount int, timeout time.Duration) error {
	s.lgr.Printf("Opening ...")
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(localID)

	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftBind, addr, maxPool, timeout, os.Stderr)
	if err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return err
	}

	path := filepath.Join(s.RaftDir, "raft.db")
	boltDB, err := raftboltdb.NewBoltStore(path)
	if err != nil {
		return err
	}
	var (
		logStore    raft.LogStore    = boltDB
		stableStore raft.StableStore = boltDB
	)

	raftSystem, err := raft.NewRaft(cfg, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return err
	}
	s.r = raftSystem

	s.lgr.Printf("isLeader=%v\n", isLeader)
	if isLeader {
		s.lgr.Printf("bootstrap cluster ...\n")
		raftSystem.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      cfg.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		})
	}

	return nil
}

func (s *Store) Join(nodeID, addr string) error {
	isDuplicateIDOrAddr := func(server raft.Server) bool {
		var (
			isDupID   = server.ID == raft.ServerID(nodeID)
			isDupAddr = server.Address == raft.ServerAddress(addr)
		)
		return isDupID || isDupAddr
	}
	isMember := func(server raft.Server) bool {
		var (
			isDupID   = server.ID == raft.ServerID(nodeID)
			isDupAddr = server.Address == raft.ServerAddress(addr)
		)
		return isDupID && isDupAddr
	}

	cfg := s.r.GetConfiguration()
	if err := cfg.Error(); err != nil {
		return err
	}

	servers := cfg.Configuration().Servers
	for _, server := range servers {
		if isDuplicateIDOrAddr(server) {
			if isMember(server) {
				return nil
			}

			result := s.r.RemoveServer(server.ID, 0, 0)
			if err := result.Error(); err != nil {
				return err
			}
		}
	}

	s.lgr.Printf("add voter nodeID=%s, serverAddr=%s\n", nodeID, addr)
	indexFuture := s.r.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if indexFuture.Error() != nil {
		return fmt.Errorf("Failed to AddVoter(%s, %s): %s", nodeID, addr, indexFuture.Error().Error())
	}

	s.lgr.Printf("node %s at %s joined successfully", nodeID, addr)

	return nil
}

func (s *Store) Nodes() ([]raft.Server, error) {
	cfg := s.r.GetConfiguration()
	if err := cfg.Error(); err != nil {
		return nil, err
	}

	return cfg.Configuration().Servers, nil
}

func (s *Store) Stats() map[string]string {
	return s.r.Stats()
}

func (s *Store) Acquire(id string) (string, error) {
	if s.r.State() != raft.Leader {
		return "", ErrNotLeader
	}

	if _, ok := s.m.Load(id); ok {
		return "", ErrLockAlreadyAcquired
	}
	s.m.Store(id, id)
	return id, nil
}

func (s *Store) Release(id string) error {
	if s.r.State() != raft.Leader {
		return ErrNotLeader
	}

	if _, ok := s.m.Load(id); !ok {
		return ErrLockNotFound
	}
	s.m.Delete(id)
	return nil
}
