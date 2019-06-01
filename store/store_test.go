package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	s := NewStore()
	s.RaftBind = "127.0.0.1:0"
	s.RaftDir = tmpDir
	assert.NotNil(t, s)
	assert.NoError(t, s.Open(true, "node0", 3, 2, 10*time.Second))
}

func TestOpenSingleNode(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	s := NewStore()
	s.RaftBind = "127.0.0.1:0"
	s.RaftDir = tmpDir
	assert.NotNil(t, s)
	assert.NoError(t, s.Open(true, "node9", 3, 2, 10*time.Second))

	time.Sleep(3 * time.Second)

	acquire := func(id string) error {
		_, err := s.Acquire(id)
		return err
	}
	tests := []struct {
		Fn       func(string) error
		LockName string
	}{
		{
			Fn:       acquire,
			LockName: "testlock",
		},
		{
			Fn:       s.Release,
			LockName: "testlock",
		},
	}

	for _, tt := range tests {
		assert.NoError(t, tt.Fn(tt.LockName))
		time.Sleep(500 * time.Millisecond)
	}
}
