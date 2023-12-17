package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	defaultWatchDir        = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
	robertWatchDir         = "/mnt/c/Users/bob/OneDrive/Pictures/Camera Roll"
	defaultDestinationDir  = "/mnt/c/Users/bob/OneDrive/Camera"
	imageExtensions        = []string{".jpg", ".jpeg"}
	videoExtensions        = []string{".mp4", ".mov"}
	bannedExtensions       = []string{".png"}
)

func main() {
	fmt.Println("This is a placeholder for the main function.")
}
