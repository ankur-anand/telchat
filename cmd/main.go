package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/ankur-anand/telchat/pkg"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[telchat] ")
	dir := os.TempDir()
	file := dir + "/telchat.log"
	cs, err := pkg.NewChatServer(file)
	log.Printf("log will be written to %s \n", file)
	if err != nil {
		log.Fatalln(err)
	}
	go cs.ServeHTTP(":3002")
	go cs.ServeTelnet(":3001")

	<-c
	log.Printf("shutting down server")
	cs.Shutdown()
}
