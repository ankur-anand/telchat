package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func TestRegisterClient(t *testing.T) {
	t.Parallel()
	ds := newChatDataStore(ioutil.Discard)
	err := ds.registerClient("test-1", nil)
	if err == nil {
		t.Errorf("expected nilConn err got nil")
	}
	server, _ := net.Pipe()
	err = ds.registerClient("test-1", server)
	if err != nil {
		t.Errorf("expected nil err got %v", err)
	}
	err = ds.registerClient("test-1", server)
	if err == nil {
		t.Errorf("expected duplicate client err got nil")
	}
}

func TestBroadcastMsgSubscribeAndUnSubscribeOnMetaRoom(t *testing.T) {
	t.Parallel()
	ds := newChatDataStore(ioutil.Discard)
	var err error
	const roomName = "themetaroom"
	servers := make([]net.Conn, 0, 10)
	clients := make([]net.Conn, 0, 10)
	for i := 0; i < 10; i++ {
		server, client := net.Pipe()
		err = ds.registerClient(fmt.Sprintf("test%d", i), server)
		if err != nil {
			t.Errorf("expected nil err got %v", err)
			continue
		}
		servers = append(servers, server)
		clients = append(clients, client)
	}
	msg := []byte("hi there")
	testClientRead := func(t *testing.T) {
		t.Helper()
		for _, client := range clients {
			err := client.SetReadDeadline(time.Now().Add(time.Millisecond * 50))
			if err != nil {
				t.Errorf("SetReadDeadline failed: %v\n", err)
			}
			b := make([]byte, 8)
			n, err := client.Read(b)
			if n != 8 || err != nil {
				t.Errorf("expected 8 byte to be read got %d or err to be nil got %v", n, err)
			}
		}
	}
	// broadcast to the all the servers on meta room
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	dummyClient := "dummyClient"
	ds.broadcastMsg(ctx, dummyClient, roomName, msg)
	testClientRead(t)

	// Unsubscribe one client.
	ds.unSubscribeClient("test9", roomName)
	// again broadcast to the all the servers on meta room
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ds.broadcastMsg(ctx, dummyClient, roomName, msg)
	for i, client := range clients {
		err := client.SetReadDeadline(time.Now().Add(time.Millisecond * 50))
		if err != nil {
			t.Errorf("SetReadDeadline failed: %v\n", err)
		}
		b := make([]byte, 8)
		n, err := client.Read(b)
		if i == 9 && err == nil {
			t.Errorf("expected err of type read pipe: deadline exceeded got nil")
		}
		if i != 9 && (n != 8 || err != nil) {
			t.Errorf("expected 8 byte to be read got %d or err to be nil got %v", n, err)
		}
	}

	// subscribe the client again.
	ds.subscribeClient("test9", roomName)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ds.broadcastMsg(ctx, dummyClient, roomName, msg)
	testClientRead(t)
}
