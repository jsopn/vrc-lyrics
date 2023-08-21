package app

import (
	"fmt"
	"log"
	"time"

	"github.com/jsopn/vrc-lyrics/internal/config"
	"github.com/jsopn/vrc-lyrics/pkg/osc"
	"github.com/jsopn/vrc-lyrics/pkg/spotify"
)

func Run(cfg *config.Config) error {
	oscClient := osc.New(cfg.VRChat.OSCHost, cfg.VRChat.OSCPort, cfg.VRChat.Ratelimit)
	spt := spotify.New(cfg.Spotify.Token, cfg.Spotify.SPDCCookie)

	log.Println("Connecting to Spotify's WebSocket")
	playbackChan, err := spt.ConnectWebsocket()
	if err != nil {
		return err
	}

	defer spt.CloseWebsocket()

	if err := spt.RegisterDevice(); err != nil {
		return err
	}

	if err := spt.UpdateState(); err != nil {
		return err
	}

	log.Println("Connected.")

	ticker := time.NewTicker(time.Duration(cfg.General.UpdateRate) * time.Millisecond)
	defer ticker.Stop()

	var playbackState *spotify.PlaybackState
	var syncedLyrics []spotify.LyricsLines
	var trackMetadata *spotify.TrackMetadata
	var syncedLyricsTrackID string
	var lastWords string

	for {
		select {
		case ps := <-playbackChan:
			playbackState = ps
			log.Printf("[ Updated playback state. | TrackID: %s ]", playbackState.TrackID)

			if syncedLyricsTrackID != playbackState.TrackID {
				syncedLyrics, _ = spt.GetLyrics(playbackState.TrackID)
				trackMetadata, _ = spt.GetMetadata(playbackState.TrackID)
				syncedLyricsTrackID = playbackState.TrackID
				lastWords = ""

				if len(syncedLyrics) > 0 {
					log.Println("Lyrics updated.")
				} else {
					log.Println("No lyrics found for this track.")
				}
			}

		case <-ticker.C:
			if playbackState == nil || trackMetadata == nil || !playbackState.IsPlaying {
				continue
			}

			delta := time.Since(playbackState.UpdatedAt)
			currentPosition := playbackState.Current + delta

			if playbackState.IsPaused {
				currentPosition = playbackState.Current
			}

			data := map[string]interface{}{
				"trackID":    playbackState.TrackID,
				"artist":     trackMetadata.Artists,
				"album":      trackMetadata.Album,
				"name":       trackMetadata.Name,
				"currentPos": fmt.Sprintf("%d:%02d", int(currentPosition.Seconds())/60, int(currentPosition.Seconds())%60),
				"duration":   fmt.Sprintf("%d:%02d", int(playbackState.Duration.Seconds())/60, int(playbackState.Duration.Seconds())%60),
			}

			if playbackState.IsPaused && cfg.VRChat.PausedFormat != "" {
				ticker.Reset(5 * time.Second)

				oscClient.Send(cfg.VRChat.PausedFormat, data, false)
				continue
			}

			ticker.Reset(500 * time.Millisecond)
			line := spotify.GetCurrentWords(syncedLyrics, int(currentPosition.Milliseconds()))

			if len(syncedLyrics) == 0 || line == "" {
				oscClient.Send(cfg.VRChat.NoLyricsFormat, data, false)
				continue
			}

			lastWords = line
			data["line"] = line

			skipRate := lastWords == ""
			oscClient.Send(cfg.VRChat.LyricsFormat, data, skipRate)
		}
	}
}
