package core

import (
	"testing"
	"time"
)

func TestMemBoard(t *testing.T) {
	b := NewMockBoard()
	b.Start()

	time.Sleep(30 * time.Second)
}
