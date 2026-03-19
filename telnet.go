package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
)

// telnet 命令小结
// 251~254（WILL/WONT/DO/DONT）
//
//	固定再跟 1 字节 option 号，整包 3 字节。
//
// 250（SB）
//
//	后面是不定长 payload，一直读到下一个 IAC SE（240） 才算完；payload 里出现 IAC IAC 表示字面量 0xFF，不做命令解析。
//
// 其余命令（236~249，除 250）
//
//	后面不带任何额外数据，长度就是 2 字节。
const (
	IAC      = 255 // Telnet 命令开始
	WILL     = 251
	WONT     = 252
	DO       = 253
	DONT     = 254
	SB       = 250 // Subnegotiation Begin
	SE       = 240 // Subnegotiation End
	LINEMODE = 34  // 行模式选项
)

//2025-12-08 v0.1 丢弃掉所有的IAC命令

// TelnetConn 实现 protocol.ProtocolConn 接口
type TelnetConn struct {
	raw    net.Conn
	reader *bufio.Reader
	echo   int32 // atomic bool
	width  int32
	height int32
	closed int32
	err    error
	cmd    chan string
}

type ch_CMD chan string

func NewConnection(conn net.Conn) ProtocolConn {

	tc := &TelnetConn{
		raw:    conn,
		reader: bufio.NewReader(conn),
		cmd:    make(chan string, 32),
	}
	//make sure the connection is alive, if it's a tcp connection
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	tc.echo = 1
	slog.Debug("Start a new telnet connection")
	//go tc.negotiate()
	//tc.EnableLineMode()
	go tc.startReading()
	return tc
}

func (tc *TelnetConn) negotiate() {
	slog.Debug("negotiate")
	for {
		b, err := tc.reader.Peek(1) // 只偷看，不消耗！
		if err != nil {
			return
		}
		if b[0] == IAC {
			tc.reader.ReadByte()
			tc.handleIAC()
		}
	}
}

func (tc *TelnetConn) handleIAC() {
	slog.Debug("Start handle IAC")
	// 简单版：只处理最常用的几个
	cmd := make([]byte, 3)
	if _, err := io.ReadFull(tc.raw, cmd[:2]); err != nil {
		slog.Error("handleIAC ", "readfull err,", err)
		return
	}
	slog.Debug("handleIAC", "%x", cmd)
	if cmd[0] == SB && cmd[1] == 31 { // NAWS
		size := make([]byte, 4)
		if _, err := io.ReadFull(tc.raw, size); err == nil && size[3] == SE {
			w := int(size[0])<<8 | int(size[1])
			h := int(size[2])<<8 | int(size[3])
			atomic.StoreInt32(&tc.width, int32(w))
			atomic.StoreInt32(&tc.height, int32(h))
		}
	}
	// 其他 WILL/DO 我们可以选择性响应
}

// 读取网络传送来的命令，分开用户命令和IAC命令
func (tc *TelnetConn) startReading() {
	for {
		//从连接中读取信息，直到收到回车
		line, err := tc.reader.ReadString('\n')
		tc.err = err
		if err != nil {
			if err == io.EOF {
				slog.Debug("client shutdown")
				return
			}
			slog.Debug("read network error:", "error", err)
		}
		//看看当前是否已经收到了IAC命令，如果是，命令的开头
		userCmd := filterCommand(line)
		tc.cmd <- userCmd
	}
}

func filterCommand(cmd string) string {
	//去掉末尾的回车换行
	cmd = strings.TrimRight(cmd, "\r\n")
	c := []byte(cmd)
	ret := make([]byte, 0)
	for i := 0; i < len(c); {
		if c[i] == IAC {
			//判断是否是字面量255
			if i+1 < len(c) && c[i+1] == IAC {
				ret = append(ret, IAC)
				i = i + 2
				continue
			}
			//开始处理IAC命令
			if i+1 < len(c) {
				switch c[i+1] {
				case WILL, WONT, DO, DONT:
					i = i + 3
				case SB:
					for i < len(c) {
						i++
						if c[i] == SE {
							break
						}
					}
				}
			}
			continue
		}
		ret = append(ret, c[i])
		i++
	}
	return string(ret)
}

// 实现接口：读取一行（去掉 \r\n）
func (tc *TelnetConn) ReadLine() (string, error) {
	cmd := <-tc.cmd
	return cmd, tc.err
	/*
		line, err := tc.reader.ReadString('\n')
		slog.Debug("readline"," got",line)
		if err != nil {
			return "", err
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		return line, nil
	*/
}

// 实现接口：发送字符串
func (tc *TelnetConn) Write(msg string) error {
	if atomic.LoadInt32(&tc.closed) == 1 {
		return io.EOF
	}
	_, err := tc.raw.Write([]byte(msg))
	return err
}

func (tc *TelnetConn) WriteLine(msg string) error {
	return tc.Write(msg + "\r\n")
}

// 实现接口：关闭
func (tc *TelnetConn) Close() error {
	atomic.StoreInt32(&tc.closed, 1)
	return tc.raw.Close()
}

// 实现接口：客户端类型
func (tc *TelnetConn) ClientType() string {
	return "telnet"
}

// 实现接口：终端大小
func (tc *TelnetConn) TerminalSize() (int, int) {
	return int(atomic.LoadInt32(&tc.width)), int(atomic.LoadInt32(&tc.height))
}

// 密码时关闭回显
func (tc *TelnetConn) DisableEcho() {
	atomic.StoreInt32(&tc.echo, 0)
	tc.raw.Write([]byte{IAC, WILL, 1}) // 告诉客户端：我来回显（实际我们不回）
}

// 恢复回显
func (tc *TelnetConn) EnableEcho() {
	atomic.StoreInt32(&tc.echo, 1)
	tc.raw.Write([]byte{IAC, WONT, 1}) // 让客户端自己回显
}

// 是否回显（业务层判断密码时用）
func (tc *TelnetConn) IsEcho() bool {
	return atomic.LoadInt32(&tc.echo) == 1
}

func (tc *TelnetConn) EnableLineMode() {
	// IAC DO LINEMODE
	tc.raw.Write([]byte{IAC, DO, LINEMODE})

	// 同时请求回显由客户端处理（减少服务器负担）
	// IAC WILL ECHO
	//tc.raw.Write([]byte{IAC, WILL, 1}) // 1 = ECHO
}

func (tc *TelnetConn) DisableLineMode() {

}
