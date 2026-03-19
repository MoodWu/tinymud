package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	//"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

type RoomConfig struct {
	ID     string             `yaml:"ID"`
	Name   string             `yaml:"Name"`
	Length int                `yaml:"Length"`
	Width  int                `yaml:"Width"`
	Desc   string             `yaml:"Desc"`
	Items  map[string]RawItem `yaml:"Items"`
	Exits  []*RawExit         `yaml:"Exits"`
}

type RawItem struct {
	Kind         string `yaml:"Kind"`
	Name         string `yaml:"Name"`
	DisplayName  string `yaml:"DisplayName"`
	Desc         string `yaml:"Desc"`
	RespawnTick  int    `yaml:"RespawnTick"`
	RespawnCount int    `yaml:"RespawnCount"`
	RespawnMax   int    `yaml:"RespawnMax"`
	Nutrition    int    `yaml:"Nutrition"`
	Count        int    `yaml:"Count"`
}

type RawExit struct {
	Direction string `yaml:"Direction"`
	Room      string `yaml:"Room"`
}

func LoadMaps(ctx context.Context, dir string) {

	files := LoadDir(dir)
	for _, afile := range files {
		LoadMap(ctx, afile)
	}
	world.DefaultRoom = world.RoomMap["1"]
}
func LoadDir(dir string) []string {

	files := make([]string, 0)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Debug(fmt.Sprintf("访问 %s 时出错: %v\n", path, err))
			return err
		}

		// 检查是否是文件且扩展名匹配
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".yaml") {
			//slog.Debug(fmt.Sprintln("找到:", path))
			files = append(files, path)
		}

		return nil
	})
	return files
}

func LoadMap(ctx context.Context, fileName string) {
	//slog.Debug("LoadMap","filename",fileName)
	//read file
	data, err := os.ReadFile(fileName)
	if err != nil {
		slog.Debug("Load room read config err", "file", fileName, "err", err)
		panic(err)
	}

	//slog.Debug("data length","len",len(data))
	var rc RoomConfig

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	//decoder.KnownFields(true)

	if err := decoder.Decode(&rc); err != nil {
		slog.Debug("Load room decode config err", "file", fileName, "err", err)
		panic(err)
	}

	slog.Debug("read room config", "room config", rc)
	/*
		viper.SetConfigFile(fileName)
		if err := viper.ReadInConfig();err!=nil {
			slog.Debug("read yaml error","err",err)
			panic(err)
		}
		roomID := viper.GetString("ID") */
	room := Room{
		ID:          rc.ID,
		Name:        rc.Name,
		Length:      rc.Length,
		Width:       rc.Width,
		Desc:        rc.Desc,
		Exits:       make([]*Exit, 0),
		Departure:   make(chan *Player),
		Arrival:     make(chan *Player),
		Commands:    make(chan *Command, 10),
		Ticker:      make(chan struct{}),
		Tick:        1,
		Items:       make(map[string]*Inventory),
		respawnHeap: make([]*RespawnEvent, 0),
	}

	for _, v := range rc.Exits {
		room.Exits = append(room.Exits, &Exit{Direction: v.Direction, Room: v.Room})
	}
	for k, v := range rc.Items {
		//slog.Debug("RawItem","Key",k)
		switch v.Kind {
		case "Food":
			food := BaseFood{
				BaseItem: BaseItem{
					Kind:        v.Kind,
					Name:        v.Name,
					DisplayName: v.DisplayName,
					Desc:        v.Desc,
				},
				RespawnTick:  v.RespawnTick,
				RespawnCount: v.RespawnCount,
				RespawnMax:   v.RespawnMax,
				Nutrition:    v.Nutrition,
			}
			room.Items[k] = &Inventory{&food, v.Count}
			room.RegisterRespawnEvent(v.RespawnTick, k, &food)
		default:
			slog.Debug("Unkonwn type item", "Item", v)
		}
	}

	world.RoomMap[rc.ID] = &room
	go room.Run(ctx)
}
