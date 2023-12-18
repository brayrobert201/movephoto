package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config := loadConfig()
	if config.DefaultWatchDir == "" {
		t.Error("Failed to load config")
	}
}

func TestPurgeUnwanted(t *testing.T) {
	// TODO: Create a temporary directory with a file having a banned extension
	// TODO: Call purge_unwanted()
	// TODO: Check if the file was deleted
}

func TestMoveFiles(t *testing.T) {
	// TODO: Create a temporary directory with a file
	// TODO: Call move_files()
	// TODO: Check if the file was moved
}

func TestMovePhotos(t *testing.T) {
	// TODO: Create a temporary directory with a photo file
	// TODO: Call move_photos()
	// TODO: Check if the photo was moved
}

func TestMoveVideos(t *testing.T) {
	// TODO: Create a temporary directory with a video file
	// TODO: Call move_videos()
	// TODO: Check if the video was moved
}
