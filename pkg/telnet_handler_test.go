package pkg

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	update   = flag.Bool("update", false, "update the golden files of this test")
	printLog = flag.Bool("log", false, "print all logs from this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !*printLog {
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}

func TestDisHelpCommand(t *testing.T) {
	t.Parallel()
	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Log("update golden file")
		if err := ioutil.WriteFile(gp, []byte(disHelpCommand()), 0644); err != nil {
			t.Fatalf("failed to update golden file: %s", err)
		}
	}
	g, err := ioutil.ReadFile(gp)
	if err != nil {
		t.Fatalf("failed reading .golden: %s", err)
	}
	t.Log(disHelpCommand())
	if !bytes.Equal([]byte(disHelpCommand()), g) {
		t.Errorf("written in disHelpCommand does not match .golden file")
	}
}

func TestFormatDM(t *testing.T) {
	t.Parallel()
	out := formatDM("Ankur", "default", "hi there")
	t.Log(formatDM("Ankur", "default", "hi there"))
	// dates part always chage so we match only sub slice
	subSlices := []struct {
		name string
	}{
		{
			name: "Ankur",
		},
		{
			"@",
		},
		{
			"default",
		},
		{
			":",
		},
		{
			name: "hi there",
		},
	}
	for _, ss := range subSlices {
		t.Run(ss.name, func(t *testing.T) {
			if !bytes.Contains([]byte(out), []byte(ss.name)) {
				t.Errorf("subslice %s not found", ss.name)
			}
		})
	}
}

func TestFormatCMDErr(t *testing.T) {
	t.Parallel()
	out := formatCMDErr("hi error")
	t.Log(formatCMDErr("hi error"))
	subSlices := []struct {
		name string
	}{
		{
			name: "[Error]",
		},
		{
			":",
		},
		{
			"invalid command",
		},
		{
			"`hi error`",
		},
	}
	for _, ss := range subSlices {
		t.Run(ss.name, func(t *testing.T) {
			if !bytes.Contains([]byte(out), []byte(ss.name)) {
				t.Errorf("subslice %s not found", ss.name)
			}
		})
	}
}

func TestServeConn(t *testing.T) {
	t.Parallel()
	ts := newTelnetS(ioutil.Discard)
	sc, cc := net.Pipe()
	go ts.serveConn(sc)

	b := make([]byte, 4096)
	_, err := cc.Read(b)
	must(t, err)
	// write the name to the chat server
	err = cc.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))
	must(t, err)
	_, err = cc.Write([]byte("ankur\n\r"))
	must(t, err)
	b = make([]byte, 4096)
	_, err = cc.Read(b)
	must(t, err)
	b = make([]byte, 128)
	n, err := cc.Read(b)
	must(t, err)
	expected := infoDisplay("ankur", metaRoom)
	if !bytes.Equal(b[:n], []byte(expected)) {
		t.Log(expected)
	}
}

func TestServeConnDrop(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("recover func should have been called")
		}
	}()
	// this should not panic
	ts := newTelnetS(ioutil.Discard)
	sc, cc := net.Pipe()
	go ts.serveConn(sc)

	b := make([]byte, 4096)
	_, err := cc.Read(b)
	must(t, err)
	// write the name to the chat server
	err = cc.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))
	must(t, err)
	time.Sleep(time.Millisecond * 10)
	err = cc.Close()
	must(t, err)
}

func TestIgnoreAllowClientServeConn(t *testing.T) {
	t.Parallel()
	ts := newTelnetS(ioutil.Discard)
	sc1, cc1 := net.Pipe()
	go ts.serveConn(sc1)
	initialRead(t, cc1, []byte("ankur\n\r"))

	// start two new client.
	sc2, cc2 := net.Pipe()
	go ts.serveConn(sc2)
	initialRead(t, cc2, []byte("anand\n\r"))

	sc3, cc3 := net.Pipe()
	go ts.serveConn(sc3)
	initialRead(t, cc3, []byte("ankuranand\n\r"))

	// ankuranand to ignore msg from ankur
	msg := []byte("/client ignore ankur\n\r")
	writeMsg(t, cc3, msg)
	done := make(chan bool)
	ts.hook = func() {
		done <- true
	}

	select {
	case <-time.After(time.Second * 2):
		t.Error("timeout waiting for hook call back")
	case <-done:
	}

	// ankur client send msg
	msg = []byte("hello everyone\n\r")
	writeMsg(t, cc1, msg)

	// anand client read msg
	readM := make([]byte, 512)
	err := readMsg(t, cc2, readM)
	must(t, err)
	if !bytes.Contains(readM, []byte("hello everyone")) {
		t.Errorf("expected msg: %s not found in received msg", "hello everyone")
	}

	// ankuranand client read msg
	readM = make([]byte, 512)
	err = readMsg(t, cc3, readM)
	if err == nil {
		t.Error("expected read deadline error got nil")
	}

	// allow the client back.
	msg = []byte("/client allow ankur\n\r")
	writeMsg(t, cc3, msg)

	select {
	case <-time.After(time.Second * 2):
		t.Error("timeout waiting for hook call back")
	case <-done:
	}

	// ankur client send msg
	msg = []byte("hello everyone\n\r")
	writeMsg(t, cc1, msg)
	for _, cl := range []net.Conn{cc2, cc3} {
		readM := make([]byte, 512)
		err := readMsg(t, cl, readM)
		must(t, err)
		if !bytes.Contains(readM, []byte("hello everyone")) {
			t.Errorf("expected msg: %s not found in received msg", "hello everyone")
		}
	}
}

func initialRead(t *testing.T, cc net.Conn, name []byte) {
	t.Helper()
	defer func() {
		err := cc.SetDeadline(time.Time{})
		must(t, err)
	}()
	err := cc.SetDeadline(time.Now().Add(time.Millisecond * 100))
	must(t, err)
	b := make([]byte, 4096)
	_, err = cc.Read(b)
	must(t, err)
	// send the name
	_, err = cc.Write(name)
	must(t, err)
	b = make([]byte, 4096)
	// read the help response helpDMsg
	_, err = cc.Read(b)
	must(t, err)
	// read the infoPrompt
	b = make([]byte, 128)
	_, err = cc.Read(b)
	must(t, err)
}

func writeMsg(t *testing.T, cc net.Conn, msg []byte) {
	t.Helper()
	defer func() {
		err := cc.SetDeadline(time.Time{})
		must(t, err)
	}()
	err := cc.SetDeadline(time.Now().Add(time.Millisecond * 100))
	must(t, err)
	_, err = cc.Write(msg)
	must(t, err)
}

func readMsg(t *testing.T, cc net.Conn, msg []byte) error {
	t.Helper()
	defer func() {
		err := cc.SetDeadline(time.Time{})
		must(t, err)
	}()
	err := cc.SetDeadline(time.Now().Add(time.Millisecond * 100))
	must(t, err)
	_, err = cc.Read(msg)
	return err
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Error(err)
	}
}
