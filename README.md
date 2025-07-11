# gostash

A simple CLI tool to stash and unstash files in your working directory by date. Useful for quickly clearing your workspace and restoring files later.

## Features

- **Stash**: Moves all files (except `.gostash.json` and today's stash folder) into a dated folder.
- **Unstash**: Restores files from a specific date's stash folder back to the root directory.
- Maintains a `.gostash.json` config file to track stashed files by date.

## Setup

1. **Install Go** (if you don't have it):
   - Download and install from [https://golang.org/dl/](https://golang.org/dl/)
2. **Clone this repository**:
   ```sh
   git clone https://github.com/Skynse/gostash.git
   cd gostash
   ```
3. **(Optional) Build the binary**:
   ```sh
   go build gostash.go
   ```
   This will create a `gostash.exe` (Windows) or `gostash` (Linux/Mac) binary in the current directory.

## Usage

You can run gostash directly with Go or use the built binary:

```sh
# Run with Go
go run gostash.go <command> [date]

# Or run the built binary
./gostash <command> [date]
```

### Commands

- `stash`: Moves all files in the current directory into a folder named `gostash-YYYY-MM-DD`.
- `unstash [YYYY-MM-DD]`: Moves files from the specified date's stash folder back to the root directory. If no date is provided, uses today's date.

### Examples

Stash today's files:
```sh
go run gostash.go stash
```

Unstash files from today:
```sh
go run gostash.go unstash
```

Unstash files from a specific date:
```sh
go run gostash.go unstash 2024-06-01
```

## How it works

- On `stash`, all files except `.gostash.json` and the current day's stash folder are moved into a new folder named `gostash-YYYY-MM-DD`.
- The `.gostash.json` file keeps track of which files were stashed on which date.
- On `unstash`, files are moved back to the root directory and the entry is removed from the config.