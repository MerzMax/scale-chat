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

const loadTestResultFilesDir = "../results/"
const outputDir = loadTestResultFilesDir + "graphics/"

func main() {
	// READ IN FILES
	var filePaths []string

	err := filepath.Walk(loadTestResultFilesDir, func(path string, info os.FileInfo, err error) error {
		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	filePaths = filterFilePaths(filePaths)

	// PARSE THE CSV DATA

	// key = File name
	fileData := make(map[string][]client.MessageEventEntry)

	for _, filePath := range filePaths {
		res, err := parseCsvFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		fileData[trimFilePathAndCsvSuffix(filePath)] = res

		log.Println(trimFilePathAndCsvSuffix(filePath) + ": File was parsed successfully")
	}

	// CALCULATE RTT for each message

	// key = File name
	roundTripTimeEntriesMap := make(map[string][]MessageLatencyEntry)

	for key, data := range fileData {
		rttEntries := calculateMessageLatency(data)
		roundTripTimeEntriesMap[key] = rttEntries

		log.Println(key + ": Data was processed successfully")

		Plot(key, &rttEntries)

		log.Println(key + ": Data was plotted successfully")
	}

	log.Println("")
	log.Println("------------------------------------------------------------------------------------")
	log.Println("COMPLETED")
	log.Println("Files will be stored at: " + outputDir)
	log.Println("------------------------------------------------------------------------------------")
}

type MessageLatencyEntry struct {
	SenderMsgEvent client.MessageEventEntry
	SenderId       string
	MessageId      uint64
	RttInNs        int64
	LatenciesInNs  []int64
}

// Filter a list of file paths. All csv files will be returned
func filterFilePaths(files []string) []string {
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

	log.Println(trimFilePathAndCsvSuffix(filepath) + ": File will be parsed...")

	// Open the csv file
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("could not open the file: ", err)
		return nil, err
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
		if i <= 0 {
			continue
		}

		var msgEventEntry client.MessageEventEntry

		for j, field := range line {

			switch j {
			case 0:
				number, err := strconv.ParseUint(field, 10, 64)
				if err != nil {
					return nil, err
				}
				msgEventEntry.MessageId = number
			case 1:
				msgEventEntry.SenderId = field
			case 2:
				msgEventEntry.ClientId = field
			case 3:
				switch field {
				case "Sent":
					msgEventEntry.Type = client.Sent
				case "Received":
					msgEventEntry.Type = client.Received
				default:
					return nil, errors.New("Unknown type: " + field)
				}
			case 4:
				seconds, err := strconv.ParseInt(field, 10, 64)
				if err != nil {
					return nil, err
				}
				msgEventEntry.TimeStamp = time.UnixMicro(seconds)
			}
		}

		msgEventEntries = append(msgEventEntries, msgEventEntry)
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

func trimFilePathAndCsvSuffix(filepath string) string {
	return strings.TrimPrefix(strings.TrimSuffix(filepath, ".csv"), loadTestResultFilesDir)
}
