//go:build darwin || windows

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"objection-keys/internal/tray"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	soundsDir := flag.String("sounds", "", "path to sound effects directory")
	flag.Parse()

	tray.Run(resolveSoundsDir(*soundsDir))
}

func resolveSoundsDir(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	if _, err := os.Stat("sounds"); err == nil {
		return "sounds"
	}

	exe, err := os.Executable()
	if err != nil {
		return "sounds"
	}

	exeSounds := filepath.Join(filepath.Dir(exe), "sounds")
	if _, err := os.Stat(exeSounds); err == nil {
		return exeSounds
	}

	resourcesSounds := filepath.Join(filepath.Dir(exe), "..", "Resources", "sounds")
	if _, err := os.Stat(resourcesSounds); err == nil {
		return resourcesSounds
	}

	return "sounds"
}
