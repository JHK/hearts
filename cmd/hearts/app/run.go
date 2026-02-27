package app

import (
	"flag"
	"log"

	"github.com/JHK/hearts/internal/webui"
)

func Run() {
	addr := flag.String("addr", "127.0.0.1:8080", "web listen address")
	flag.Parse()

	log.Printf("Hearts web server listening at http://%s", *addr)
	if err := webui.Run(webui.Config{Addr: *addr}); err != nil {
		log.Fatalf("web server failed: %v", err)
	}
}
