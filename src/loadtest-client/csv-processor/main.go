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

func main() {
	// READ IN FILES
	var fileNames []string

	root := "../loadtest-results"
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
		fileData[fileName] = res
	}

	// CALCULATE RTT for each message
	roundTripTimes := make(map[string][]RoundTripTimeEntry)

	for key, data := range fileData {
		rttEntries := calculateRtts(data)
		roundTripTimes[key] = rttEntries
	}

	log.Println("---------------------")
	log.Println("")
	log.Println("COMPLETED")
	log.Println("")
	log.Println("---------------------")
}

type RoundTripTimeEntry struct {
	ClientId string
	MessageId string
	rttInMs uint
}

// Filter a list of filenames. All csv files will be returned
func filterFileNames(files []string) []string {
	var res []string
	for _, file := range files{
		if strings.HasSuffix(file, ".csv") {
			res = append(res, file)
		}
	}
	return res
}

// Parse a csv file at the filepath and convert the data in an array of MessageEventEntry structs
func parseCsvFile(filepath string) ([]client.MessageEventEntry, error) {
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

// Convert an array of string arrays int an array of MessageEventEntry structs
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
					msgEventEntry.ClientId = field
				} else if j == 2 {
					if field == "Sent" {
						msgEventEntry.Type = client.Sent
					} else if field == "Received" {
						msgEventEntry.Type = client.Received
					} else {
						return nil, errors.New("Unknown type: " + field)
					}
				} else if j == 3 {
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

func calculateRtts(msgEventEntries []client.MessageEventEntry) []RoundTripTimeEntry {
	var rttEntries []RoundTripTimeEntry
	
	return rttEntries
}

