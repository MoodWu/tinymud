# TinyMUD - 多用户地下城游戏引擎

一个用Go编写的轻量级但功能完整的MUD（Multi-User Dungeon）游戏引擎，支持多客户端连接、AI-驱动NPC、实时性能监控。项目在一个最小可运行的MUD系统的基础上，探索更多的技术可能性。

## 🎮 主要特性

- **多用户支持**：通过Telnet、Web、App等多种协议连接
- **完整的游戏系统**：房间、玩家、物品、NPC管理
- **命令系统**：灵活的命令处理框架
- **AI NPC**：使用AI驱动的智能NPC角色
- **地图系统**：基于YAML的世界配置
- **性能监控**：集成Prometheus指标收集和Grafana可视化
- **可配置**：通过YAML文件灵活定义游戏世界

## 📁 项目结构

```
tinymud/
├── main.go              # 应用入口
├── player.go            # 玩家系统
├── room.go              # 房间系统
├── item.go              # 物品系统
├── command.go           # 命令处理
├── map.go               # 地图管理
├── registry.go          # 注册表
├── telnet.go            # Telnet协议支持
├── npc/                 # NPC系统
│   └── npc.go
├── ai/                  # AI引擎
│   ├── client.go
│   ├── service.go
│   └── prompt.go
├── metrics/             # 性能指标
│   └── metrics.go
├── tools/               # 开发工具
│   └── codegen.go
├── moniter/             # 监控配置
│   ├── docker-compose.yml
│   ├── prometheus.yml
│   └── grafana-data/
├── default.yaml         # 默认世界配置
├── a2.yaml              # 额外世界配置
└── item.yml             # 物品定义
```

## 🚀 快速开始

### 前置要求

- Go 1.25.3 或更高版本
- Docker & Docker Compose（用于监控）

### 构建项目

```bash
cd tinymud
go build -o tinymud .
```

### 运行游戏服务器

```bash
./tinymud
```

服务器将在默认端口监听Telnet连接。

### 启用监控

在 `moniter/` 目录下运行：

```bash
docker-compose up -d
```

- **Prometheus**: http://localhost:9090 - 指标收集
- **Grafana**: http://localhost:3000 - 可视化仪表板

## 🎯 核心模块

### 玩家系统 (player.go)
管理玩家状态、库存、位置和属性。

### 房间系统 (room.go)
定义游戏世界的房间、出口和环境。

### 物品系统 (item.go)
物品定义、库存管理和物品响生机制。

### NPC系统 (npc/)
创建可交互的非玩家角色。

### AI引擎 (ai/)
为NPC提供智能行为和对话能力。

### 命令系统 (command.go)
实现玩家命令的处理和执行。

## ⚙️ 配置

### 世界配置示例 (default.yaml)

```yaml
ID: 1
Name: default room
Length: 3
Width: 3
Desc: an empty room
Items:
  Apple:
    Kind: Food
    DisplayName: 苹果
    Nutrition: 20
Exits:
  - Direction: north
    Room: 2
```

## 📊 监控指标

项目集成了Prometheus进行性能监控：

- 玩家连接数
- 命令执行统计
- NPC活动
- 内存使用
- 响应时间

## 🔧 支持的协议

- **Telnet**: 传统MUD客户端连接
- **Web**: 网页浏览器支持
- **App**: 原生应用连接

## 📦 依赖

- `gopkg.in/yaml.v3` - YAML配置解析
- `github.com/prometheus/client_golang` - Prometheus指标

## 🛠️ 开发

### 添加新命令

在 `command.go` 中定义命令处理函数，在CommandMap中注册。

### 扩展NPC

在 `npc/` 目录中实现NPC行为逻辑。



## 📞 联系

如有问题或建议，请通过项目Issues提出。

---

**开始你的MUD冒险之旅吧！** 🎲✨
