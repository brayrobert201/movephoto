package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"log"
	"image/jpeg"
)

var (
	defaultWatchDir        = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
	robertWatchDir         = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
	defaultDestinationDir  = "/mnt/c/Users/bob/OneDrive/Camera"
	imageExtensions        = []string{".jpg", ".jpeg"}
	videoExtensions        = []string{".mp4", ".mov", ".avi", ".flv", ".wmv", ".mkv"}
	bannedExtensions       = []string{".png"}
)

func main() {
	purge_unwanted(robertWatchDir, bannedExtensions)
	move_photos(robertWatchDir, defaultDestinationDir, imageExtensions)
	move_videos(robertWatchDir, defaultDestinationDir)
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
				img, err := jpeg.DecodeConfig(imgFile)
				if err != nil {
					log.Fatal(err)
				}
				date_taken := img.ModTime()
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
}
func move_videos(watch_dir string, destination_dir string) error {
	files, err := ioutil.ReadDir(watch_dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		for _, ext := range video_extensions {
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
}
