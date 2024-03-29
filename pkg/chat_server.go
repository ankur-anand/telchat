package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"
)

// ChatServer holds the chat server application
type ChatServer struct {
	telnetHandler  *telnetHandler
	inShutdown     int32 // accessed atomically (non-zero means we're in Shutdown)
	telnetListener net.Listener
	messageIO      *messageIO
	restAPIHandler *restAPIHandler
	server         *http.Server
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
	cStore := newChatDataStore(ioutil.Discard)
	return &ChatServer{telnetHandler: newTelnetHFromChatStore(mIo, cStore), messageIO: mIo, restAPIHandler: newRestAPIHandler(mIo, cStore)}, nil
}

// ServeHTTP Serves the Rest HTTP API Call.
func (cs *ChatServer) ServeHTTP(addr string) {
	server := &http.Server{}
	server.Addr = addr
	server.Handler = cs.restAPIHandler
	cs.server = server
	log.Printf("http server starting on address: %s", addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// ServeTelnet responds to the telnet request.
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
			if conn != nil {
				err := conn.Close()
				if err != nil {
					log.Printf("unable to close accept connection, error: %s\n", err)
				}
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

	err = cs.server.Shutdown(context.Background())
	if err != nil {
		log.Printf("err closing http Server, err: %v \n", err)
	}
}

func (cs *ChatServer) shuttingDown() bool {
	return atomic.LoadInt32(&cs.inShutdown) != 0
}
