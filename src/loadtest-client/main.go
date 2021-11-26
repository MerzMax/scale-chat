package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
	"os/signal"
	"scale-chat/client"
	"time"
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

	var msgEvents chan client.MessageEventEntry

	// If the application isn't started in loadtest mode there is just one client that will be started.
	if !*loadtest {
		*numOfClients = 1
	} else {
		msgEvents = make(chan client.MessageEventEntry)
	}

	go processMessageEvents(msgEvents)

	// Create numOfClients clients that can chat
	for i := 0; i < *numOfClients; i++ {
		log.Printf("Creating client number: %v / %v", i + 1, *numOfClients)

		// Listen to system interrupts -> program will be stopped
		closeConnection := make(chan os.Signal, 1)
		signal.Notify(closeConnection, os.Interrupt)

		go func() {
			chatClient := client.Client{
				ServerUrl:        	*serverUrl,
				CloseConnection:  	closeConnection,
				IsLoadtestClient: 	*loadtest,
				MsgFrequency: 		*msgFrequency,
				MsgSize: 			*msgSize,
				MessageEvents: 		msgEvents,
			}

			err := chatClient.Start()
			if err != nil {
				log.Fatalf("%v", err)
			}
		}()
	}

	select {
	case <-sysInterrupt:
		log.Println("SYS INTERRUPT")
	}
}

// The function processes MessageEventEntries and writes a csv with the data collected
func processMessageEvents (messageEvents chan client.MessageEventEntry) {
	// Create new file and prepare writer
	fileName := "loadtest-client-" +  time.Now().Format("02-01-2006-15-04-05") + ".csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("%v", err)
	}

	csvWriter := csv.NewWriter(file)

	// Process incoming messages
	for {
		select {
		case msgEventEntry := <- messageEvents:
			err = csvWriter.Write(msgEventEntry.ConvertToStringArray())
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
	}
}

