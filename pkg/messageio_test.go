package pkg

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMessageIO(t *testing.T) {
	file, err := ioutil.TempFile("", "telchat.*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	readfile, err := os.OpenFile(file.Name(), os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	msgs := [][]byte{[]byte("hi message 1"), []byte("hi message 2"), []byte("hi message 3")}
	mio := newMessageIO(file, readfile)
	defer mio.Close()
	for _, m := range msgs {
		_, err = mio.Write(m)
		if err != nil {
			t.Errorf("err writing to mio failed %v", err)
		}
	}

	err = mio.Sync()
	if err != nil {
		t.Errorf("err sync call, err: %v", err)
	}
	time.Sleep(500 * time.Millisecond) // some io breather

	go func() {
		msgs := [][]byte{[]byte("hi message new 1"), []byte("hi message new 2"), []byte("hi message new 3")}
		for _, m := range msgs {
			_, err := mio.Write(m)
			if err != nil {
				t.Errorf("err writing to mio failed %v", err)
			}
		}
	}()
	buff, err := mio.ReadAll()
	if err != nil {
		t.Errorf("err reading file content, err: %v", err)
	}
	for _, m := range msgs {
		if !bytes.Contains(buff, m) {
			t.Errorf("expected msg: %s not found in received msg", m)
		}
	}
	time.Sleep(500 * time.Millisecond) // some io breather
	buff, err = mio.ReadAll()
	if err != nil {
		t.Errorf("err reading file content, err: %v", err)
	}
	for _, m := range msgs {
		if !bytes.Contains(buff, m) {
			t.Errorf("expected msg: %s not found in received msg", m)
		}
	}
	msgs = [][]byte{[]byte("hi message new 1"), []byte("hi message new 2"), []byte("hi message new 3")}
	for _, m := range msgs {
		if !bytes.Contains(buff, m) {
			t.Errorf("expected msg: %s not found in received msg", m)
		}
	}
}
