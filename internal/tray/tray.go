//go:build darwin || windows

package tray

import (
	"errors"
	"fmt"
	"log"

	"github.com/getlantern/systray"

	"objection-keys/internal/app"
)

type Runner struct {
	app *app.App
}

func Run(soundsDir string) {
	r := &Runner{app: app.New(soundsDir)}
	systray.Run(r.onReady, r.onExit)
}

func (r *Runner) onReady() {
	systray.SetTemplateIcon(templateIconBytes(), regularIconBytes())
	systray.SetTitle("")
	systray.SetTooltip("Objection Keys")

	statusItem := systray.AddMenuItem("Starting...", "Current status")
	statusItem.Disable()

	toggleItem := systray.AddMenuItem("Pause Sounds", "Pause or resume keyboard sounds")
	volumeMenu := systray.AddMenuItem("Volume", "Set sound effect volume")
	volumeItems := addVolumeItems(volumeMenu)
	permissionItem := systray.AddMenuItem("Check Accessibility Permission", "Request macOS Accessibility permission")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Quit Objection Keys")
	setVolumeChecks(volumeItems, r.app.Volume())

	if err := r.start(); err != nil {
		setStatus(statusItem, toggleItem, err)
	} else {
		setRunningStatus(statusItem, toggleItem, r.app.Enabled())
	}

	go func() {
		for {
			select {
			case <-toggleItem.ClickedCh:
				if !r.app.Running() {
					if err := r.start(); err != nil {
						setStatus(statusItem, toggleItem, err)
						continue
					}
				} else {
					r.app.SetEnabled(!r.app.Enabled())
				}
				setRunningStatus(statusItem, toggleItem, r.app.Enabled())

			case <-volumeItems[0].item.ClickedCh:
				r.setVolume(volumeItems, volumeItems[0].volume)
			case <-volumeItems[1].item.ClickedCh:
				r.setVolume(volumeItems, volumeItems[1].volume)
			case <-volumeItems[2].item.ClickedCh:
				r.setVolume(volumeItems, volumeItems[2].volume)
			case <-volumeItems[3].item.ClickedCh:
				r.setVolume(volumeItems, volumeItems[3].volume)

			case <-permissionItem.ClickedCh:
				statusItem.SetTitle("Waiting for permission...")
				if app.PromptAccessibility() {
					if err := r.start(); err != nil {
						setStatus(statusItem, toggleItem, err)
						continue
					}
					setRunningStatus(statusItem, toggleItem, r.app.Enabled())
				} else {
					setStatus(statusItem, toggleItem, app.ErrPermissionRequired)
				}

			case <-quitItem.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func (r *Runner) onExit() {
	r.app.Stop()
}

func (r *Runner) setVolume(items []volumeItem, volume float64) {
	r.app.SetVolume(volume)
	setVolumeChecks(items, volume)
}

func (r *Runner) start() error {
	if err := r.app.Start(); err != nil {
		log.Printf("failed to start Objection Keys: %v", err)
		return err
	}
	return nil
}

type volumeItem struct {
	item   *systray.MenuItem
	volume float64
}

func addVolumeItems(parent *systray.MenuItem) []volumeItem {
	return []volumeItem{
		{item: parent.AddSubMenuItemCheckbox("25%", "Set volume to 25%", false), volume: 0.25},
		{item: parent.AddSubMenuItemCheckbox("50%", "Set volume to 50%", false), volume: 0.50},
		{item: parent.AddSubMenuItemCheckbox("75%", "Set volume to 75%", false), volume: 0.75},
		{item: parent.AddSubMenuItemCheckbox("100%", "Set volume to 100%", false), volume: 1.00},
	}
}

func setVolumeChecks(items []volumeItem, volume float64) {
	for _, item := range items {
		if item.volume == volume {
			item.item.Check()
		} else {
			item.item.Uncheck()
		}
	}
}

func setRunningStatus(statusItem, toggleItem *systray.MenuItem, enabled bool) {
	if enabled {
		statusItem.SetTitle("Listening")
		toggleItem.SetTitle("Pause Sounds")
		toggleItem.Uncheck()
		return
	}

	statusItem.SetTitle("Paused")
	toggleItem.SetTitle("Resume Sounds")
	toggleItem.Check()
}

func setStatus(statusItem, toggleItem *systray.MenuItem, err error) {
	if errors.Is(err, app.ErrPermissionRequired) {
		statusItem.SetTitle("Accessibility permission required")
		toggleItem.SetTitle("Start Listening")
		return
	}

	statusItem.SetTitle(fmt.Sprintf("Error: %v", err))
	toggleItem.SetTitle("Retry")
}
