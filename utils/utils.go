package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	logFile = "../vcs/log.txt"
)

func CompareCommit(commitID string) (bool, error) {
	// read first line of logFile, get commit hash
	// commitId == commitHash ? return true(no new changes from prev commit) : return false //

	logfile, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) { // no commits to compare, return false with no error
			return false, nil
		} else {
			return false, fmt.Errorf("error while opening logFile: %w", err)
		}
	}
	defer logfile.Close()
	scanner := bufio.NewScanner(logfile)
	line := 1
	for scanner.Scan() {
		if line == 2 {
			break
		}
		preCommitID := strings.Split(scanner.Text(), " ")[1]
		if preCommitID == commitID {
			return true, nil
		} else {
			line++
		}

	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error while scanning logFile: %w", err)
	}
	return false, nil
}

func CreateFile(name string, data string) error {
	d := []byte(data)
	err := os.WriteFile(name, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Check(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}
