// gostash.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StashMode string

const (
	ModeSimple     StashMode = "simple"
	ModeCategorize StashMode = "categorize"
)

type StashConfig struct {
	FilesByDate map[string]StashEntry `json:"files_by_date"`
}

type StashEntry struct {
	Mode       StashMode           `json:"mode"`
	Files      []string            `json:"files"`
	Folders    []string            `json:"folders"`
	Categories map[string][]string `json:"categories,omitempty"` // category -> files
}

type CLI struct {
	Command    string
	Date       string
	Categorize bool
	Help       bool
}

// File type detection
func getFileCategory(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	categories := map[string][]string{
		"images":      {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".ico"},
		"documents":   {".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt", ".pages"},
		"videos":      {".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v"},
		"audio":       {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"},
		"archives":    {".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz"},
		"code":        {".go", ".py", ".js", ".html", ".css", ".java", ".cpp", ".c", ".h"},
		"data":        {".json", ".xml", ".csv", ".sql", ".db", ".sqlite"},
		"executables": {".exe", ".msi", ".deb", ".rpm", ".dmg", ".app"},
	}

	for category, extensions := range categories {
		for _, validExt := range extensions {
			if ext == validExt {
				return category
			}
		}
	}

	return "misc"
}

func parseCLI() CLI {
	var cli CLI

	flag.BoolVar(&cli.Categorize, "categorize", true, "Categorize files by type (default: true)")
	flag.BoolVar(&cli.Categorize, "c", true, "Categorize files by type (shorthand)")
	flag.BoolVar(&cli.Help, "help", false, "Show help")
	flag.BoolVar(&cli.Help, "h", false, "Show help (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command> [date]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  stash     Stash files in current directory\n")
		fmt.Fprintf(os.Stderr, "  unstash   Unstash files (optionally specify date)\n")
		fmt.Fprintf(os.Stderr, "  list      List available stashes\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s stash                    # Stash with categorization\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -categorize=false stash  # Stash without categorization\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s unstash 2024-01-15      # Unstash specific date\n", os.Args[0])
	}

	flag.Parse()

	if cli.Help {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	cli.Command = args[0]
	if len(args) > 1 {
		cli.Date = args[1]
	}

	return cli
}

func getTodayStashFolder() string {
	return "gostash-" + time.Now().Format("2006-01-02")
}

func saveConfig(config StashConfig) error {
	file, err := os.OpenFile("./.gostash.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}

func loadConfig() (StashConfig, error) {
	config := StashConfig{
		FilesByDate: make(map[string]StashEntry),
	}

	if _, err := os.Stat("./.gostash.json"); os.IsNotExist(err) {
		// Create empty config file
		return config, saveConfig(config)
	}

	file, err := os.Open("./.gostash.json")
	if err != nil {
		return config, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("error decoding JSON: %w", err)
	}

	return config, nil
}

func isBlacklisted(name string, stashFolder string) bool {
	blacklist := []string{".gostash.json", stashFolder}
	for _, item := range blacklist {
		if item == name {
			return true
		}
	}
	return false
}

func createStashStructure(baseFolder string, categorize bool, categories map[string]bool) error {
	if err := os.MkdirAll(baseFolder, 0755); err != nil {
		return fmt.Errorf("error creating base folder: %w", err)
	}

	if categorize {
		for category := range categories {
			categoryPath := filepath.Join(baseFolder, category)
			if err := os.MkdirAll(categoryPath, 0755); err != nil {
				return fmt.Errorf("error creating category folder %s: %w", category, err)
			}
		}
	}

	return nil
}

func stashFiles(config *StashConfig, categorize bool) error {
	stashFolder := getTodayStashFolder()

	// First pass: determine what categories we need
	categories := make(map[string]bool)
	var filesToStash []os.DirEntry
	var foldersToStash []os.DirEntry

	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	for _, entry := range entries {
		if isBlacklisted(entry.Name(), stashFolder) {
			continue
		}

		if entry.IsDir() {
			foldersToStash = append(foldersToStash, entry)
		} else {
			filesToStash = append(filesToStash, entry)
			if categorize {
				category := getFileCategory(entry.Name())
				categories[category] = true
			}
		}
	}

	if len(filesToStash) == 0 && len(foldersToStash) == 0 {
		fmt.Println("No files or folders to stash")
		return nil
	}

	// Create stash structure
	if err := createStashStructure(stashFolder, categorize, categories); err != nil {
		return err
	}

	// Initialize stash entry
	entry := StashEntry{
		Files:   make([]string, 0),
		Folders: make([]string, 0),
	}

	if categorize {
		entry.Mode = ModeCategorize
		entry.Categories = make(map[string][]string)
	} else {
		entry.Mode = ModeSimple
	}

	// Stash files
	for _, file := range filesToStash {
		var destPath string

		if categorize {
			category := getFileCategory(file.Name())
			destPath = filepath.Join(stashFolder, category, file.Name())
			entry.Categories[category] = append(entry.Categories[category], destPath)
		} else {
			destPath = filepath.Join(stashFolder, file.Name())
		}

		if err := os.Rename(file.Name(), destPath); err != nil {
			fmt.Printf("Warning: failed to move file %s: %v\n", file.Name(), err)
			continue
		}

		entry.Files = append(entry.Files, destPath)
	}

	// Stash folders
	for _, folder := range foldersToStash {
		destPath := filepath.Join(stashFolder, folder.Name())

		if err := os.Rename(folder.Name(), destPath); err != nil {
			fmt.Printf("Warning: failed to move folder %s: %v\n", folder.Name(), err)
			continue
		}

		entry.Folders = append(entry.Folders, destPath)
	}

	config.FilesByDate[stashFolder] = entry

	fmt.Printf("Stashed %d files and %d folders", len(entry.Files), len(entry.Folders))
	if categorize {
		fmt.Printf(" (categorized into %d categories)", len(entry.Categories))
	}
	fmt.Println()

	return saveConfig(*config)
}

func unstashFiles(config *StashConfig, dateFolder string) error {
	if dateFolder == "" {
		dateFolder = getTodayStashFolder()
	} else if !strings.HasPrefix(dateFolder, "gostash-") {
		dateFolder = "gostash-" + dateFolder
	}

	entry, found := config.FilesByDate[dateFolder]
	if !found {
		return fmt.Errorf("no stash found for date: %s", dateFolder)
	}

	if len(entry.Files) == 0 && len(entry.Folders) == 0 {
		return fmt.Errorf("no files or folders to unstash for date: %s", dateFolder)
	}

	// Unstash files
	for _, filePath := range entry.Files {
		fileName := filepath.Base(filePath)
		destPath := "./" + fileName

		if err := os.Rename(filePath, destPath); err != nil {
			fmt.Printf("Warning: failed to restore file %s: %v\n", filePath, err)
			continue
		}
	}

	// Unstash folders
	for _, folderPath := range entry.Folders {
		folderName := filepath.Base(folderPath)
		destPath := "./" + folderName

		if err := os.Rename(folderPath, destPath); err != nil {
			fmt.Printf("Warning: failed to restore folder %s: %v\n", folderPath, err)
			continue
		}
	}

	// Clean up empty stash folder
	if err := os.RemoveAll(dateFolder); err != nil {
		fmt.Printf("Warning: failed to remove stash folder %s: %v\n", dateFolder, err)
	}

	// Remove from config
	delete(config.FilesByDate, dateFolder)

	fmt.Printf("Unstashed %d files and %d folders from %s\n",
		len(entry.Files), len(entry.Folders), dateFolder)

	return saveConfig(*config)
}

func listStashes(config StashConfig) {
	if len(config.FilesByDate) == 0 {
		fmt.Println("No stashes found")
		return
	}

	fmt.Println("Available stashes:")
	for date, entry := range config.FilesByDate {
		fmt.Printf("  %s: %d files, %d folders",
			strings.TrimPrefix(date, "gostash-"), len(entry.Files), len(entry.Folders))

		if entry.Mode == ModeCategorize && len(entry.Categories) > 0 {
			fmt.Printf(" (categorized: %s)", strings.Join(getKeys(entry.Categories), ", "))
		}
		fmt.Println()
	}
}

func getKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	cli := parseCLI()

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	switch cli.Command {
	case "stash":
		if err := stashFiles(&config, cli.Categorize); err != nil {
			fmt.Printf("Error stashing files: %v\n", err)
			os.Exit(1)
		}
	case "unstash":
		if err := unstashFiles(&config, cli.Date); err != nil {
			fmt.Printf("Error unstashing files: %v\n", err)
			os.Exit(1)
		}
	case "list":
		listStashes(config)
	default:
		fmt.Printf("Unknown command: %s\n", cli.Command)
		flag.Usage()
		os.Exit(1)
	}
}
