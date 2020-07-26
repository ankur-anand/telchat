package main

import (
	"log"

	"github.com/ankur-anand/telchat/pkg"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[telchat] ")

	cs := pkg.NewChatServer()
	cs.ServeTelnet(":3001")
}
