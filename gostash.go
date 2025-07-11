package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Stash struct {
	FilesByDate map[string][]string `json:"files_by_date"`
}

func save_config(stash Stash) {
	fmt.Println("Saving config file:", ".gostash.json")

	file, err := os.OpenFile("./.gostash.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	jsonData, err := json.Marshal(stash)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
	}
}

func get_today_gostash_folder() string {
	return "gostash-" + time.Now().Format("2006-01-02")
}

func load_config(stash *Stash) {
	check_config_exists()

	fmt.Println("Loading config file:", ".gostash.json")

	file, err := os.Open("./.gostash.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	// Decode JSON directly into the stash pointer
	dec := json.NewDecoder(file)
	if err := dec.Decode(stash); err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
}

func moveFile(src, dst string) error {
	return os.Rename(src, dst) // Moves the file from src to dst
}

func create_gostash_folder() {
	folder_name := get_today_gostash_folder()
	err := os.Mkdir(folder_name, 0755)
	if err != nil {
		fmt.Println("Error creating gostash folder", err)
	}
}

func check_name_in_blacklist(name string) bool {
	today := get_today_gostash_folder()
	blacklist := []string{".gostash.json", today}
	for _, item := range blacklist {
		if item == name {
			return true
		}
	}
	return false
}

func stash(stash *Stash) {
	create_gostash_folder()
	fmt.Println("Stashing")

	todayFolder := get_today_gostash_folder()
	if stash.FilesByDate == nil {
		stash.FilesByDate = make(map[string][]string)
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || check_name_in_blacklist(entry.Name()) {
			continue
		}

		srcPath := entry.Name()
		dstPath := filepath.Join(todayFolder, entry.Name())

		// Move file to stash folder
		if err := moveFile(srcPath, dstPath); err != nil {
			fmt.Println("Error moving file:", err)
			continue
		}

		// Append the destination path to the list for today's date
		stash.FilesByDate[todayFolder] = append(stash.FilesByDate[todayFolder], dstPath)
	}

	save_config(*stash)
	fmt.Println("Stashed", len(stash.FilesByDate[todayFolder]), "files")
}

func unstash(stash *Stash) {
	// Get date from user or use today's date
	var dateFolder string
	if len(os.Args) > 2 {
		dateFolder = "gostash-" + os.Args[2]
	} else {
		dateFolder = get_today_gostash_folder()
	}

	filesToUnstash, found := stash.FilesByDate[dateFolder]
	if !found || len(filesToUnstash) == 0 {
		fmt.Printf("No files found for date: %s\n", dateFolder)
		return
	}

	for _, filePath := range filesToUnstash {
		fileName := filepath.Base(filePath)
		destPath := "./" + fileName

		// Move file back to the root directory
		if err := os.Rename(filePath, destPath); err != nil {
			fmt.Printf("Error moving file %s back to root: %v\n", filePath, err)
			continue
		}
	}

	// Remove the date entry after unstashing
	delete(stash.FilesByDate, dateFolder)
	save_config(*stash)
	fmt.Printf("Unstashed %d files from %s\n", len(filesToUnstash), dateFolder)
}

func check_config_exists() {
	configPath := "./.gostash.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Config file does not exist")
		create_config()
	} else {
		fmt.Println("Config file exists")
	}
}

func create_config() {
	configPath := "./.gostash.json"

	file, err := os.Create(configPath)
	if err != nil {
		fmt.Println("Error creating config file")
	}
	defer file.Close()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gostash <command> [date]")
		os.Exit(1)
	}

	check_config_exists()

	stashConfig := &Stash{}
	load_config(stashConfig)

	switch os.Args[1] {
	case "stash":
		stash(stashConfig)
	case "unstash":
		unstash(stashConfig)
	}
}
