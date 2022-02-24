package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
)

type Command struct {
	name      string
	processor func(loadTest LoadTest, messages []Message) [][]string
}

func parseFlags() (*string, *Command, error) {
	var command string
	var metaFile string

	flag.StringVar(&command, "command", "",
		"Command you want me to execute. I can calculate latencies with 'latency'"+
			" or count sent and received messages with 'sent-vs-received'")
	flag.StringVar(&metaFile, "file", "", "Meta file holding load test result file names")
	flag.Parse()

	if len(os.Args[1:]) < 2 || len(metaFile) < 1 {
		return nil, nil, errors.New(
			fmt.Sprintf("Invalid signature. Use --help to see the necessary options."),
		)
	}

	commands := []*Command{
		{
			name:      "sent-vs-received",
			processor: CompareCount,
		},
		{
			name:      "latency",
			processor: CalculateLatency,
		},
	}

	for _, c := range commands {
		if command == c.name {
			return &metaFile, c, nil
		}
	}

	return nil, nil, errors.New("unknown command")
}

func main() {
	metaFile, command, err := parseFlags()
	if err != nil {
		log.Fatal(err.Error())
	}

	outDir, err := createOutDir(metaFile)
	if err != nil {
		log.Fatal("Failed to create out directory:", err)
	}
	loadTests := loadResultFiles(metaFile)
	log.Println("Loaded result files")

	wg := &sync.WaitGroup{}
	wg.Add(len(loadTests))

	for _, loadTest := range loadTests {
		loadTest := loadTest
		go func() {
			messages := parseFile(loadTest.filename)
			results := command.processor(loadTest, messages)
			writeResults(
				path.Join(
					*outDir,
					fmt.Sprintf("%s_%d-%d.csv", command.name, loadTest.roomCount, loadTest.roomSize),
				),
				results,
			)
			log.Printf("Finished processing file for %d rooms with %d clients each",
				loadTest.roomCount, loadTest.roomSize)
			wg.Done()
		}()
	}

	wg.Wait()
}
