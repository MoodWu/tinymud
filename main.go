package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	//"errors"
	//"strings"
	"context"
	"sync"
	"time"
)

type ProtocolConn interface {
	ReadLine() (string, error)
	Write(string) error
	WriteLine(string) error // 可选
	Close() error
	ClientType() string // "telnet" / "web" / "app"
	TerminalSize() (width, height int)
}

type World struct {
	Commands chan *Command
	RoomMap  map[string]*Room
	Mutex    sync.RWMutex
}

var defaultRoom *Room
var CommandMap map[string]CommandFunc

// var RoomMap map[string]*Room
var globalTick chan struct{}
var world World

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
	slog.Info("Start MuD Service")

	world.Commands = make(chan *Command, 100)
	world.RoomMap = make(map[string]*Room, 0)

	globalContext, cancel := context.WithCancel(context.Background())

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	globalTick = make(chan struct{})
	//read map
	LoadMaps(globalContext, ".")
	//start all protocol listenging
	go StartTelnetServer(globalContext, ":4001")
	//Start clock tick
	go StartTick(globalContext)

	// 等待信号
	sig := <-sigChan
	cancel() // 取消全局上下文，通知所有 goroutine 停止
	slog.Info(fmt.Sprintf("\n收到信号 %v，开始优雅退出...\n", sig))
}

func StartTelnetServer(ctx context.Context, address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Telnet listen", " error ", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			slog.Debug("Telnet server stops")
			return
		default:

			conn, err := ln.Accept()
			if err != nil {
				slog.Error("telnet accepth ", "error:", err)
				continue
			}
			telnetConn := NewConnection(conn)

			go HandlePlayerInit(ctx, telnetConn)
		}
	}
}

func HandlePlayerInit(ctx context.Context, conn ProtocolConn) {
	//ask for signin/signup
	conn.WriteLine("Welcome to mini MUD")
	conn.WriteLine("1)Sigin")
	conn.WriteLine("2)Sigup")

	str, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}

	slog.Debug("user choice:", "choice", str)

	//read name & passwd
	conn.WriteLine("Please enter username:")
	username, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}

	slog.Debug("user name:", "name", username)
	conn.WriteLine("Please enter password:")
	passwd, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}
	slog.Debug("user passwd:", "passwd", passwd)

	//check or save

	//construct player object
	NewPlayer(ctx, username, defaultRoom, conn)

}

func StartTick(ctx context.Context) {

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Debug("Tick stops")
				return
			case <-ticker.C:
				select {
				case globalTick <- struct{}{}:
				default:
					slog.Debug("global tick channel is full, skip this tick")
				}

			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Debug("Tick handler stops")
				close(globalTick)
				return
			case <-globalTick:
				world.Mutex.RLock()
				//slog.Debug("Tick")
				for _, r := range world.RoomMap {
					r.Ticker <- struct{}{}
				}
				world.Mutex.RUnlock()
			}
		}
	}()
}

func (w *World) Run(ctx context.Context) {
	for {
		select {
		case cmd := <-w.Commands:
			HandleCommand(ctx, cmd)
		case <-ctx.Done():
			slog.Info("Received quit signal, shutting down...")
			return
		}
	}
}

func HandleCommand(ctx context.Context, cmd *Command) {
	slog.Debug("handle command:", "cmd", fmt.Sprintf("%#v", cmd))
	//parse command
	//Check command route
	player := cmd.Player
	worker, ok := CommandMap[cmd.Verb]
	if ok {
		slog.Debug("call worker")
		worker(ctx, cmd)
		slog.Debug("end call worker")
		return
	}
	//the command is not in the list,send directly to room ,to see if it's a script command
	if player.Room != nil {
		slog.Debug("Send to room")
		player.Room.Commands <- cmd
	}
	slog.Debug("Unknown commnad send directly to room")

}
