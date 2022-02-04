package main

import (
	"context"
	"encoding/csv"
	"flag"
	"github.com/google/uuid"
	"log"
	"os"
	"os/signal"
	"scale-chat/client"
	"sync"
	"time"
)

func main() {
	// Read cmd line arguments
	loadTest := flag.Bool("load-test", false,
		"Flag indicates weather the client should start in the load test mode")

	serverUrl := flag.String("server-url", "ws://localhost:8080/ws",
		"The url of the server to connect to")

	msgFrequency := flag.Int("msg-frequency", 1000,
		"The frequency of the messages in ms (just for load test mode")

	msgSize := flag.Int("msg-size", 256,
		"The size of the messages in bytes (just for load test mode)")

	numOfRooms := flag.Int("rooms", 1,
		"Number of chat rooms that will be initialized (just for load test mode)")

	roomSize := flag.Int("room-size", 1,
		"Number of clients that will be started per room (just for load test mode")

	flag.Parse()

	var msgEvents chan *client.MessageEventEntry
	waitGroup := &sync.WaitGroup{}
	numOfWaits := (*numOfRooms * *roomSize) + 1
	if !*loadTest {
		*numOfRooms = 1
		*roomSize = 1
		numOfWaits = 1
	}

	var cancelFuncs []*context.CancelFunc

	// If the application isn't started in load test mode there is just one client that will be started.
	// If the application is in load test mode, a csv file with client rtt will be written
	if *loadTest {
		msgEvents = make(chan *client.MessageEventEntry, 100)
		ctx, cancelFunc := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, &cancelFunc)
		go processMessageEvents(msgEvents, ctx, waitGroup)
	}

	// Create rooms
	for i := 0; i < *numOfRooms; i++ {
		room := uuid.New().String()

		// Create clients per room
		for j := 1; j <= *roomSize; j++ {
			log.Printf("Creating client number: %v / %v", j, *roomSize)

			// Listen to system interrupts -> program will be stopped
			closeConnection := make(chan os.Signal, 1)
			signal.Notify(closeConnection, os.Interrupt)

			ctx, cancelFunc := context.WithCancel(context.Background())
			cancelFuncs = append(cancelFuncs, &cancelFunc)

			go func() {
				chatClient := client.Client{
					Context:          ctx,
					WaitGroup:        waitGroup,
					ServerUrl:        *serverUrl,
					CloseConnection:  closeConnection,
					IsLoadTestClient: *loadTest,
					MsgFrequency:     *msgFrequency,
					MsgSize:          *msgSize,
					MsgEvents:        msgEvents,
					Room:             room,
				}

				err := chatClient.Start()
				if err != nil {
					log.Fatalf("%v", err)
				}
			}()
		}
	}

	waitGroup.Add(numOfWaits)

	// Listen to system interrupts -> program will be stopped
	sysInterrupt := make(chan os.Signal, 1)
	signal.Notify(sysInterrupt, os.Interrupt)

	<-sysInterrupt
	log.Println("Shutting down clients...")

	for _, cancelFunc := range cancelFuncs {
		(*cancelFunc)()
	}
	waitGroup.Wait()
	log.Println("All clients finished, shutting down now")
}

// The function processes MessageEventEntries and writes a csv with the data collected
func processMessageEvents(
	messageEvents <-chan *client.MessageEventEntry,
	ctx context.Context,
	waitGroup *sync.WaitGroup,
) {
	// Create new file and prepare writer
	fileName := "load-test-client-" + time.Now().Format("2006-01-02-15-04-05") + ".csv"
	file, err := os.Create("./results/" + fileName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Fatal("Failed to close CSV file:", err)
		}

		log.Println("CSV file is closed")
		waitGroup.Done()
	}()
	csvWriter := csv.NewWriter(file)

	// Process incoming messages
outer:
	for {
		select {
		case entry := <-messageEvents:
			err = csvWriter.Write(entry.StringArray())
			if err != nil {
				log.Fatalf("%v", err)
			}
		case <-ctx.Done():
			break outer
		}
	}

	// Write remaining bytes in buffer to file
	csvWriter.Flush()
}
