package main

import (
	"fmt"
	//"log/slog"
)

// 最小的公共接口（房间真正关心的）
type Item interface {
	GetID() string
	GetName() string
	GetDisplayName() string
	GetDescription() string
	GetWeight() int // 重量、价值等通用属性

	CanGet() bool // 是否可以拿起来
	CanDrop() bool

	OnGet(p *Player) error
	OnDrop(p *Player) error
	// 可选：Save() map[string]any, Load(data map[string]any)
}

type Food interface {
	Kind() string
	RespawnCount() int
	RespawnTick() int
	RespawnMax() int
	Nutrition() int
}

type IEat interface {
	OnEat(player *Player)
}

type IGet interface {
	OnGet(player *Player)
}

type IDrop interface {
	OnDrop(player *Player)
}

type IRespawn interface {
	OnRespawn(r *Room)
	GetRespawnID() string
}

type BaseItem struct {
	Kind        string
	Name        string
	DisplayName string
	Desc        string
}

func (i *BaseItem) GetID() string {
	return ""
}

func (i *BaseItem) GetName() string {
	return i.Name
}

func (i *BaseItem) GetDisplayName() string {
	return i.DisplayName
}

func (i *BaseItem) GetDescription() string {
	return i.Desc
}

func (i *BaseItem) GetWeight() int {
	return 0
}

func (i *BaseItem) OnGet(player *Player) error {
	//player.Room.mutex.Lock()
	//palyer.Room.Itmes[i.]
	//player.

	return nil
}
func (i *BaseItem) OnDrop(player *Player) error {
	return nil
}

type BaseFood struct {
	BaseItem
	RespawnTick  int
	RespawnMax   int
	RespawnCount int
	Nutrition    int
}

func (f *BaseFood) CanGet() bool {
	return true
}
func (f *BaseFood) CanDrop() bool {
	return true
}

func (f *BaseFood) OnEat(player *Player) {
	hunger := player.Hunger - f.Nutrition

	if hunger < 0 {
		hunger = 0
	}
	player.Mutex.Lock()
	defer player.Mutex.Unlock()

	player.Hunger = hunger
	if player.Hunger < 10 {
		player.Notify <- &CommandResult{0, "You're full,you can not eat anymore"}
	} else {
		player.Notify <- &CommandResult{0, fmt.Sprintf("You eat some %s and feel much better. ", f.Name)}
	}
}

func (f *BaseFood) OnRespawn(r *Room) {

	//slog.Debug("OnRespawn","food",f,"cuurent", r.Items[f.Name].Count)

	max := f.RespawnMax
	if max == 0 {
		max = r.Items[f.Name].Count
	}
	step := f.RespawnCount
	r.Items[f.Name].Count += step
	// slog.Debug("Respawn","food",f.Name,"count",r.Items[f.Name].Count,"step",step)
	if r.Items[f.Name].Count > max {
		r.Items[f.Name].Count = max
	}
	//slog.Debug("OnRespawn End","food",f,"cuurent", r.Items[f.Name].Count)
}

func (f *BaseFood) GetRespawnID() string {
	return f.Name
}
