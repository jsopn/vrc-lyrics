package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jsopn/vrc-lyrics/pkg/spotify"
)

type GeneralConfig struct {
	// Lyrics update rate
	UpdateRate int
}

type SpotifyConfig struct {
	// SP_DC cookie from the open.spotify.com, which will be used to get the token
	SPDCCookie string

	// A token that will be used to access the spotify API. This is automatically set using the SP_DC cookie. **Do not change**.
	Token *spotify.Token
}

type VRChatConfig struct {
	// Connection data for VRChat's OSC
	OSCHost string
	OSCPort int

	// OSC Ratelimit (milliseconds per message)
	Ratelimit int

	// The formatted string that will be displayed if there **are lyrics** in the current track
	// Available fields: {{.artist}}, {{.name}}, {{.line}}, {{.trackID}}
	LyricsFormat string

	// Formatted string that will be displayed if there are **no lyrics** in the track
	// Available fields: {{.artist}}, {{.name}}, {{.trackID}}
	NoLyricsFormat string

	// Formatted string that will be displayed if track is paused
	// Available fields: {{.artist}}, {{.name}}, {{.trackID}}
	PausedFormat string
}

type Config struct {
	General GeneralConfig
	Spotify SpotifyConfig
	VRChat  VRChatConfig
}

func ParseConfig(path string) (config *Config, err error) {
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func WriteConfig(path string, config *Config) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	return toml.NewEncoder(f).Encode(config)
}
