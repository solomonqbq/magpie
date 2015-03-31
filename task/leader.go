package main

import (
	"fmt"
	"time"
)

//代理组长接口，组长有权分配组内成员任务，同一时间一个组内只能有一个组长
type ProxyLeader struct {
	//宣告成为组长
	AnnounceLeader func() bool
	//宣告间隔
	anounce_interval time.Duration
	//是否是组长
	is_leader bool
	//运行状态
	running bool
}

func NewProxyLeader(announceLeader func() bool) *ProxyLeader {
	pl := new(ProxyLeader)
	pl.AnnounceLeader = announceLeader
	return pl
}
func main() {
	plx := new(ProxyLeader)
	plx.AnnounceLeader = func() bool { return true }
	plx.Start()
	fmt.Println(plx.AnnounceLeader())
}

//当前是否是组长
func (p *ProxyLeader) IsLeader() bool {
	return p.is_leader && p.running
}

//启动运行
func (p *ProxyLeader) Start() {
	p.running = true
	go func() {
		for p.running {
			p.is_leader = p.AnnounceLeader()
			time.Sleep(p.anounce_interval)
		}
	}()
	return
}

func (p *ProxyLeader) Stop() {
	p.running = false
}
