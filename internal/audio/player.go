// Package audio provides WAV sound playback using ebitengine/oto.
package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ebitengine/oto/v3"
)

// Player manages audio playback for sound effects.
type Player struct {
	ctx     *oto.Context
	mu      sync.Mutex
	sounds  map[string][]byte
	players []*oto.Player
	volume  float64
}

// NewPlayer creates a new audio player.
func NewPlayer() (*Player, error) {
	ctx, readyCh, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   22050,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, err
	}
	<-readyCh

	return &Player{
		ctx:    ctx,
		sounds: make(map[string][]byte),
		volume: 1,
	}, nil
}

// Register loads a WAV file and registers it by name.
func (p *Player) Register(name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// Parse WAV header to find the "data" chunk offset.
	// WAV files may contain extra chunks (e.g., LIST metadata) before the data chunk,
	// so we cannot assume a fixed 44-byte header.
	dataBuf := make([]byte, 12)
	if _, err := io.ReadFull(f, dataBuf); err != nil {
		f.Close()
		return fmt.Errorf("read wav header: %w", err)
	}

	// Verify RIFF header
	if string(dataBuf[:4]) != "RIFF" || string(dataBuf[8:12]) != "WAVE" {
		f.Close()
		return fmt.Errorf("not a valid WAV file")
	}

	var sawFormat bool

	// Seek to the first chunk (offset 12) and scan for the format and data chunks.
	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(f, chunkHeader); err != nil {
			f.Close()
			return fmt.Errorf("read chunk header: %w", err)
		}

		chunkID := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		if chunkID == "fmt " {
			if chunkSize < 16 {
				f.Close()
				return fmt.Errorf("invalid WAV fmt chunk")
			}
			format := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, format); err != nil {
				f.Close()
				return fmt.Errorf("read fmt chunk: %w", err)
			}

			audioFormat := binary.LittleEndian.Uint16(format[0:2])
			channels := binary.LittleEndian.Uint16(format[2:4])
			sampleRate := binary.LittleEndian.Uint32(format[4:8])
			bitsPerSample := binary.LittleEndian.Uint16(format[14:16])
			if audioFormat != 1 || channels != 2 || sampleRate != 22050 || bitsPerSample != 8 {
				f.Close()
				return fmt.Errorf("unsupported WAV format: want 8-bit unsigned PCM, 22050 Hz, stereo")
			}

			sawFormat = true
			if chunkSize%2 != 0 {
				if _, err := f.Seek(1, io.SeekCurrent); err != nil {
					f.Close()
					return err
				}
			}
			continue
		}

		if chunkID == "data" {
			if !sawFormat {
				f.Close()
				return fmt.Errorf("missing WAV fmt chunk before data")
			}

			data := make([]byte, chunkSize)
			if _, err := io.ReadFull(f, data); err != nil {
				f.Close()
				return fmt.Errorf("read data chunk: %w", err)
			}

			p.sounds[name] = unsigned8ToSigned16LE(data)
			return nil
		}

		// Skip this chunk's content (WAV chunks can be odd-sized, padded with a byte)
		pad := int64(chunkSize) % 2
		if _, err := f.Seek(int64(chunkSize)+pad, 1); err != nil {
			f.Close()
			return err
		}
	}
}

func unsigned8ToSigned16LE(src []byte) []byte {
	dst := make([]byte, len(src)*2)
	for i, sample := range src {
		v := int16((int(sample) - 128) << 8)
		binary.LittleEndian.PutUint16(dst[i*2:i*2+2], uint16(v))
	}
	return dst
}

// Play starts playing a registered sound effect.
// If the sound is already playing, it resets and replays from the beginning.
func (p *Player) Play(name string) {
	data, ok := p.sounds[name]
	if !ok {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	active := p.players[:0]
	for _, player := range p.players {
		if player.IsPlaying() {
			active = append(active, player)
		}
	}
	p.players = active

	player := p.ctx.NewPlayer(bytes.NewReader(data))
	player.SetVolume(p.volume)
	player.Play()
	p.players = append(p.players, player)
}

func (p *Player) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.volume = volume
	for _, player := range p.players {
		player.SetVolume(volume)
	}
}

func (p *Player) Volume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

// Close stops playback and releases all resources.
func (p *Player) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, player := range p.players {
		player.Pause()
		player.Close()
	}
	p.players = nil
}
