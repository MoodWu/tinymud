package main

import(
	"fmt"
	"log/slog"
	"context"
	//"errors"
	"strings"
)

type Command struct {
	Player *Player
	Verb string
	Args string
	Raw string
}

type CommandResult struct{
	Code int
	Msg string
}

var Alias map[string]string

type CommandFunc func(context context.Context,cmd *Command) (error,CommandResult)


func (c *Command) Parse(value string){
	cmds := strings.Split(value," ")
	c.Verb = cmds[0]
	c.Args = strings.Join(cmds[1:]," ")
	c.Raw = value
}

func (c *Command) GoString() string{
	return "Player:" + c.Player.NickName + ",Command Raw:"+c.Raw
}

func (c *Command) checkCmd(verb string) error {
	if c.Verb != verb{
		return fmt.Errorf("command %s wanted but %s found ",verb,c.Verb)
	}
	return nil
}

func GoFunc (context context.Context, cmd *Command) (error,CommandResult){
	if err:=cmd.checkCmd("go"); err != nil {
		return err,CommandResult{}
	}
	player := cmd.Player
	if player.Room == nil {
		return nil,CommandResult{0,"you can't walk in void space"}
	}

	//send command to room
	room := player.Room
	room.Commands <- cmd
	return nil ,CommandResult{}
}

func LookFunc (context context.Context, cmd *Command) (error,CommandResult){
	if err := cmd.checkCmd("look");err != nil{
		return err,CommandResult{}
	}
	player := cmd.Player
	if player.Room == nil {
		return nil,CommandResult{0,"you can't look in void space"}
	}

	player.Room.Commands <- cmd

	return nil ,CommandResult{}
}

func GetFunc (context context.Context, cmd *Command) (error,CommandResult){
	if err := cmd.checkCmd("get");err != nil{
		return err,CommandResult{}
	}
	player := cmd.Player
	itemName := cmd.Args
	slog.Debug("GetFunc","itemName",itemName)
	player.Room.Mutex.Lock()
	player.Mutex.Lock()
	defer player.Mutex.Unlock()
	defer player.Room.Mutex.Unlock()
	defer slog.Debug("End Get")

	slog.Debug("begin Get")
	item,ok := player.Room.Items[itemName]
	if !ok || item.Count == 0 {
		player.Notify <- &CommandResult{0,fmt.Sprintf("There's no %s in this room",itemName)}
		slog.Debug("return Get")
		return nil,CommandResult{0,fmt.Sprintf("There's no %s in this room",itemName)}
	}
	item.Count = item.Count - 1
	inv,ok := player.Inventory[itemName]
	if !ok{
		player.Inventory[itemName] = &Inventory{Item:item.Item,Count:1}
	} else {
		inv.Count ++
	}

	
	return nil,CommandResult{}
}