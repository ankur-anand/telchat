package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/ankur-anand/telchat/pkg"
)

var configLocation = flag.String("config", "./config.json", "config.json file location")

type config struct {
	LogFile    string `json:"log_file"`
	TelnetAddr string `json:"telnet_addr"`
	HTTPAddr   string `json:"http_addr"`
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[telchat] ")

	flag.Parse()
	// read file.
	cb, err := ioutil.ReadFile(*configLocation)
	if err != nil {
		log.Fatal(err)
	}

	var cg config
	err = json.Unmarshal(cb, &cg)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cs, err := pkg.NewChatServer(cg.LogFile)
	log.Printf("log will be written to %s \n", cg.LogFile)
	if err != nil {
		log.Fatalln(err)
	}
	go cs.ServeHTTP(cg.HTTPAddr)
	go cs.ServeTelnet(cg.TelnetAddr)

	<-c
	log.Printf("shutting down server")
	cs.Shutdown()
}
