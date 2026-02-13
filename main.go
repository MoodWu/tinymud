package main

import(
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"net"
	"io"
	//"errors"
	//"strings"
	//"context"
	"sync"
	"time"
)


type ProtocolConn interface {
    ReadLine() (string, error)
    Write(string) error
    WriteLine(string) error // 可选
    Close() error
    ClientType() string    // "telnet" / "web" / "app"
    TerminalSize() (width, height int)
}

var defaultRoom *Room
var CommandMap map[string]CommandFunc
var Rooms []*Room
var RoomMap map[string]*Room
var globalTick chan struct{}
var mutex sync.RWMutex

func main(){
    opts := &slog.HandlerOptions{Level: slog.LevelDebug}
    slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
	slog.Info("Start MuD Service")

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	globalTick = make(chan struct{})
	//read map
	LoadMaps(".")
	//start all protocol listenging
	go StartTelnetServer(":4001")	
	//Start clock tick
	go StartTick()

	// 等待信号
	sig := <-sigChan
	slog.Info(fmt.Sprintf("\n收到信号 %v，开始优雅退出...\n", sig))
}

func StartTelnetServer(address string){
	ln,err := net.Listen("tcp",address)
	if err != nil {
		slog.Error("Telnet listen"," error ",err)
		return 
	}
  for {
		conn,err := ln.Accept()
		if err != nil {
			slog.Error("telnet accepth ","error:",err)
			continue
		}
		telnetConn := NewConnection(conn)
		
		go HandlePlayerInit(telnetConn)
	}
}


func HandlePlayerInit(conn ProtocolConn) {
	//ask for signin/signup
	conn.WriteLine("Welcome to mini MUD")
	conn.WriteLine("1)Sigin")
	conn.WriteLine("2)Sigup")

	str, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}

	slog.Debug("user choice:","choice",str)

		//read name & passwd
	conn.WriteLine("Please enter username:")
	username, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}

	slog.Debug("user name:","name",username)
	conn.WriteLine("Please enter password:")
	passwd, err := conn.ReadLine()
	if err == io.EOF {
		slog.Debug("player exit")
		return
	}
	slog.Debug("user passwd:","passwd",passwd)


	//check or save

	
	//construct player object
	NewPlayer(username,defaultRoom,conn)

}

func StartTick() {

	go func(){
		ticker := time.NewTicker(200 * time.Millisecond)
    	defer ticker.Stop()
		for range ticker.C {
        	globalTick <- struct{}{}
    	}
	}()

	go func(){
		for range globalTick{
			mutex.RLock()
			//slog.Debug("Tick")
			for _,r := range RoomMap{
				r.Ticker <- struct{}{}
			}
			mutex.RUnlock()
		}
	}()


}
