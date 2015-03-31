package core

import (
	"fmt"
	"testing"
	"time"
)

func TestMemBoard(t *testing.T) {
	b := NewMockBoard()
	b.Start()
	fmt.Println(b)
	m := NewMockMember(b, "test_group")
	m.Regist()
	m.Start()

	time.Sleep(30 * time.Second)
}
