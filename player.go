package main

import (
	// "fmt"
	"context"
	"io"
	"log/slog"

	// "context"
	"sync"
)

type Player struct {
	ID            string
	NickName      string
	LastLoginTime string
	Conn          ProtocolConn
	Hunger        int
	Room          *Room
	Notify        chan *CommandResult
	Command       chan *Command
	Error         chan error
	Ticker        chan struct{}
	Inventory     map[string]*Inventory
	Mutex         sync.RWMutex
	TalkingNPC    string
}

func NewPlayer(ctx context.Context, nickName string, room *Room, conn ProtocolConn) *Player {
	player := &Player{
		ID:        "",
		NickName:  nickName,
		Room:      room,
		Conn:      conn,
		Notify:    make(chan *CommandResult, 10),
		Command:   make(chan *Command),
		Error:     make(chan error),
		Inventory: make(map[string]*Inventory),
	}
	room.Arrival <- player
	player.Read(ctx)
	player.Run(ctx)
	return player
}
func (p *Player) Read(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				str, err := p.Conn.ReadLine()
				if err == io.EOF {
					slog.Debug("player exit")
					p.Error <- err
					//maybe it's  a goodtime to save user's data
					break
				}
				cmd := Command{}

				if p.TalkingNPC != "" {
					if str == "bye" {
						p.Notify <- &CommandResult{0, "You end the conversation with " + p.TalkingNPC}
						p.TalkingNPC = ""
						continue
					} else {
						cmd.Verb = "talk"
						cmd.Args = []string{p.TalkingNPC, str}
						cmd.Raw = "talk " + p.TalkingNPC + " " + str
					}
				} else {
					cmd.Parse(str)
				}
				cmd.Player = p
				slog.Debug("player read command", "cmd", cmd.Raw)
				world.Commands <- &cmd
			}
		}
	}()
}

func (p *Player) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ret := <-p.Notify:
			slog.Debug("user gets notified:", "msg", ret.Msg)
			p.Conn.WriteLine(ret.Msg)
		case err := <-p.Error:
			slog.Debug("get error:", "err", err)
			return
		case <-p.Ticker:
			slog.Debug("tiker singal")
			p.OnTick()
		case cmd := <-p.Command:
			slog.Debug("player cmnd", "cmd", cmd)
			//exeute
		}
	}
}

func (p *Player) OnTick() {
	//each turn player gets  hungrier,hunger reaches 100 means player is starving
	p.Mutex.Lock()
	if p.Hunger < 100 {
		p.Hunger++
	}
	p.Mutex.Unlock()
}
