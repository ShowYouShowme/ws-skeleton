package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type T1 struct {
	conn *websocket.Conn
	addr string
}

type T2 struct {
	addr   string
	buffer []byte
}

type SessionManager struct {
	c1 chan T1

	c2 chan T2

	c3 chan string

	c4 chan string

	m1 map[string]*websocket.Conn

	// TODO 这里还需要增加一个channel来处理web 后台的 http请求

	// TODO main函数只负责读取数据, 全部业务逻辑都写在这里,就是单线程的,不需要担心多线程数据竞态问题

	// TODO 实现一个专门的协程,来将用户数据写入DB
}

var instance *SessionManager

func GetSessionManager() *SessionManager {
	if instance == nil {
		instance = &SessionManager{
			c1: make(chan T1),
			c2: make(chan T2),
			c3: make(chan string),
			c4: make(chan string),
			m1: map[string]*websocket.Conn{},
		}
	}
	return instance
}

func (receiver *SessionManager) onConnection(conn *websocket.Conn, addr string) {
	receiver.c1 <- T1{conn: conn, addr: addr}
}

func (receiver *SessionManager) onMessage(addr string, buffer []byte) {
	receiver.c2 <- T2{addr: addr, buffer: buffer}
}

func (receiver *SessionManager) onClose(addr string) {
	receiver.c3 <- addr
}

func (receiver *SessionManager) onError(addr string) {
	receiver.c4 <- addr
}

func (receiver *SessionManager) run() {
	for {
		select {
		// 建立链接
		case elem := <-receiver.c1:
			if _, ok := receiver.m1[elem.addr]; ok {
				// 链接已经存在,重复建立链接,这个可能性不大吧
			} else {
				log.Println("新链接到来")
				receiver.m1[elem.addr] = elem.conn
			}
			// 收到消息
		case elem := <-receiver.c2:
			if conn, ok := receiver.m1[elem.addr]; ok {
				message := elem.buffer
				log.Printf("recv: %s", message)
				conn.WriteMessage(websocket.TextMessage, message) // TODO 游戏开发改为BinaryMessage,否则不能发送非utf8字符
			} else {
				panic("invalid connection!")
			}
		case addr := <-receiver.c3:
			if conn, ok := receiver.m1[addr]; ok {
				log.Printf("对方主动关闭链接!")
				conn.Close()
			} else {
				panic("invalid connection!")
			}
		case addr := <-receiver.c4:
			if conn, ok := receiver.m1[addr]; ok {
				log.Printf("链接发生错误!")
				conn.Close()
			} else {
				panic("invalid connection!")
			}
		}
	}

}
