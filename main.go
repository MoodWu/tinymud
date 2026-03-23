package main

import (
	"flag"
	"fmt"
	"game/ai"
	"game/npc"
	"io"
	"log"
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
	Commands    chan *Command
	RoomMap     map[string]*Room
	Mutex       sync.RWMutex
	DefaultRoom *Room
	CommandMap  map[string]CommandFunc
	GlobalTick  chan struct{}
	NPCs        map[string]*npc.NPC
}

// var defaultRoom *Room
// var CommandMap map[string]CommandFunc

// // var RoomMap map[string]*Room
// var globalTick chan struct{}
var world *World

func NewWorld(aiClient *ai.Client) *World {
	w := &World{
		Commands:   make(chan *Command, 100),
		RoomMap:    make(map[string]*Room, 0),
		CommandMap: make(map[string]CommandFunc, 0),
		GlobalTick: make(chan struct{}, 100),
	}

	merchant := &npc.NPC{
		Name:        "merchant",
		Personality: "a greedy medieval merchant who loves gold",
		Client:      aiClient,
		Memory:      make(map[string]*npc.Memory),
	}

	w.NPCs = map[string]*npc.NPC{
		"merchant": merchant,
	}
	return w
}
func main() {
	apiKeyFlag := flag.String("api-key", "", "LLM API Key")
	flag.Parse()

	// 2️⃣ 环境变量兜底
	apiKey := *apiKeyFlag
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}

	if apiKey == "" {
		log.Fatal("API key is required (use --api-key or set LLM_API_KEY)")
	}

	aiClient := &ai.Client{
		APIKey: apiKey,
		URL:    "https://api.deepseek.com/chat/completions",
		Model:  "deepseek-chat",
	}
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
	slog.Info("Start MuD Service")

	world = NewWorld(aiClient)

	globalContext, cancel := context.WithCancel(context.Background())

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//read map
	LoadMaps(globalContext, ".")
	RegisterCommands()
	//start all protocol listenging
	go StartTelnetServer(globalContext, ":4001")
	//Start clock tick
	go StartTick(globalContext)

	go world.Run(globalContext)

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
	NewPlayer(ctx, username, world.DefaultRoom, conn)

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
				case world.GlobalTick <- struct{}{}:
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
				close(world.GlobalTick)
				return
			case <-world.GlobalTick:
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
			slog.Debug("world receive command", "command", cmd.Raw)
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
	worker, ok := world.CommandMap[cmd.Verb]
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
