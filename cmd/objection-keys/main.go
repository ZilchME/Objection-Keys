//go:build darwin

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"objection-keys/internal/keyboard"
	"objection-keys/internal/sound"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	promptFlag := flag.Bool("prompt", false, "open system settings to grant accessibility permission")
	flag.Parse()

	// Check accessibility permission first
	if !keyboard.HasAccessibilityPermission() {
		if *promptFlag {
			log.Println("Opening system settings for Accessibility permission...")
			if keyboard.PromptAccessibility() {
				log.Println("Permission granted! Restart the app.")
			} else {
				log.Println("Permission was not granted. Please grant it manually:")
				log.Println("  System Settings → Privacy & Security → Accessibility")
				log.Println("  Then click '+' and add this application")
			}
			os.Exit(1)
		}

		log.Println("⚠️  Accessibility permission required for keyboard monitoring")
		log.Println("")
		log.Println("  Please grant permission manually:")
		log.Println("  System Settings → Privacy & Security → Accessibility")
		log.Println("  Then click '+' and add this application")
		log.Println("")
		log.Println("  Or run with --prompt to open system settings:")
		log.Println("    ./objection-keys --prompt")
		os.Exit(1)
	}

	// Register all sound effects (also initializes the audio player)
	sounds, err := sound.LoadAll("sounds")
	if err != nil {
		log.Fatalf("Failed to load sounds: %v", err)
	}
	defer sounds.Close()

	// Start keyboard listener
	kb, err := keyboard.New()
	if err != nil {
		log.Fatalf("Failed to start keyboard listener: %v", err)
	}
	defer kb.Close()

	// Signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("⚖️ Listening to keyboard... (press Ctrl+C to exit)")

	// Event loop
	for {
		select {
		case keyName := <-kb.Keys():
			sounds.Play(keyName)
		case <-sigCh:
			log.Println("\nShutting down...")
			return
		}
	}
}
