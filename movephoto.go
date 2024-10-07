package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	exiftool "github.com/barasher/go-exiftool"
	"gopkg.in/yaml.v2"
)

// WatchDir represents a directory to watch along with the action to perform and optional prefixes
type WatchDir struct {
	Path          string   `yaml:"path"`
	Action        string   `yaml:"action"`        // "move" or "copy"
	IncludePrefix []string `yaml:"includePrefix"` // List of prefixes to include (optional)
}

// Config holds the configuration data
type Config struct {
	WatchDirs             []WatchDir `yaml:"watchDirs"`
	DefaultDestinationDir string     `yaml:"defaultDestinationDir"`
	ImageExtensions       []string   `yaml:"imageExtensions"`
	VideoExtensions       []string   `yaml:"videoExtensions"`
	BannedExtensions      []string   `yaml:"bannedExtensions"`
	LockFilePath          string     `yaml:"lockFilePath"`
}

// Global variable to keep track of processed files
var processedFiles map[string]struct{}

// Path to the processed files list
var processedFilesPath string

func loadConfig() Config {
	if _, err := os.Stat(*configFilePath); os.IsNotExist(err) {
		fmt.Printf("%s does not exist. Please create it and run the program again.\n", *configFilePath)
		os.Exit(0)
	}

	config := Config{}
	data, err := os.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}

var (
	pollingInterval = flag.Int("polling-interval", 30, "Polling interval in seconds for checking new files in the watch directories")
	debug           = flag.Bool("debug", false, "Enable debug output")
	watch           = flag.Bool("watch", false, "Enable regular scanning of the source directories")
	configFilePath  = flag.String("config", "/etc/movephoto_config.yml", "Path to the configuration file")
	minFileSize     = int64(102400) // Minimum file size in bytes (100KB)
)

func main() {
	flag.Parse() // Parse the command-line flags

	if *debug {
		log.Printf("[%s] Debug mode enabled\n", currentTime())
	}

	config := loadConfig()

	// Add .heic to the list of image extensions
	config.ImageExtensions = append(config.ImageExtensions, ".heic")

	// Set the path for processed_files.txt under the destination directory
	processedFilesPath = filepath.Join(config.DefaultDestinationDir, "processed_files.txt")

	// Load processed files
	processedFiles = loadProcessedFiles(processedFilesPath)

	if !*watch {
		if *debug {
			log.Printf("[%s] Performing a single scan...\n", currentTime())
		}
		processFiles(config)
	} else {
		if *debug {
			log.Printf("[%s] Starting regular scanning of source directories...\n", currentTime())
		}
		ticker := time.NewTicker(time.Duration(*pollingInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if *debug {
				log.Printf("[%s] Polling for new files...\n", currentTime())
			}
			processFiles(config)
		}
	}
}

func processFiles(config Config) {
	for _, watchDir := range config.WatchDirs {
		purge_unwanted(watchDir.Path, config.BannedExtensions)
		switch watchDir.Action {
		case "move":
			move_photos(watchDir.Path, config.DefaultDestinationDir, config.ImageExtensions)
			move_videos(watchDir.Path, config.DefaultDestinationDir, config.VideoExtensions)
		case "copy":
			copy_photos(watchDir.Path, config.DefaultDestinationDir, config.ImageExtensions, watchDir.IncludePrefix)
			copy_videos(watchDir.Path, config.DefaultDestinationDir, config.VideoExtensions, watchDir.IncludePrefix)
		default:
			log.Printf("Unknown action %s for watch directory %s", watchDir.Action, watchDir.Path)
		}
	}

	// No need to save processed files here since it's being updated during processing
}

func purge_unwanted(watch_dir string, banned_extensions []string) error {
	files, err := os.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, ext := range banned_extensions {
			if strings.ToLower(filepath.Ext(file.Name())) == ext {
				os.Remove(filepath.Join(watch_dir, file.Name()))
			}
		}
	}
	return nil
}

func move_files(watch_dir string, destination_dir string, extensions []string, get_destination_dir func(filePath string, file os.FileInfo) string) error {
	files, err := os.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			continue // Skip if we can't get file info
		}

		if !info.Mode().IsRegular() {
			continue // Skip non-regular files
		}

		if !hasExtension(info.Name(), extensions) {
			continue
		}

		// Skip files smaller than the minimum size
		if info.Size() < minFileSize {
			if *debug {
				log.Printf("[%s] Skipping file (too small): %s\n", currentTime(), info.Name())
			}
			continue
		}

		sourcePath := filepath.Join(watch_dir, info.Name())
		full_destination_dir := get_destination_dir(sourcePath, info)
		full_destination := filepath.Join(full_destination_dir, info.Name())
		if _, err := os.Stat(full_destination_dir); os.IsNotExist(err) {
			os.MkdirAll(full_destination_dir, os.ModePerm)
		}
		if _, err := os.Stat(full_destination); os.IsNotExist(err) {
			err := copyAndVerify(sourcePath, full_destination)
			if err != nil {
				log.Printf("[%s] Failed to move file: %s\n", currentTime(), err)
			} else {
				// Delete the source file after successful copy and verification
				err = os.Remove(sourcePath)
				if err != nil {
					log.Printf("[%s] Failed to delete source file: %s\n", currentTime(), err)
				} else {
					log.Printf("[%s] Moved file: %s to %s\n", currentTime(), sourcePath, full_destination)
				}
			}
		}
	}
	return nil
}

func copy_files(watch_dir string, destination_dir string, extensions []string, includePrefix []string, get_destination_dir func(filePath string, file os.FileInfo) string) error {
	files, err := os.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			continue // Skip if we can't get file info
		}

		if !info.Mode().IsRegular() {
			continue // Skip non-regular files
		}

		if !hasExtension(info.Name(), extensions) {
			continue
		}

		// Skip files smaller than the minimum size
		if info.Size() < minFileSize {
			if *debug {
				log.Printf("[%s] Skipping file (too small): %s\n", currentTime(), info.Name())
			}
			continue
		}

		// Apply includePrefix filtering if includePrefix is not empty
		if len(includePrefix) > 0 && !hasPrefix(info.Name(), includePrefix) {
			continue // Skip the file
		}

		filePath := filepath.Join(watch_dir, info.Name())

		if _, ok := processedFiles[filePath]; ok {
			// File has already been processed
			continue
		}

		full_destination_dir := get_destination_dir(filePath, info)
		full_destination := filepath.Join(full_destination_dir, info.Name())
		if _, err := os.Stat(full_destination_dir); os.IsNotExist(err) {
			os.MkdirAll(full_destination_dir, os.ModePerm)
		}
		if _, err := os.Stat(full_destination); os.IsNotExist(err) {
			err := copyAndVerify(filePath, full_destination)
			if err != nil {
				log.Printf("[%s] Failed to copy file: %s\n", currentTime(), err)
			} else {
				log.Printf("[%s] Copied file: %s to %s\n", currentTime(), filePath, full_destination)
				// Add to processed files and update the file immediately
				processedFiles[filePath] = struct{}{}
				appendToProcessedFiles(processedFilesPath, filePath)
			}
		} else {
			// Destination file already exists
			processedFiles[filePath] = struct{}{}
			// Ensure it's in the processed files list
			appendToProcessedFiles(processedFilesPath, filePath)
		}
	}
	return nil
}

func move_photos(watch_dir string, destination_dir string, image_extensions []string) error {
	return move_files(watch_dir, destination_dir, image_extensions, func(filePath string, file os.FileInfo) string {
		var date_taken time.Time
		var err error

		// Attempt to get date from EXIF data
		date_taken, err = getPhotoTimestamp(filePath)
		if err != nil {
			// If EXIF data is not available, attempt to parse date from filename
			date_taken, err = parseDateFromFilename(file.Name())
			if err != nil {
				// Skip the file if neither EXIF data nor filename date is available
				fmt.Printf("Skipping file %s: no valid date found\n", filePath)
				return ""
			}
		}

		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%02d - %s", month_taken, month_name),
			fmt.Sprintf("%04d-%02d-%02d", year_taken, int(month_taken), day_taken),
		)
	})
}

func move_videos(watch_dir string, destination_dir string, video_extensions []string) error {
	// For videos, we'll continue to use the file modification time
	return move_files(watch_dir, destination_dir, video_extensions, func(filePath string, file os.FileInfo) string {
		date_taken := file.ModTime()
		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%02d - %s", month_taken, month_name),
			fmt.Sprintf("%04d-%02d-%02d", year_taken, int(month_taken), day_taken),
		)
	})
}

func copy_photos(watch_dir string, destination_dir string, image_extensions []string, includePrefix []string) error {
	return copy_files(watch_dir, destination_dir, image_extensions, includePrefix, func(filePath string, file os.FileInfo) string {
		date_taken, err := getPhotoTimestamp(filePath)
		if err != nil {
			// Fallback to file modification time
			date_taken = file.ModTime()
		}
		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%02d - %s", month_taken, month_name),
			fmt.Sprintf("%04d-%02d-%02d", year_taken, int(month_taken), day_taken),
		)
	})
}

func copy_videos(watch_dir string, destination_dir string, video_extensions []string, includePrefix []string) error {
	// For videos, we'll continue to use the file modification time
	return copy_files(watch_dir, destination_dir, video_extensions, includePrefix, func(filePath string, file os.FileInfo) string {
		date_taken := file.ModTime()
		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%02d - %s", month_taken, month_name),
			fmt.Sprintf("%04d-%02d-%02d", year_taken, int(month_taken), day_taken),
		)
	})
}

// getPhotoTimestamp extracts the DateTimeOriginal from the photo's EXIF data using ExifTool
func getPhotoTimestamp(filePath string) (time.Time, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return time.Time{}, fmt.Errorf("Error when creating Exiftool: %v", err)
	}
	defer et.Close()

	fileInfos := et.ExtractMetadata(filePath)
	if len(fileInfos) == 0 {
		return time.Time{}, fmt.Errorf("No metadata extracted for file: %s", filePath)
	}

	fi := fileInfos[0]
	if fi.Err != nil {
		return time.Time{}, fi.Err
	}

	dateTimeOriginal, ok := fi.Fields["DateTimeOriginal"].(string)
	if !ok {
		return time.Time{}, fmt.Errorf("DateTimeOriginal not found in EXIF data")
	}

	dateFormats := []string{
		"2006:01:02 15:04:05",
		"2006:01:02 15:04:05-07:00",
		"2006:01:02 15:04:05Z07:00",
	}

	var parsedTime time.Time
	var parseErr error
	for _, format := range dateFormats {
		parsedTime, parseErr = time.Parse(format, dateTimeOriginal)
		if parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		return time.Time{}, fmt.Errorf("Error parsing DateTimeOriginal: %v", parseErr)
	}

	return parsedTime, nil
}

// hasExtension checks if the filename has one of the specified extensions
func hasExtension(filename string, extensions []string) bool {
	fileExt := strings.ToLower(filepath.Ext(filename))
	for _, ext := range extensions {
		if fileExt == ext {
			return true
		}
	}
	return false
}

// hasPrefix checks if the filename starts with any of the specified prefixes
func hasPrefix(filename string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

// currentTime returns the current time as a formatted string with timezone.
func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}

// copyAndVerify copies a file from src to dst and verifies the integrity
func copyAndVerify(src, dst string) error {
	// Copy the file
	err := copyFile(src, dst)
	if err != nil {
		return err
	}

	// Check the file size of the destination
	destInfo, err := os.Stat(dst)
	if err != nil {
		// Delete the incomplete destination file
		os.Remove(dst)
		return fmt.Errorf("error stating destination file: %v", err)
	}

	if destInfo.Size() == 0 {
		// Delete the empty destination file
		os.Remove(dst)
		return fmt.Errorf("destination file %s has zero size after copy", dst)
	}

	// Compute checksums of source and destination
	srcChecksum, err := computeFileChecksum(src)
	if err != nil {
		// Delete the destination file if checksum fails
		os.Remove(dst)
		return fmt.Errorf("error computing checksum of source file: %v", err)
	}

	destChecksum, err := computeFileChecksum(dst)
	if err != nil {
		// Delete the destination file if checksum fails
		os.Remove(dst)
		return fmt.Errorf("error computing checksum of destination file: %v", err)
	}

	if srcChecksum != destChecksum {
		// Delete the destination file if checksums don't match
		os.Remove(dst)
		return fmt.Errorf("checksum mismatch between source and destination files for %s", src)
	}

	return nil
}

// Function to parse date from filename
func parseDateFromFilename(filename string) (time.Time, error) {
	// Define regex patterns for iPhone and Pixel filenames
	iphonePattern := regexp.MustCompile(`(\d{8})_\d{9}_iOS`)
	pixelPattern := regexp.MustCompile(`PXL_(\d{8})_\d{9}`)

	// Check for iPhone pattern
	if matches := iphonePattern.FindStringSubmatch(filename); matches != nil {
		return time.Parse("20060102", matches[1])
	}

	// Check for Pixel pattern
	if matches := pixelPattern.FindStringSubmatch(filename); matches != nil {
		return time.Parse("20060102", matches[1])
	}

	return time.Time{}, fmt.Errorf("no date found in filename")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	// Check if destination file exists
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination file %s already exists", dst)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, sourceFileStat.Mode())
	if err != nil {
		return err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	if err != nil {
		// Delete the incomplete destination file
		os.Remove(dst)
		return err
	}

	if nBytes == 0 {
		// Delete the empty destination file
		os.Remove(dst)
		return fmt.Errorf("copied zero bytes from %s to %s", src, dst)
	}

	return nil
}

// computeFileChecksum computes the MD5 checksum of a file
func computeFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

// loadProcessedFiles loads the list of processed files from a file
func loadProcessedFiles(filePath string) map[string]struct{} {
	processedFiles := make(map[string]struct{})
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return processedFiles
		}
		log.Fatalf("Error opening processed files list: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		processedFiles[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading processed files list: %v", err)
	}
	return processedFiles
}

// appendToProcessedFiles appends a file path to the processed files list
func appendToProcessedFiles(filePath, processedFile string) {
	// Ensure the directory exists
	processedDir := filepath.Dir(filePath)
	if _, err := os.Stat(processedDir); os.IsNotExist(err) {
		os.MkdirAll(processedDir, os.ModePerm)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening processed files list for appending: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(processedFile + "\n"); err != nil {
		log.Fatalf("Error writing to processed files list: %v", err)
	}
}
