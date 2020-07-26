package pkg

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync/atomic"
)

// ChatServer holds the chat server application
type ChatServer struct {
	telnetHandler  *telnetHandler
	inShutdown     int32 // accessed atomically (non-zero means we're in Shutdown)
	telnetListener net.Listener
}

// NewChatServer returns an initialized ChatServer
func NewChatServer() *ChatServer {
	return &ChatServer{telnetHandler: newTelnetS(ioutil.Discard)}
}

func (cs *ChatServer) ServeTelnet(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("unable to listen to chat address, error: %s", err))
	}
	defer l.Close()
	log.Printf("telnet chat server started on address: %s", addr)
	cs.telnetListener = l
	for {
		conn, err := l.Accept()
		if err != nil {
			// We are not trying to retry for Accept.
			// Also these error can crop up due to temp err
			// no need to close or log.
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				return
			}
			log.Printf("unable to accept a connection, error: %s\n", err)
			err := conn.Close()
			if err != nil {
				log.Printf("unable to close accept connection, error: %s\n", err)
			}
			return
		}
		// if shutting down close the connection.
		if cs.shuttingDown() {
			err := conn.Close()
			if err != nil {
				log.Printf("unable to close accept connection, error: %s\n", err)
			}
			return
		}
		go cs.telnetHandler.serveConn(conn)
	}
}

// Shutdown tries to gracefully shuts down the chat server
func (cs *ChatServer) Shutdown() {
	atomic.StoreInt32(&cs.inShutdown, 1)
	err := cs.telnetListener.Close()
	if err != nil {
		log.Printf("unable to close listener conn, err: %v", err)
	}
	cs.telnetHandler.chatStore.closeAllConn()
}

func (cs *ChatServer) shuttingDown() bool {
	return atomic.LoadInt32(&cs.inShutdown) != 0
}
