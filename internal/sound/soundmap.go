// Package sound provides sound effect loading and playback mapping.
package sound

import (
	"path/filepath"
	"strings"

	"objection-keys/internal/audio"
)

// KeyType defines the category of keyboard keys.
type KeyType int

const (
	KeyTypeAlpha KeyType = iota
	KeyTypeNumber
	KeyTypeSpace
	KeyTypeEnter
	KeyTypeBackspace
	KeyTypeEsc
)

// KeyName constants for the main event loop.
const (
	KeySpace     = "space"
	KeyEnter     = "enter"
	KeyBackspace = "backspace"
	KeyEsc       = "esc"
)

// SoundMap maps key types to their sound effect names.
type SoundMap struct {
	player   *audio.Player
	keyType  map[string]KeyType
	lastTime int64 // nanoseconds of last key press
}

var (
	alphabetKeys = []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l",
		"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}
	numberKeys = []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
		"-", "=", "+", "/", "*", "(", ")", "&", "^", "%", "$",
		"@", "!", "`", "~", "[", "]", "{", "}", ";", "'", "\\",
		"|", ",", "<", ">", "?",
	}
)

const fastAlphabetThresholdMs = 300

// NewSoundMap creates a new SoundMap with the given audio player.
func NewSoundMap(player *audio.Player) *SoundMap {
	sm := &SoundMap{
		player:   player,
		keyType:  make(map[string]KeyType),
		lastTime: 0,
	}

	for _, k := range alphabetKeys {
		sm.keyType[k] = KeyTypeAlpha
	}
	for _, k := range numberKeys {
		sm.keyType[k] = KeyTypeNumber
	}
	sm.keyType[KeySpace] = KeyTypeSpace
	sm.keyType[KeyEnter] = KeyTypeEnter
	sm.keyType[KeyBackspace] = KeyTypeBackspace
	sm.keyType[KeyEsc] = KeyTypeEsc

	return sm
}

// Play determines the appropriate sound effect and plays it.
func (sm *SoundMap) Play(keyName string) {
	kt, ok := sm.keyType[keyName]
	if !ok {
		return // Ignore unmapped keys
	}

	switch kt {
	case KeyTypeSpace:
		sm.player.Play("space")
	case KeyTypeEnter:
		sm.player.Play("enter")
	case KeyTypeBackspace:
		sm.player.Play("backspace")
	case KeyTypeEsc:
		sm.player.Play("esc")
	case KeyTypeNumber:
		sm.player.Play("number")
	case KeyTypeAlpha:
		sm.playAlphabet(keyName)
	}
}

func (sm *SoundMap) playAlphabet(keyName string) {
	now := nowNano()
	if sm.lastTime == 0 || (now-sm.lastTime) > (int64(fastAlphabetThresholdMs)*1_000_000) {
		sm.player.Play("alphabet_slow")
	} else {
		sm.player.Play("alphabet_fast")
	}
	sm.lastTime = now
}

func (sm *SoundMap) SetVolume(volume float64) {
	sm.player.SetVolume(volume)
}

func (sm *SoundMap) Volume() float64 {
	return sm.player.Volume()
}

// LoadAll loads all WAV files from the sounds directory and registers them.
func LoadAll(soundsDir string) (*SoundMap, error) {
	player, err := audio.NewPlayer()
	if err != nil {
		return nil, err
	}

	sounds := map[string]string{
		"alphabet_fast": filepath.Join(soundsDir, "alphabet_fast.wav"),
		"alphabet_slow": filepath.Join(soundsDir, "alphabet_slow.wav"),
		"number":        filepath.Join(soundsDir, "number.wav"),
		"space":         filepath.Join(soundsDir, "space.wav"),
		"enter":         filepath.Join(soundsDir, "enter.wav"),
		"backspace":     filepath.Join(soundsDir, "backspace.wav"),
		"esc":           filepath.Join(soundsDir, "esc.wav"),
	}

	for name, path := range sounds {
		if strings.HasSuffix(path, ".wav") {
			if err := player.Register(name, path); err != nil {
				player.Close()
				return nil, err
			}
		}
	}

	return NewSoundMap(player), nil
}

// Close releases all resources.
func (sm *SoundMap) Close() {
	sm.player.Close()
}
