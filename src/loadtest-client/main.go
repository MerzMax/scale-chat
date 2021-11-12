package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"scale-chat/client"
)

func main() {
	// Read cmd line arguments
	loadtest := flag.Bool("loadtest", false, "Flag indicates weather the client should start in " +
		"the loadtest mode")
	serverUrl := flag.String("server-url", "ws://localhost:8080/ws", "The url of the server to " +
		"connect to")
	msgFrequency := flag.Int("msg-frequency", 1000, "The frequency of the messages in ms (just " +
		"for loadtest mode")
	msgSize := flag.Int("msg-size", 256, "The size of the messages in bytes (just for loadtest " +
		"mode)")
	flag.Parse()

	// Listen to system interrupts -> program will be stopped
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, os.Interrupt)

	closeConnection := make(chan string, 1)

	go func() {
		client := client.Client{
			ServerUrl:        	*serverUrl,
			CloseConnection:  	closeConnection,
			IsLoadtestClient: 	*loadtest,
			MsgFrequency: 		*msgFrequency,
			MsgSize: 			*msgSize,
		}

		err := client.Start()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}()

	for {
		select {
		case <-sysInterrupt:
			closeConnection <- "Closing connection due to system interrupt."
			close(closeConnection)
		}
	}
}

