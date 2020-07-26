package pkg

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

// messageIO logs the messages to local log file.
type messageIO struct {
	lock      sync.Mutex // support concurrent read and write.
	mBuffer   chan []byte
	syCh      chan struct{}
	file      *os.File
	readFiled *os.File // to support concurrent read and write op's to the same underlying file.
}

// Write to the message buffer.
func (m *messageIO) Write(p []byte) (n int, err error) {
	m.mBuffer <- p
	return len(p), nil
}

// ReadAll Content of a file.
func (m *messageIO) ReadAll() ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	// flush write
	// err := m.file.Sync()
	// if err != nil {
	// 	return nil, err
	// }
	_, err := m.readFiled.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	fileinfo, err := m.readFiled.Stat()
	if err != nil {
		return nil, err
	}
	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	_, err = m.readFiled.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buffer, nil
}

// Close the underlying file
func (m *messageIO) Close() error {
	err := m.readFiled.Close()
	err = m.file.Close()
	return err
}

func (m *messageIO) Sync() error {
	m.syCh <- struct{}{}
	return nil
}

func newMessageIO(file *os.File, readFile *os.File) *messageIO {
	mio := &messageIO{
		file:      file,
		mBuffer:   make(chan []byte, 100),
		syCh:      make(chan struct{}),
		readFiled: readFile,
	}
	go batchWriteMessage(mio)
	return mio
}

// Write the Message in Batch
func batchWriteMessage(m *messageIO) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	ticker := time.NewTicker(time.Millisecond * 200)
	var err error
	for {
		select {
		case <-ticker.C:
			if len(buffer.Bytes()) > 0 {
				err = m.fileWrite(buffer.Bytes())
				if err != nil {
					panic(err)
				}
				buffer.Reset()
			}

		case record := <-m.mBuffer:
			buffer.Write(record)
			if len(buffer.Bytes()) >= 1024 {
				err = m.fileWrite(buffer.Bytes())
				if err != nil {
					panic(err)
				}
				buffer.Reset()
			}
		case <-m.syCh:
			if len(buffer.Bytes()) > 0 {
				err = m.fileWrite(buffer.Bytes())
				if err != nil {
					panic(err)
				}
				buffer.Reset()
			}
			break
		}
	}
}

func (m *messageIO) fileWrite(msg []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, err := m.file.Write(msg)
	return err
}
