package main

// Importing packages in Go is similar to Python's import statements.
// For example, "fmt" is similar to Python's built-in "print" function.
import (
	"fmt" // Similar to Python's print function
	"image/jpeg" // Used for handling JPEG images
	"io/ioutil" // Provides I/O utility functions
	"log" // Used for logging
	"os" // Provides a platform-independent interface to operating system functionality
	"path/filepath" // Provides utility functions for manipulating filename paths
	"strings" // Provides string manipulation functions
	"time" // Provides functionality for measuring and displaying time
	"bufio" // Used for buffered I/O
	"gopkg.in/yaml.v2" // Used for handling YAML files
)

// In Go, a struct is a collection of fields, similar to a class in Python.
// Here, Config is a struct that holds configuration data.
type Config struct {
	DefaultWatchDir       string   `yaml:"defaultWatchDir"` // Similar to Python's class attribute
	AdditionalWatchDirs   []string `yaml:"additionalWatchDirs"` // Similar to Python's list
	DefaultDestinationDir string   `yaml:"defaultDestinationDir"` // Similar to Python's class attribute
	ImageExtensions       []string `yaml:"imageExtensions"` // Similar to Python's list
	VideoExtensions       []string `yaml:"videoExtensions"` // Similar to Python's list
	BannedExtensions      []string `yaml:"bannedExtensions"` // Similar to Python's list
	LockFilePath          string   `yaml:"lockFilePath"` // Similar to Python's class attribute
}

// In Go, a function is defined using the "func" keyword, similar to Python's "def" keyword.
// Here, loadConfig is a function that loads the configuration data from a YAML file.
func loadConfig() Config { // Similar to Python's def keyword
	// The if statement in Go is similar to Python's if statement.
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) { // Similar to Python's if statement
		fmt.Println("config.yaml does not exist. Copying config.yaml.example to config.yaml...")
		// Error handling in Go is done using if statements, similar to Python's try-except blocks.
		err := os.Link("config.yaml.example", "config.yaml") // Similar to Python's os.link function
		if err != nil { // Similar to Python's if statement
			log.Fatalf("error: %v", err) // Similar to Python's print function
		}
		fmt.Println("config.yaml created. Please edit it and run the program again.")
		os.Exit(0) // Similar to Python's sys.exit function
	}

	// The := operator in Go is a shorthand for declaring and initializing a variable, similar to Python's = operator.
	config := Config{} // Similar to Python's class instantiation
	data, err := ioutil.ReadFile("config.yaml") // Similar to Python's open function
	if err != nil { // Similar to Python's if statement
		log.Fatalf("error: %v", err) // Similar to Python's print function
	}
	err = yaml.Unmarshal([]byte(data), &config) // Similar to Python's yaml.load function
	if err != nil { // Similar to Python's if statement
		log.Fatalf("error: %v", err) // Similar to Python's print function
	}
	return config // Similar to Python's return statement
}

func main() {
	fmt.Println("Starting movephoto...")
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		fmt.Println("config.yaml does not exist. Would you like to copy config.yaml.example to config.yaml? (y/n)")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if response == "y\n" {
			os.Link("config.yaml.example", "config.yaml")
		}
	}
	config := loadConfig()

	for {
		if _, err := os.Stat(config.LockFilePath); os.IsNotExist(err) {
			break
		} else {
			info, _ := os.Stat(config.LockFilePath)
			if time.Since(info.ModTime()).Hours() > 24 {
				os.Remove(config.LockFilePath)
				break
			}
			fmt.Println("Waiting for lock file to be removed...")
			time.Sleep(30 * time.Second)
		}
	}

	os.Create(config.LockFilePath)
	defer os.Remove(config.LockFilePath)

	fmt.Println("Starting to move photos and videos...")
	for _, watchDir := range config.AdditionalWatchDirs {
		purge_unwanted(watchDir, config.BannedExtensions)
		move_photos(watchDir, config.DefaultDestinationDir, config.ImageExtensions)
		move_videos(watchDir, config.DefaultDestinationDir, config)
	}
}

func purge_unwanted(watch_dir string, banned_extensions []string) error {
	files, err := ioutil.ReadDir(watch_dir)
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

func move_files(watch_dir string, destination_dir string, extensions []string, get_destination_dir func(file os.FileInfo) string) error {
	files, err := ioutil.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, ext := range extensions {
			if strings.ToLower(filepath.Ext(file.Name())) == ext {
				full_destination_dir := get_destination_dir(file)
				full_destination := filepath.Join(full_destination_dir, file.Name())
				if _, err := os.Stat(full_destination_dir); os.IsNotExist(err) {
					os.MkdirAll(full_destination_dir, os.ModePerm)
				}
				if _, err := os.Stat(full_destination); os.IsNotExist(err) {
					os.Rename(filepath.Join(watch_dir, file.Name()), full_destination)
				}
			}
		}
	}
	return nil
}

func move_photos(watch_dir string, destination_dir string, image_extensions []string) error {
	return move_files(watch_dir, destination_dir, image_extensions, func(file os.FileInfo) string {
		imgFile, err := os.Open(filepath.Join(watch_dir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		_, err = jpeg.DecodeConfig(imgFile)
		if err != nil {
			log.Fatal(err)
		}
		info, err := imgFile.Stat()
		if err != nil {
			log.Fatal(err)
		}
		date_taken := info.ModTime()
		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%d - %s", month_taken, month_name),
			fmt.Sprintf("%d-%d-%d", year_taken, month_taken, day_taken),
		)
	})
}

func move_videos(watch_dir string, destination_dir string, video_extensions []string) error {
	return move_files(watch_dir, destination_dir, video_extensions, func(file os.FileInfo) string {
		info, err := os.Stat(filepath.Join(watch_dir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		date_taken := info.ModTime()
		year_taken, month_taken, day_taken := date_taken.Date()
		month_name := month_taken.String()
		return filepath.Join(
			destination_dir,
			fmt.Sprintf("%d", year_taken),
			fmt.Sprintf("%d - %s", month_taken, month_name),
			fmt.Sprintf("%d-%d-%d", year_taken, month_taken, day_taken),
		)
	})
}
