package main

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"path/filepath"
	"scale-chat/client"
	"strconv"
	"strings"
	"time"
)

const root = "../load-test-results/"
const outputDir = "./load-test-graphics/"

func main() {
	// READ IN FILES
	var fileNames []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		fileNames = append(fileNames, path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fileNames = filterFileNames(fileNames)

	// PARSE THE CSV DATA
	fileData := make(map[string][]client.MessageEventEntry)

	for _, fileName := range fileNames {
		res, err := parseCsvFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(strings.TrimPrefix(strings.TrimSuffix(fileName, ".csv"), root))

		fileData[strings.TrimPrefix(strings.TrimSuffix(fileName, ".csv"), root)] = res
	}

	// CALCULATE RTT for each message

	// key = File name
	roundTripTimeEntriesMap := make(map[string][]MessageLatencyEntry)

	for key, data := range fileData {
		rttEntries := calculateMessageLatency(data)
		roundTripTimeEntriesMap[key] = rttEntries

		PlotRtts(key, &rttEntries)
	}

	//log.Println(roundTripTimeEntriesMap)

	log.Println("")
	log.Println("---------------------")
	log.Println("COMPLETED CALCULATIONS")
	log.Println("---------------------")
	log.Println("")
}

type MessageLatencyEntry struct {
	SenderMsgEvent client.MessageEventEntry
	SenderId       string
	MessageId      uint64
	RttInNs        int64
	LatenciesInNs  []int64
}

// Filter a list of filenames. All csv files will be returned
func filterFileNames(files []string) []string {
	var res []string
	for _, file := range files {
		if strings.HasSuffix(file, ".csv") {
			res = append(res, file)
		}
	}
	return res
}

// Parses a csv file at the filepath and convert the data in an array of MessageEventEntry structs
func parseCsvFile(filepath string) ([]client.MessageEventEntry, error) {

	log.Println("parsing CSV file: " + filepath)

	// Open the csv file
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Reading everything from the csv file
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	msgEventEntries, err := parseMessageEventEntries(data)
	if err != nil {
		return nil, err
	}

	return msgEventEntries, nil
}

// Converts an array of string arrays into an array of MessageEventEntry structs
func parseMessageEventEntries(data [][]string) ([]client.MessageEventEntry, error) {
	var msgEventEntries []client.MessageEventEntry

	for i, line := range data {
		if i > 0 { // omit header line
			var msgEventEntry client.MessageEventEntry
			for j, field := range line {
				if j == 0 {
					number, err := strconv.ParseUint(field, 10, 64)
					if err != nil {
						return nil, err
					}
					msgEventEntry.MessageId = number
				} else if j == 1 {
					msgEventEntry.SenderId = field
				} else if j == 2 {
					msgEventEntry.ClientId = field
				} else if j == 3 {
					if field == "Sent" {
						msgEventEntry.Type = client.Sent
					} else if field == "Received" {
						msgEventEntry.Type = client.Received
					} else {
						return nil, errors.New("Unknown type: " + field)
					}
				} else if j == 4 {
					seconds, err := strconv.ParseInt(field, 10, 64)
					if err != nil {
						return nil, err
					}
					msgEventEntry.TimeStamp = time.UnixMicro(seconds)
				}
			}
			msgEventEntries = append(msgEventEntries, msgEventEntry)
		}
	}

	return msgEventEntries, nil
}

// Calculates the RTT for a given set of MessageEventEntry structs
func calculateMessageLatency(msgEventEntries []client.MessageEventEntry) []MessageLatencyEntry {
	var messageLatencies []MessageLatencyEntry

	var sentMsgEventEntries []client.MessageEventEntry
	var receivedMsgEventEntries []client.MessageEventEntry

	// Filter for send and received messages
	for _, msgEventEntry := range msgEventEntries {
		if msgEventEntry.Type == client.Sent {
			sentMsgEventEntries = append(sentMsgEventEntries, msgEventEntry)
		} else if msgEventEntry.Type == client.Received {
			receivedMsgEventEntries = append(receivedMsgEventEntries, msgEventEntry)
		}
	}

	// Pair sent and received messages to calculate rtt and latencies
	for _, sentMsgEventEntry := range sentMsgEventEntries {

		msgLatency := MessageLatencyEntry{
			MessageId:      sentMsgEventEntry.MessageId,
			SenderId:       sentMsgEventEntry.ClientId,
			SenderMsgEvent: sentMsgEventEntry,
		}

		for _, receivedMsgEventEntry := range receivedMsgEventEntries {
			if sentMsgEventEntry.MessageId == receivedMsgEventEntry.MessageId {
				if sentMsgEventEntry.SenderId == receivedMsgEventEntry.SenderId {
					msgLatency.RttInNs = receivedMsgEventEntry.TimeStamp.Sub(sentMsgEventEntry.TimeStamp).Nanoseconds()
					msgLatency.LatenciesInNs = append(msgLatency.LatenciesInNs, msgLatency.RttInNs)
				} else {
					latency := receivedMsgEventEntry.TimeStamp.Sub(sentMsgEventEntry.TimeStamp).Nanoseconds()
					msgLatency.LatenciesInNs = append(msgLatency.LatenciesInNs, latency)
				}
			}
		}

		messageLatencies = append(messageLatencies, msgLatency)
	}

	return messageLatencies
}
