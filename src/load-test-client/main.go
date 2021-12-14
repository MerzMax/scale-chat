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
	loadTest := flag.Bool("load-test", false, "Flag indicates weather the client should start in "+
		"the load test mode")

	serverUrl := flag.String("server-url", "ws://localhost:8080/ws", "The url of the server to "+
		"connect to")

	msgFrequency := flag.Int("msg-frequency", 1000, "The frequency of the messages in ms (just "+
		"for load test mode")

	msgSize := flag.Int("msg-size", 256, "The size of the messages in bytes (just for load test "+
		"mode)")

	numOfClients := flag.Int("clients", 1, "Number of clients that will be started (just for "+
		"load test mode")

	flag.Parse()

	// Listen to system interrupts -> program will be stopped
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, os.Interrupt)

	var msgEvents chan *client.MessageEventEntry

	// If the application isn't started in load test mode there is just one client that will be started.
	// If the application is in load test mode, a csv file with client rtt will be written
	if !*loadTest {
		*numOfClients = 1
	} else {
		msgEvents = make(chan *client.MessageEventEntry)
		go processMessageEvents(msgEvents)
	}

	// Create numOfClients clients that can chat
	for i := 0; i < *numOfClients; i++ {
		log.Printf("Creating client number: %v / %v", i+1, *numOfClients)

		// Listen to system interrupts -> program will be stopped
		closeConnection := make(chan os.Signal, 1)
		signal.Notify(closeConnection, os.Interrupt)

		go func() {
			chatClient := client.Client{
				ServerUrl:        *serverUrl,
				CloseConnection:  closeConnection,
				IsLoadTestClient: *loadTest,
				MsgFrequency:     *msgFrequency,
				MsgSize:          *msgSize,
				MsgEvents:        msgEvents,
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
func processMessageEvents(messageEvents chan *client.MessageEventEntry) {
	// Create new file and prepare writer
	fileName := "load-test-client-" + time.Now().Format("2006-01-02-15-04-05") + ".csv"
	file, err := os.Create("./load-test-results/" + fileName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)

	// Process incoming messages
	for {
		select {
		case msgEventEntry := <-messageEvents:
			err = csvWriter.Write(msgEventEntry.StringArray())
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
	}
}
