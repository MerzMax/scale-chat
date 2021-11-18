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
	numOfClients := flag.Int("clients", 1, "Number of clients that will be started (just for " +
		"loadtest mode")
	flag.Parse()

	// Listen to system interrupts -> program will be stopped
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, os.Interrupt)

	clientsCloseConnection := make([]chan string, 0)

	// If the application isn't started in loadtest mode there is just one client that will be started.
	if !*loadtest {
		*numOfClients = 1
	}

	// Create numOfClients clients that can chat
	for i := 0; i < *numOfClients; i++ {

		log.Printf("Creating client number: %v / %v", i + 1, numOfClients)

		closeConnection := make(chan string, 1)
		clientsCloseConnection = append(clientsCloseConnection, closeConnection)

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

	}

	for {
		select {
		case <-sysInterrupt:
			for _, closeConnection := range clientsCloseConnection {
				closeConnection <- "Closing connection due to system interrupt."
				close(closeConnection)
			}
		}
	}
}

func startClient(closeConnection chan string){

}
