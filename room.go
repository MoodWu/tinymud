package main

import (
	"container/heap"
	"fmt"
	"log/slog"
	"sync"
)

type Room struct {
	ID        string
	Name      string
	Length    int
	Width     int
	Desc      string
	Exits     []*Exit
	Departure chan *Player
	Arrival   chan *Player
	Commands  chan *Command
	Players   []*Player
	Ticker    chan struct{}
	Tick      int
	Mutex     sync.RWMutex
	Items     map[string]*Inventory
	// respawn   map[int]map[string]IRespawn
	respawnHeap []*RespawnEvent
}

type RespawnEvent struct {
	Interval int
	NextTick int
	ItemName string
	Event    IRespawn
}

type Inventory struct {
	Item
	Count int
}

type Exit struct {
	Direction string
	Room      string
}

func (r *Room) Len() int {
	return len(r.respawnHeap)
}

func (r *Room) Less(i, j int) bool {
	return r.respawnHeap[i].NextTick < r.respawnHeap[j].NextTick
}

func (r *Room) Swap(i, j int) {
	r.respawnHeap[i], r.respawnHeap[j] = r.respawnHeap[j], r.respawnHeap[i]
}

func (r *Room) Push(x any) {
	event := x.(*RespawnEvent)
	r.respawnHeap = append(r.respawnHeap, event)
}

func (r *Room) Pop() any {
	n := len(r.respawnHeap)
	event := r.respawnHeap[n-1]
	r.respawnHeap = r.respawnHeap[:n-1]
	return event
}

func (r *Room) RegisterRespawnEvent(tick int, item string, ev IRespawn) {
	//slog.Debug("RegisterRespawnEvent","tick",tick,"item",item)
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	event := &RespawnEvent{
		Interval: tick,
		NextTick: r.Tick + tick,
		ItemName: item,
		Event:    ev,
	}

	//add to respawn heap
	heap.Push(r, event)

	//slog.Debug("register respanw event","item",item)
}

func (r *Room) Run() {
	slog.Debug("Room starts to run", "Room ID", r.ID)
	for {
		select {
		case cmd := <-r.Commands:
			slog.Debug("room receive command", "command", cmd.Raw)
			r.HandleCommand(cmd)
		case p := <-r.Arrival:
			slog.Debug("New man arrive", "player", p.NickName)
			r.Enter(p)
		// case p := <-r.Departure:
		// 	slog.Debug("user leave", "user", p.NickName)
		// 	r.Leave(p)
		case <-r.Ticker:
			r.OnTick()
		}
	}
}
func (r *Room) HandleCommand(cmd *Command) {
	switch cmd.Verb {
	case "go":
		slog.Debug("go command")
		r.Move(cmd.Player, cmd.Args)
	case "look":
		slog.Debug("look command")
		r.Look(cmd.Player)
	case "get":
		slog.Debug("get command")
		r.Get(cmd.Player, cmd.Args)
	default:
		slog.Debug("default handler")
	}
}

func (r *Room) Get(player *Player, itemName string) {
	r.Mutex.Lock()
	player.Mutex.Lock()
	defer player.Mutex.Unlock()
	defer r.Mutex.Unlock()
	defer slog.Debug("End Get")

	slog.Debug("begin Get")
	item, ok := player.Room.Items[itemName]
	if !ok || item.Count == 0 {
		player.Notify <- &CommandResult{0, fmt.Sprintf("There's no %s in this room", itemName)}
		slog.Debug("return Get")
		return
	}
	item.Count = item.Count - 1
	inv, ok := player.Inventory[itemName]
	if !ok {
		player.Inventory[itemName] = &Inventory{Item: item.Item, Count: 1}
	} else {
		inv.Count++
	}

}

func (r *Room) Move(player *Player, dirction string) {
	flag := true
	for _, et := range r.Exits {
		if et.Direction == dirction {
			// Leave current room
			//r.Departure <- player
			r.Leave(player)
			// Enter new room
			nr, ok := world.RoomMap[et.Room]
			if ok {
				nr.Arrival <- player
				// nr.Enter(player)
				flag = false
				break
			}
		}
	}
	if flag {
		// notify user wrong dirction
		player.Notify <- &CommandResult{0, "wrong directon"}
	}
}

func (r *Room) Look(player *Player) {
	ret := r.Desc
	objects := ""
	for i, v := range r.Items {
		if v.Count > 0 {
			objects += fmt.Sprintf("\r\n%s(%s),Count:%d", v.GetDisplayName(), i, v.Count)
		}
	}
	if objects != "" {
		ret += "\r\n" + "There are some items in this room:" + objects
	}
	ret = ret + "\r\n" + "obvious exits:"
	for _, e := range r.Exits {
		ret = ret + e.Direction
	}

	pc := len(r.Players)
	if pc == 1 {
		ret = ret + "\r\n" + "There is no other player in this room. \r\n"
	} else {
		ret = ret + "\r\n" + fmt.Sprintf("There are %d other players in this room. \r\n", pc-1)
		for i, u := range r.Players {
			if u.NickName == player.NickName {
				continue
			}

			ret += fmt.Sprintf("%d)%s ", i+1, u.NickName)
			if i > 9 {
				ret = ret + "..."
				break
			}
		}
	}

	player.Notify <- &CommandResult{0, ret}
}

func (r *Room) NotifyLeave(player *Player) {
	for _, u := range r.Players {
		if u.NickName == player.NickName {
			continue
		}
		u.Notify <- &CommandResult{0, fmt.Sprintf("user %s left the room ", player.NickName)}
	}
}

func (r *Room) NotifyEntry(player *Player) {
	for _, u := range r.Players {
		if u.NickName == player.NickName {
			continue
		}
		u.Notify <- &CommandResult{0, fmt.Sprintf("user %s entered the room  ", player.NickName)}
	}
}

func (r *Room) Enter(player *Player) {
	player.Room = r
	r.Players = append(r.Players, player)
	r.NotifyEntry(player)
	r.Look(player)
}

func (r *Room) Leave(player *Player) {
	for i, u := range r.Players {
		if u.NickName == player.NickName {
			copy(r.Players[i:], r.Players[i+1:])
			r.Players = r.Players[:len(r.Players)-1]
			break
		}
	}
	player.Room = nil
	r.NotifyLeave(player)
}

func (r *Room) OnTick() {
	r.Mutex.Lock()
	r.Tick++
	tick := r.Tick

	for len(r.respawnHeap) > 0 && r.respawnHeap[0].NextTick <= tick {
		event := heap.Pop(r).(*RespawnEvent)
		slog.Debug("Respawn event triggered", "item", event.ItemName, "tick", tick)
		event.Event.OnRespawn(r)
		event.NextTick = tick + event.Interval
		heap.Push(r, event)
	}

	r.Mutex.Unlock()
	//slog.Debug("room tick","room id",r.ID,"tick",tick)
}
