package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Read last run info from saved file
func getLastRunInfo(statusFileName string) (int, int, [4]int, error) {

	var (
		f         io.Reader
		scanner   *bufio.Scanner
		err       error
		lastTime  int
		lastCount int
		r         int
		last      [4]int
	)

	if statusFileName == "" {
		statusFileName = "status.file"
	}

	f, err = os.Open(statusFileName)

	// If not exist or error, just ignore, hope we can create when done
	if err == nil {
		scanner = bufio.NewScanner(f)

		// Read values in loop
		for i := 0; i < 4; i++ {
			scanner.Scan()
			r, err = strconv.Atoi(strings.TrimSpace(scanner.Text()))
			checkErr(err)
			last[i] = r
		}

		// Choose our values
		switch argStatsMetric {
		case "r":
			lastTime = last[0]
			lastCount = last[1]
		case "e":
			lastTime = last[2]
			lastCount = last[3]
		case "l": // Do nothing
		default:
			panic("Invdalid Stats Metric")
		}
	} else {
		if flagVerbose {
			fmt.Printf("Error opening status file: %s, ignoring.\n\n", statusFileName)
		}
		lastTime = 0  // Set zero if no file or error
		lastCount = 0 // Set zero if no file or error
	}
	return lastTime, lastCount, last, nil
}

// Save last run info to saved file
func saveLastRunInfo(statusFile string, lastRunTime int, lastRunCount int, last [4]int) (error) {
	var (
		fw  io.Writer
		err error
		s   string
	)

	if statusFile == "" {
		statusFile = "status.file"
	}
	fw, err = os.Create(statusFile)
	checkErr(err)

	switch argStatsMetric {
	case "r":
		last[0] = lastRunTime
		last[1] = lastRunCount
	case "e":
		last[2] = lastRunTime
		last[3] = lastRunCount
	case "l": // Do nothing
	default:
		panic("Invdalid Stats Metric in saveLastRunInfo")
	}

	// Loop writing
	for i := 0; i < 4; i++ {
		s = fmt.Sprintf("%d\n", last[i])
		_, err = io.WriteString(fw, s)
		checkErr(err)
	}

	return nil
}
