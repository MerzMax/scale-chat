package main

import (
	"encoding/csv"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type LoadTest struct {
	roomCount uint
	roomSize  uint
	filename  string
}

func parseFile(filename string) []Message {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("Cannot read file:", filename)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal("Failed to close", filename)
		}
	}(f)

	csvReader := csv.NewReader(f)

	lines, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Cannot parse CSV:", filename)
	}

	messages := make([]Message, len(lines))

	for i, line := range lines {
		seconds, err := strconv.ParseInt(line[4], 10, 64)
		if err != nil {
			log.Printf("Cannot parse line %d in file %s", i, filename)
			continue
		}

		messages[i] = Message{
			id:        line[0],
			sender:    line[1],
			receiver:  line[2],
			msgType:   line[3],
			timestamp: time.UnixMicro(seconds),
		}
	}

	return messages
}

func createOutDir(metaFile *string) (*string, error) {
	outerPath := "./out"
	_, innerPath := path.Split(*metaFile)
	innerPath = strings.TrimSuffix(innerPath, path.Ext(innerPath))
	innerPath = path.Join(outerPath, innerPath)

	for _, p := range []string{outerPath, innerPath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			continue
		}
		err := os.Mkdir(p, 0755)
		if err != nil {
			return nil, err
		}
	}

	return &innerPath, nil
}

func loadResultFiles(metaFile *string) []LoadTest {
	f, err := os.Open(*metaFile)
	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(f)
	lines, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var loadTests []LoadTest

	for _, line := range lines {
		roomCount, err := strconv.ParseUint(line[0], 10, 32)
		if err != nil {
			log.Fatal(err)
		}

		roomSize, err := strconv.ParseUint(line[1], 10, 32)
		if err != nil {
			log.Fatal(err)
		}

		loadTests = append(loadTests, LoadTest{
			roomCount: uint(roomCount),
			roomSize:  uint(roomSize),
			filename:  line[2],
		})
	}

	if len(loadTests) < 1 {
		log.Fatal("Could not find any load test files")
	}

	return loadTests
}

func writeResults(filename string, results [][]string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Failed to create output file", err)
	}

	csvWriter := csv.NewWriter(f)
	err = csvWriter.WriteAll(results)
	if err != nil {
		log.Fatal("Failed to write output", err)
	}

	csvWriter.Flush()
	err = f.Close()
	if err != nil {
		log.Println("Failed to close CSV file", err)
	}
}
