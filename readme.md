# MovePhoto Script

This script is designed to organize photos and videos from specified directories into a destination directory. The organization is done based on the date the photos or videos were taken.

## Configuration

The script is configured using a `config.yaml` file. If a `config.yaml` file doesn't exist when the script is run, it will offer to copy the contents of `config.yaml.example` as a template. 

An example configuration file `config.yaml.example` is provided. Copy this file to `config.yaml` and modify it to suit your needs. If `config.yaml` needed to be created, the script will exit with a message asking the user to review the contents of `config.yaml` before running the script again.

The `config.yaml` file has the following fields:

- `defaultWatchDir`: The default directory to watch for new photos and videos.
- `additionalWatchDirs`: An array of additional directories to watch for new photos and videos.
- `defaultDestinationDir`: The directory where photos and videos will be moved to.
- `imageExtensions`: An array of file extensions to consider as images.
- `videoExtensions`: An array of file extensions to consider as videos.
- `bannedExtensions`: An array of file extensions to ignore and delete.
- `lockFilePath`: The path to the lock file used to prevent multiple instances of the script from running at the same time.

## How the Script Works

The script uses the metadata of the photo and video files to decide where to move them. Specifically, it uses the modification date of the files. It organizes the files into directories based on the year, month, and day the files were last modified.

## Setting up a Cron Job

To run this script every hour, you can set up a cron job. Here's how:

1. Open your crontab file with the command `crontab -e`.
2. Add the following line to the file:

```
0 * * * * /path/to/script/movephoto.go
```

Replace `/path/to/script/movephoto.go` with the actual path to the `movephoto.go` script. This will run the script at the start of every hour.

Remember to save and close the file after making these changes.

## Setting up a Go Build

If you are a developer who hasn't used Go, you will need to set up a Go build. Here's how:

1. Install Go on your machine. You can download it from the [official Go website](https://golang.org/dl/).
2. Clone this repository to your local machine.
3. Navigate to the directory containing the `movephoto.go` file.
4. Run the command `go build`. This will compile the Go code into an executable file.

## Resolving Missing go.sum Entry Error

If you encounter an error message like `missing go.sum entry for module providing package gopkg.in/yaml.v2 (imported by movephoto); to add: go get movephoto`, it means that the `go.sum` file is missing an entry for the `gopkg.in/yaml.v2` package. This package is required by the `movephoto` module.

To resolve this issue, run the command `go get movephoto` in your terminal. This command will update your `go.sum` file with the required dependencies for the `movephoto` module.
