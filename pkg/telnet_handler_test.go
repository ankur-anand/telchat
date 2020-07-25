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

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Error(err)
	}
}
