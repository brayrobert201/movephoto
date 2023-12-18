# MovePhoto Script

This script is designed to organize photos and videos from specified directories into a destination directory. The organization is done based on the date the photos or videos were taken.

## Configuration

The script is configured using a `config.yaml` file. An example configuration file `config.yaml.example` is provided. Copy this file to `config.yaml` and modify it to suit your needs.

The `config.yaml` file has the following fields:

- `defaultWatchDir`: The default directory to watch for new photos and videos.
- `additionalWatchDirs`: An array of additional directories to watch for new photos and videos.
- `defaultDestinationDir`: The directory where photos and videos will be moved to.
- `imageExtensions`: An array of file extensions to consider as images.
- `videoExtensions`: An array of file extensions to consider as videos.
- `bannedExtensions`: An array of file extensions to ignore and delete.
- `lockFilePath`: The path to the lock file used to prevent multiple instances of the script from running at the same time.

## Setting up a Cron Job

To run this script every hour, you can set up a cron job. Open your crontab file with the command `crontab -e` and add the following line:

```
0 * * * * /path/to/script/movephoto.go
```

Replace `/path/to/script/movephoto.go` with the actual path to the `movephoto.go` script. This will run the script at the start of every hour.
