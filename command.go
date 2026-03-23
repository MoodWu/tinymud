package main

import (
	"fmt"
	"log/slog"

	// "log/slog"
	"context"
	//"errors"
	"strings"
)

type Command struct {
	Player *Player
	Verb   string
	Args   []string
	Raw    string
}

type CommandResult struct {
	Code int
	Msg  string
}

var Alias map[string]string

type CommandFunc func(context context.Context, cmd *Command) (error, CommandResult)

func (c *Command) Parse(value string) {
	cmds := strings.Split(value, " ")
	c.Verb = cmds[0]
	c.Args = cmds[1:]
	c.Raw = value
}

func (c *Command) GoString() string {
	return "Player:" + c.Player.NickName + ",Command Raw:" + c.Raw
}

func (c *Command) checkCmd(verb string) error {
	if c.Verb != verb {
		return fmt.Errorf("command %s wanted but %s found ", verb, c.Verb)
	}
	return nil
}

func GeneralCommandFunc(context context.Context, cmd *Command, wantedCmd string, errString string) (error, CommandResult) {
	if err := cmd.checkCmd(wantedCmd); err != nil {
		return err, CommandResult{}
	}
	player := cmd.Player
	if player.Room == nil {
		return nil, CommandResult{0, errString}
	}

	//send command to room
	room := player.Room
	room.Commands <- cmd
	return nil, CommandResult{}
}

func GoFunc(context context.Context, cmd *Command) (error, CommandResult) {
	return GeneralCommandFunc(context, cmd, "go", "you can't go anywhere in void space")
}

func LookFunc(context context.Context, cmd *Command) (error, CommandResult) {
	return GeneralCommandFunc(context, cmd, "look", "there's nothing to look at in void space")
}

func GetFunc(context context.Context, cmd *Command) (error, CommandResult) {
	return GeneralCommandFunc(context, cmd, "get", "there's nothing to get ")
}

func TalkFunc(ctx context.Context, cmd *Command) (error, CommandResult) {
	slog.Debug("TalkFunc called with command:", "cmd", cmd.Raw)

	if len(cmd.Args) < 2 {
		return nil, CommandResult{
			Msg: "Usage: talk <npc> <message>",
		}
	}

	npcName := cmd.Args[0]
	input := cmd.Args[1]

	n, ok := world.NPCs[npcName]
	if !ok {
		return nil, CommandResult{
			Msg: "No such NPC",
		}
	}

	player := cmd.Player // 假设你有

	// ⭐ 核心：异步AI调用
	go func() {
		reply, err := n.Talk(input)
		if err != nil {
			player.Notify <- &CommandResult{0, "NPC is silent..."}
			return
		}

		player.Notify <- &CommandResult{0, fmt.Sprintf("%s says: %s", npcName, reply)}
	}()

	// 立即返回（避免阻塞）
	return nil, CommandResult{
		Msg: "...",
	}
}
