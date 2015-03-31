package db

import (
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/xeniumd-china/magpie/global"
	"testing"
	"time"
)

func TestDBBoard(t *testing.T) {
	var err error
	global.Properties, err = properties.Load("../magpie.properties")
	if err != nil {
		fmt.Println(err)
	}

	b := NewDBBoard()
	b.Init()
	groups, err := b.LoadAllGroup()
	fmt.Println(groups)
	mems, err := b.LoadActiveMembers("load_balance")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mems)

	mems, err := b.LoadActiveMembers("load_balance")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mems)

	//	b.Cleanup("load_balance")
	time.Sleep(1 * time.Second)
}
