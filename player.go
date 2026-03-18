package main

import(
	// "fmt"
	"log/slog"
	"io"
	// "context"
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
		world.Commands <- &cmd
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
			slog.Debug("player cmnd","cmd",cmd)
			
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