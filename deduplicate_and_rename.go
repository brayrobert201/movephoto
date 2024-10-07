package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	exif "github.com/rwcarlsen/goexif/exif"
)

var (
	targetDir   = flag.String("dir", "", "Path to the destination directory where duplicates need to be removed")
	dryRun      = flag.Bool("dry-run", false, "Perform a dry run without deleting or renaming any files")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	minFileSize = flag.Int64("min-size", 1024, "Minimum file size (in bytes) to process")
	trashDir    = flag.String("trash-dir", "", "Directory to move duplicates instead of deleting")
)

func main() {
	flag.Parse()

	if *targetDir == "" {
		log.Fatal("Please provide the path to the destination directory using the -dir flag.")
	}

	err := processDirectory(*targetDir)
	if err != nil {
		log.Fatalf("Error processing directories: %v", err)
	}
}

func processDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	if *verbose {
		fmt.Printf("Processing %d files in directory %s\n", len(files), dirPath)
	}

	// Map to store unique identifiers and associated file paths
	uniqueMap := make(map[string][]string)

	// Regular expression to match image files
	regex := regexp.MustCompile(`(?i)^(IMG.*)\.(jpg|jpeg|png|gif|bmp)$`)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Skip files smaller than the minimum size
		if file.Size() < *minFileSize {
			if *verbose {
				fmt.Printf("Skipping file (too small): %s\n", file.Name())
			}
			continue
		}

		fileName := file.Name()
		if *verbose {
			fmt.Printf("Checking file: %s\n", fileName)
		}

		if !regex.MatchString(fileName) {
			if *verbose {
				fmt.Printf("Skipping file (no match): %s\n", fileName)
			}
			continue
		}

		if *verbose {
			fmt.Printf("Processing file: %s\n", fileName)
		}

		filePath := filepath.Join(dirPath, fileName)

		uniqueID, err := computeUniqueID(filePath)
		if err != nil {
			log.Printf("Error computing unique ID for %s: %v", filePath, err)
			continue
		}

		if *verbose {
			fmt.Printf("File: %s, Unique ID: %s\n", filePath, uniqueID)
		}

		uniqueMap[uniqueID] = append(uniqueMap[uniqueID], filePath)
	}

	// Remove duplicates
	for uniqueID, filePaths := range uniqueMap {
		if len(filePaths) > 1 {
			// Sort file paths to determine which one to keep
			sortFilesByMetadata(filePaths)

			filesToDelete := filePaths[1:]

			if *verbose {
				fmt.Printf("Unique ID %s has %d duplicates\n", uniqueID, len(filesToDelete))
			}

			for _, filePath := range filesToDelete {
				if *dryRun {
					fmt.Printf("Would delete duplicate file: %s\n", filePath)
				} else {
					if *trashDir != "" {
						// Move file to trash directory
						err := moveToTrash(filePath, *trashDir)
						if err != nil {
							log.Printf("Error moving file %s to trash: %v", filePath, err)
						} else {
							fmt.Printf("Moved duplicate file to trash: %s\n", filePath)
						}
					} else {
						err := os.Remove(filePath)
						if err != nil {
							log.Printf("Error deleting file %s: %v", filePath, err)
						} else {
							fmt.Printf("Deleted duplicate file: %s\n", filePath)
						}
					}
				}
			}
			// Keep only the first file
			uniqueMap[uniqueID] = filePaths[:1]
		}
	}

	// Map to keep track of intended new filenames to avoid conflicts
	intendedNames := make(map[string]string)   // Map from current file path to intended new filename
	existingNames := make(map[string]struct{}) // Set of intended new filenames

	// Build intended new filenames for all files
	for _, filePaths := range uniqueMap {
		for _, filePath := range filePaths {
			newFileName, err := generateUniqueFileName(filePath, existingNames)
			if err != nil {
				log.Printf("Error generating new filename for %s: %v", filePath, err)
				continue
			}

			intendedNames[filePath] = newFileName
			existingNames[newFileName] = struct{}{}
		}
	}

	// Now perform the renaming
	for filePath, newFileName := range intendedNames {
		err := performRename(filePath, newFileName)
		if err != nil {
			log.Printf("Error renaming file %s: %v", filePath, err)
		}
	}

	return nil
}

func computeUniqueID(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Attempt to read EXIF data
	x, err := exif.Decode(file)
	if err != nil {
		// If EXIF data can't be read, fall back to computing checksum
		file.Seek(0, io.SeekStart)
		data, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}
		return computeChecksum(data), nil
	}

	// Extract fields that uniquely identify the photo
	make, _ := x.Get("Make")
	model, _ := x.Get("Model")
	dateTimeOriginal, _ := x.Get("DateTimeOriginal")
	lensModel, _ := x.Get("LensModel")
	imageUniqueID, _ := x.Get("ImageUniqueID")
	serialNumber, _ := x.Get("BodySerialNumber")

	// Get string values, handling possible nil pointers
	makeStr := ""
	if make != nil {
		makeStr, _ = make.StringVal()
	}
	modelStr := ""
	if model != nil {
		modelStr, _ = model.StringVal()
	}
	dateTimeOriginalStr := ""
	if dateTimeOriginal != nil {
		dateTimeOriginalStr, _ = dateTimeOriginal.StringVal()
	}
	lensModelStr := ""
	if lensModel != nil {
		lensModelStr, _ = lensModel.StringVal()
	}
	imageUniqueIDStr := ""
	if imageUniqueID != nil {
		imageUniqueIDStr, _ = imageUniqueID.StringVal()
	}
	serialNumberStr := ""
	if serialNumber != nil {
		serialNumberStr, _ = serialNumber.StringVal()
	}

	// Concatenate metadata fields
	uniqueString := fmt.Sprintf("%v|%v|%v|%v|%v|%v", makeStr, modelStr, dateTimeOriginalStr, lensModelStr, imageUniqueIDStr, serialNumberStr)

	if *verbose {
		fmt.Printf("Metadata for %s: %s\n", filePath, uniqueString)
	}

	// Generate MD5 hash of the unique string
	hash := md5.Sum([]byte(uniqueString))
	uniqueID := hex.EncodeToString(hash[:])

	return uniqueID, nil
}

func computeChecksum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func generateUniqueFileName(filePath string, existingNames map[string]struct{}) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return "", fmt.Errorf("file %s has no extension", filePath)
	}

	timestamp, err := getPhotoTimestamp(filePath)
	if err != nil {
		return "", err
	}

	// Start with base filename
	baseName := fmt.Sprintf("IMG_%s", timestamp.Format("20060102_150405"))

	// Counter for duplicate filenames
	counter := 0
	var newFileName string

	for {
		if counter == 0 {
			newFileName = fmt.Sprintf("%s%s", baseName, ext)
		} else {
			newFileName = fmt.Sprintf("%s_%d%s", baseName, counter, ext)
		}

		// Check if the filename already exists in existingNames map
		if _, exists := existingNames[newFileName]; !exists {
			// Filename is unique
			break
		}

		counter++
	}

	// Replace any invalid characters (e.g., ':' on Windows)
	newFileName = strings.ReplaceAll(newFileName, ":", "")
	return newFileName, nil
}

func performRename(filePath, newFileName string) error {
	dir := filepath.Dir(filePath)
	newFilePath := filepath.Join(dir, newFileName)

	currentFileName := filepath.Base(filePath)
	if currentFileName == newFileName {
		// File already has the correct name
		if *verbose {
			fmt.Printf("File already has the correct name: %s\n", filePath)
		}
		return nil
	}

	if *dryRun {
		fmt.Printf("Would rename file: %s -> %s\n", filePath, newFilePath)
		return nil
	}

	err := os.Rename(filePath, newFilePath)
	if err != nil {
		return err
	}
	fmt.Printf("Renamed file: %s -> %s\n", filePath, newFilePath)
	return nil
}

func moveToTrash(filePath, trashDir string) error {
	// Ensure trash directory exists
	if _, err := os.Stat(trashDir); os.IsNotExist(err) {
		err = os.MkdirAll(trashDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fileName := filepath.Base(filePath)
	destination := filepath.Join(trashDir, fileName)

	// Avoid overwriting files in the trash directory
	for i := 1; ; i++ {
		if _, err := os.Stat(destination); os.IsNotExist(err) {
			break
		}
		destination = filepath.Join(trashDir, fmt.Sprintf("%s_%d", fileName, i))
	}

	return os.Rename(filePath, destination)
}

func sortFilesByMetadata(filePaths []string) {
	sort.Slice(filePaths, func(i, j int) bool {
		t1, err1 := getPhotoTimestamp(filePaths[i])
		t2, err2 := getPhotoTimestamp(filePaths[j])
		if err1 != nil || err2 != nil {
			return filePaths[i] < filePaths[j]
		}
		return t1.Before(t2)
	})
}

func getPhotoTimestamp(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err == nil {
		dateTimeOriginal, err := x.DateTime()
		if err == nil {
			return dateTimeOriginal, nil
		}
	}

	// Fallback to file modification time
	info, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
