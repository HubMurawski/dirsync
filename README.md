# dirsync

## Installation

### Prerequisites

- Go 1.18+

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/HubMurawski/dirsync.git
   cd dirsync
   ```

2. Build the application:
   ```
   go build
   ```
   For smaller binary you can use: 
   ```
   go build -ldflags "-s -w"
   ```

## Usage

Basic usage:
```
./dirsync <source> <destination>
```

With optional flags:
```
./dirsync [--delete-missing] [-src] <source> [-dst] <destination>
```

### Command Line Arguments

- `<source>`: Path to the source directory
- `<destination>`: Path to the target directory
- `--delete-missing`: Delete files from the target directory if they don't exist in the source directory

## Examples

1. Basic synchronization (copy and update only):
   ```
   ./dirsync /path/to/source /path/to/target
   ```

2. Synchronization with deletion of missing files:
   ```
   ./dirsync --delete-missing /path/to/source /path/to/target
   ```

## How It Works

1. The application scans both source and target directories recursively.
2. Files found in the source but not in the target are copied.
3. Files found in both locations are compared by file info, size and modification time. If they differ, the target file is updated.
4. If the `--delete-missing` flag is provided, files found in the target but not in the source are deleted.
