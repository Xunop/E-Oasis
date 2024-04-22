package main

import (
	"github.com/Xunop/e-oasis/log"
)

func main() {
	//StartServer()
	log.Info("Server started")
	defer log.Logger.Sync()
}
