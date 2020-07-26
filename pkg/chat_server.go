package pkg

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
)

// ChatServer holds the chat server application
type ChatServer struct {
	telnetHandler  *telnetHandler
	inShutdown     int32 // accessed atomically (non-zero means we're in Shutdown)
	telnetListener net.Listener
	messageIO      *messageIO
}

// NewChatServer returns an initialized ChatServer
func NewChatServer(filePath string) (*ChatServer, error) {
	fd, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	readfd, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	mIo := newMessageIO(fd, readfd)
	return &ChatServer{telnetHandler: newTelnetS(mIo), messageIO: mIo}, nil
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
		log.Printf("unable to close listener conn, err: %v \n", err)
	}
	cs.telnetHandler.chatStore.closeAllConn()
	err = cs.messageIO.Sync()
	if err != nil {
		log.Printf("err sync message logs, err: %v \n", err)
	}
	err = cs.messageIO.Close()
	if err != nil {
		log.Printf("err closing fd logs, err: %v \n", err)
	}
}

func (cs *ChatServer) shuttingDown() bool {
	return atomic.LoadInt32(&cs.inShutdown) != 0
}
