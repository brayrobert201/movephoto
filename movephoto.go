package main

import (
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"bufio"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DefaultWatchDir       string   `yaml:"defaultWatchDir"`
	AdditionalWatchDirs   []string `yaml:"additionalWatchDirs"`
	DefaultDestinationDir string   `yaml:"defaultDestinationDir"`
	ImageExtensions       []string `yaml:"imageExtensions"`
	VideoExtensions       []string `yaml:"videoExtensions"`
	BannedExtensions      []string `yaml:"bannedExtensions"`
	LockFilePath          string   `yaml:"lockFilePath"`
}

func loadConfig() Config {
	config := Config{}
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}

func main() {
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
			time.Sleep(30 * time.Second)
		}
	}

	os.Create(config.LockFilePath)
	defer os.Remove(config.LockFilePath)

	for _, watchDir := range config.AdditionalWatchDirs {
		purge_unwanted(watchDir, config.BannedExtensions)
		move_photos(watchDir, config.DefaultDestinationDir, config.ImageExtensions)
		move_videos(watchDir, config.DefaultDestinationDir)
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

func move_photos(watch_dir string, destination_dir string, image_extensions []string) error {
	files, err := ioutil.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, ext := range image_extensions {
			if strings.ToLower(filepath.Ext(file.Name())) == ext {
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
				full_destination_dir := filepath.Join(
					destination_dir,
					fmt.Sprintf("%d", year_taken),
					fmt.Sprintf("%d - %s", month_taken, month_name),
					fmt.Sprintf("%d-%d-%d", year_taken, month_taken, day_taken),
				)
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

func move_videos(watch_dir string, destination_dir string) error {
	files, err := ioutil.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, ext := range config.VideoExtensions {
			if strings.ToLower(filepath.Ext(file.Name())) == ext {
				info, err := os.Stat(filepath.Join(watch_dir, file.Name()))
				if err != nil {
					log.Fatal(err)
				}
				date_taken := info.ModTime()
				year_taken, month_taken, day_taken := date_taken.Date()
				month_name := month_taken.String()
				full_destination_dir := filepath.Join(
					destination_dir,
					fmt.Sprintf("%d", year_taken),
					fmt.Sprintf("%d - %s", month_taken, month_name),
					fmt.Sprintf("%d-%d-%d", year_taken, month_taken, day_taken),
				)
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
