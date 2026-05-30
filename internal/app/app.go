//go:build darwin

package app

import (
	"errors"
	"sync"

	"objection-keys/internal/keyboard"
	"objection-keys/internal/sound"
)

var ErrPermissionRequired = errors.New("accessibility permission required")

type App struct {
	soundsDir string

	mu      sync.RWMutex
	sounds  *sound.SoundMap
	kb      *keyboard.Listener
	done    chan struct{}
	wg      sync.WaitGroup
	running bool
	enabled bool
	volume  float64
}

func New(soundsDir string) *App {
	return &App{
		soundsDir: soundsDir,
		volume:    1,
	}
}

func (a *App) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	if !keyboard.HasAccessibilityPermission() {
		return ErrPermissionRequired
	}

	sounds, err := sound.LoadAll(a.soundsDir)
	if err != nil {
		return err
	}
	sounds.SetVolume(a.Volume())

	kb, err := keyboard.New()
	if err != nil {
		sounds.Close()
		return err
	}

	done := make(chan struct{})

	a.mu.Lock()
	a.sounds = sounds
	a.kb = kb
	a.done = done
	a.running = true
	a.enabled = true
	a.wg.Add(1)
	a.mu.Unlock()

	go a.run(kb, sounds, done)

	return nil
}

func (a *App) run(kb *keyboard.Listener, sounds *sound.SoundMap, done <-chan struct{}) {
	defer a.wg.Done()

	for {
		select {
		case keyName, ok := <-kb.Keys():
			if !ok {
				return
			}
			if a.Enabled() {
				sounds.Play(keyName)
			}
		case <-done:
			return
		}
	}
}

func (a *App) Stop() {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return
	}

	done := a.done
	kb := a.kb
	sounds := a.sounds

	a.done = nil
	a.kb = nil
	a.sounds = nil
	a.running = false
	a.enabled = false
	a.mu.Unlock()

	close(done)
	kb.Close()
	a.wg.Wait()
	sounds.Close()
}

func (a *App) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		a.enabled = enabled
	}
}

func (a *App) Enabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running && a.enabled
}

func (a *App) Running() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

func (a *App) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	a.mu.Lock()
	a.volume = volume
	sounds := a.sounds
	a.mu.Unlock()

	if sounds != nil {
		sounds.SetVolume(volume)
	}
}

func (a *App) Volume() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.volume
}

func PromptAccessibility() bool {
	return keyboard.PromptAccessibility()
}
