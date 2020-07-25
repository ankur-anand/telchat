package pkg

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
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

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Error(err)
	}
}
