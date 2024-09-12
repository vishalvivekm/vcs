package utils

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strings"
	config "vishalvivekm/vcs/constants"
	"sync"
)

const (
	logFile = "vcs/log.txt"
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

func ReturnHash(content []byte) (string, error) {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return "", err
	}
	hashedString := fmt.Sprintf("%x", hash.Sum(nil))
	return hashedString, nil
}

func ReadFileContent(name string) ([]byte, error) {

	data, err := os.ReadFile(name)
	if err != nil {
	return nil ,fmt.Errorf("error reading file, %s, %w", name, err)
	}
	return data, nil

}
func WriteLogs(logMsg string) error {
	file, err := os.OpenFile(config.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	content, err := ReadFileContent(file.Name())
	Check(err)
	if err := os.WriteFile(file.Name(), []byte(logMsg), os.ModePerm); err != nil {
		return err
	}
	if _, err = fmt.Fprint(file, "\n\n", string(content)); err != nil {
		return err
	}
	return nil
}


func ReadIndex(wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()
	file, err := os.Open(config.IndexFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		singleFile := scanner.Text()
		if len(singleFile) != 0 {
			ch <- singleFile
		}
	}
	if err := scanner.Err(); err != nil {
		Check(err)
	}
	close(ch)
}

func ReadFilesAndCopy(wg *sync.WaitGroup, ch chan string, commitHash string) {
	defer wg.Done()
	for fileName := range ch {
		content, err := ReadFileContent(fileName)
		err = os.WriteFile(fmt.Sprintf("%s/%s/%s", config.CommitsDir, commitHash, fileName), content, os.ModePerm)
		if err != nil {
			log.Fatalf("Error writing file %s to commit: %v", fileName, err)
		}
		fmt.Printf("Copied file:%s to the commit %s\n", fileName, commitHash)
	}
}

