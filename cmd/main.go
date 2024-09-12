package main

import (
	"fmt"
	"os"
	"time"
	config "vishalvivekm/vcs/constants"
	"vishalvivekm/vcs/handler"
	"vishalvivekm/vcs/types"
	"vishalvivekm/vcs/utils"
)
const (
	logFile    = "vcs/log.txt"
	indexFile  = "vcs/index.txt"
	configFile = "vcs/config.txt"
	commitsDir = "vcs/commits"
)

func main() {

	Args := os.Args
	if len(os.Args) < 2 {
		// print help
		fmt.Println(config.HelpTxt)
		return
	}

	err := os.Mkdir("vcs", 0755)
	if err != nil {
		if os.IsExist(err) {
			// do nothing
		} else {
			utils.Check(fmt.Errorf("error while creating vcs directory: %w", err))
		}
	}

	_, err = os.Stat(config.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			utils.Check(utils.CreateFile(config.ConfigFile, ""))
		} else {
			utils.Check(err)
		}
	}

	_, err = os.Stat(config.IndexFile)
	if err != nil {
		if os.IsNotExist(err) {
		utils.CreateFile(config.IndexFile, "")
	    } else {
		utils.Check(err)
	    }
	}
	switch Args[1] {
	case "config":
		handler.ConfigUser(Args)
	case "add":
		handler.AddFilesToIndex(Args)
	case "log":
		handler.DisplayLogs()
	case "commit":
		author := readFileContent(config.ConfigFile)
		if string(author) == "" {
			fmt.Println("Please configure Username first") // can prompt for help ? y or n : showHelp
			return
		}
		content := string(readFileContent(config.IndexFile))
		if len(content) == 0 {
			fmt.Println("No tracked files, add a file to the index first")
			return
		}
		if len(Args) < 3 {
			fmt.Println("Please make sure to pass the commit msg")
			return
		}

		cm := types.CommitObject{
			Author: string(author),
			Msg:    Args[2],
			Date: time.Now().Format("2006-01-02 15:04:05"),
		}
		// content = string(readFileContent(config.IndexFile))
		// if len(content) == 0 {
		// 	fmt.Println("No tracked files, add a file to the index first")
		// 	return
		// }
		handler.Commit(&cm)
		

	default:
		fmt.Printf("command: %s, not supported yet\n\nAvailable commands:\n%s\n", Args[1], config.HelpTxt)
	}

}

func readFileContent(name string) []byte {

	data, err := os.ReadFile(name)
	utils.Check(err)
	return data

}

