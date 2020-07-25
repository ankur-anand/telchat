package pkg

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// metaRoom is General chat room
	metaRoom = "default"
)

var (
	errDuplicateClient = errors.New("duplicate client")
	errNilConn         = errors.New("nil connection")
	noTimeout          = time.Time{}
)

type (
	// clientID represents name for each connected client
	clientID string

	// roomID is a named typed for each room present in the chat Server
	roomID string
)

// client is each unique client that is connected to the chatServer
type client struct {
	conn net.Conn
	// ignoreList contains all the list of client that a client has decided to ignore
	ignoreList map[clientID]struct{}
}

type (
	subscriber map[clientID]net.Conn

	chatDataStore struct {
		logWriter io.Writer
		lock      sync.RWMutex
		// clients store all clients currently connected
		// to the telnetServer
		clients map[clientID]*client
		// roomsSubscribers store all the client subscriber to particular room.
		roomsSubscribers map[roomID]subscriber
	}
)

func newChatDataStore(lw io.Writer) *chatDataStore {
	cds := chatDataStore{
		logWriter:        lw,
		clients:          make(map[clientID]*client),
		roomsSubscribers: make(map[roomID]subscriber),
	}
	cds.roomsSubscribers[metaRoom] = make(subscriber)
	return &cds
}

// isDuplicateClient returns true if clientName is already registered
func (cds *chatDataStore) isDuplicateClient(clientName string) bool {
	_, ok := cds.clients[clientID(clientName)]
	return ok
}

// registerClient registers the given client to the chat data store.
// all the registered client will be by default part of the meta room.
func (cds *chatDataStore) registerClient(clientName string, conn net.Conn) error {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	if cds.isDuplicateClient(clientName) {
		return errDuplicateClient
	}

	if conn == nil {
		return errNilConn
	}
	cid := clientID(clientName)
	client := &client{
		conn:       conn,
		ignoreList: make(map[clientID]struct{}),
	}
	cds.clients[cid] = client
	cds.roomsSubscribers[metaRoom][cid] = conn
	return nil
}

// addClientToRoom register the client to the given room in the chat
// data store
func (cds *chatDataStore) addClientToRoom(clientName, roomName string) {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	cid := clientID(clientName)
	roomId := roomID(roomName)
	client, ok := cds.clients[cid]
	// add the client to room store
	_, ok = cds.roomsSubscribers[roomId]
	if !ok {
		cds.roomsSubscribers[roomId] = make(subscriber)
	}
	cds.roomsSubscribers[roomId][cid] = client.conn
}

// removeClientFromRoom deregister the client from the given room in the chat
// data store
func (cds *chatDataStore) removeClientFromRoom(clientName, roomName string) {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	cid := clientID(clientName)
	roomId := roomID(roomName)
	// delete the client from room store
	roomM := cds.roomsSubscribers[roomId]
	delete(roomM, cid)
}

// deleteClient from the client data store as well as from
// all of the room.
func (cds *chatDataStore) deleteClient(clientName string) {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	cid := clientID(clientName)
	delete(cds.clients, cid)
}

// ignoreNamedClient add the proposed client in the current client ignore list
func (cds *chatDataStore) ignoreNamedClient(myName, clientName string) {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	mid := clientID(myName)
	cid := clientID(clientName)
	cds.clients[mid].ignoreList[cid] = struct{}{}
}

// allowNamedClient removes the proposed client from the current client ignore list
func (cds *chatDataStore) allowNamedClient(myName, clientName string) {
	cds.lock.Lock()
	defer cds.lock.Unlock()
	mid := clientID(myName)
	cid := clientID(clientName)
	delete(cds.clients[mid].ignoreList, cid)
}

// broadcast the given message to the room that the client is currently
// part of.
func (cds *chatDataStore) broadcastMsg(ctx context.Context, clientName, roomName string, msg []byte) {
	_, err := cds.logWriter.Write(msg)
	if err != nil {
		log.Println("err logging message", err)
	}
	cds.lock.RLock()
	defer cds.lock.RUnlock()
	cid := clientID(clientName)
	roomId := roomID(roomName)
	roomM := cds.roomsSubscribers[roomId]
	// IMP: Note
	// this sends the message one by one over an unsupervised goroutine.
	// No of goroutine spawned is not accounted, and also msg is dropped
	// if there is timeout while writing or an error while writing to the socket.
	for keyCID, conn := range roomM {
		// if keyCID is equal to cid don't relay the msg
		if keyCID == cid {
			continue
		}
		// if cid is in the ignore list don't broadcast
		_, ok := cds.clients[keyCID].ignoreList[cid]
		if ok {
			continue
		}

		go cds.sendMsg(ctx, conn, msg)
	}
}

// sendMsg Sends the given MSG to the client
func (cds *chatDataStore) sendMsg(ctx context.Context, conn net.Conn, msg []byte) {
	cds.lock.RLock()
	defer cds.lock.RUnlock()
	err := conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		log.Printf("SetWriteDeadline failed: %v\n", err)
	}
	defer func() {
		// reuse write conn.
		err := conn.SetWriteDeadline(noTimeout)
		if err != nil {
			log.Printf("SetWriteDeadline failed: %v\n", err)
		}
	}()
	select {
	case <-ctx.Done():
		return
	default:
		// these writes are not buffered
		_, err := conn.Write(msg)
		if err != nil {
			log.Printf("failed sending a message to network: %v\n", err)
		}
	}
}

// closeAllConn closes all active conn in the memory store.
func (cds *chatDataStore) closeAllConn() {
	cds.lock.RLock()
	defer cds.lock.RUnlock()
	for _, v := range cds.clients {
		err := v.conn.Close()
		if err != nil {
			log.Printf("error closing conn, err: %v", err)
		}
	}
}
