package main

import (
	"log"
	"os"
	"os/signal"
	"scale-chat/client"
)

func main() {
	// Listen to system interrupts -> program will be stopped
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, os.Interrupt)

	client := client.Client{
		ServerUrl:    "ws://localhost:8080/ws",
		SysInterrupt: sysInterrupt,
	}

	err := client.Start()
	if err != nil {
		log.Fatalf("%v", err)
	}
}

