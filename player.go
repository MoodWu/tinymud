package main

import(
	"fmt"
	"log/slog"
	"io"
	"context"
	"sync"
)


type Player struct{
	ID string
	NickName string
	LastLoginTime string
	Conn ProtocolConn
	Hunger int
	Room *Room
	Notify chan *CommandResult
	Command chan *Command
	Error chan error
	Ticker chan struct{}
	Inventory map[string]*Inventory
	Mutex sync.RWMutex
}


func NewPlayer(nickName string,room *Room,conn ProtocolConn) *Player {
	player := &Player{
		ID:"",
		NickName: nickName,
		Room: room,
		Conn:conn,
		Notify: make(chan *CommandResult,10),
		Command: make(chan *Command),
		Error: make(chan error),
		Inventory: make(map[string]*Inventory),
	}
	room.Arrival <- player
	player.Read()
	player.Run()
	return player
}
func (p *Player) Read(){
	go func(){
		for{
		str, err := p.Conn.ReadLine()
		if err == io.EOF {
			slog.Debug("player exit")
			p.Error <- err
			//maybe it's  a goodtime to save user's data
			break
		}
		cmd := Command{}
		cmd.Parse(str)
		cmd.Player = p
		p.Command <- &cmd
		}

	}()
}

func (p *Player) Run(){
	for{
		select{
		case ret := <- p.Notify:
			slog.Debug("user gets notified:","msg",ret.Msg)
			p.Conn.WriteLine(ret.Msg)
		case err := <- p.Error:
			slog.Debug("get error:","err",err)
			return
		case <-p.Ticker:
			slog.Debug("tiker singal")
			p.OnTick()
		case cmd := <- p.Command:
			
			slog.Debug("user command:","cmd",fmt.Sprintf("%#v",cmd))
			//parse command
			//Check command route
			worker,ok := CommandMap[cmd.Verb]
			if ok{
				slog.Debug("call worker")
			worker(context.Background(),p,cmd)
			break
			}
			//the command is not in the list,send directly to room ,to see if it's a script command
			if p.Room != nil {
				slog.Debug("Send to room")
				p.Room.Commands <- cmd
			}
			slog.Debug("Unknown commnad send directly to room")

			//exeute
		}
	}
}

func (p *Player) OnTick() {
	//each turn player gets  hungrier,hunger reaches 100 means player is starving
	if p.Hunger <100{
		p.Hunger++
	}	
}