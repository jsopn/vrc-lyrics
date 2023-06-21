package app

import (
	"log"
	"time"

	"github.com/jsopn/vrc-lyrics/internal/config"
	"github.com/jsopn/vrc-lyrics/pkg/osc"
	"github.com/jsopn/vrc-lyrics/pkg/spotify"
)

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func Run(cfg *config.Config) error {
	oscClient := osc.New(cfg.VRChat.OSCHost, cfg.VRChat.OSCPort)
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

	ticker := time.NewTicker(500 * time.Millisecond)
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

				if len(syncedLyrics) > 0 {
					log.Println("Lyrics updated.")
				} else {
					log.Println("No lyrics found for this track.")
				}

				reverse(syncedLyrics)
			}

		case <-ticker.C:
			if playbackState == nil || trackMetadata == nil || !playbackState.IsPlaying {
				continue
			}

			data := map[string]interface{}{
				"trackID": playbackState.TrackID,
				"artist":  trackMetadata.Artists,
				"album":   trackMetadata.Album,
				"name":    trackMetadata.Name,
			}

			if playbackState.IsPaused && cfg.VRChat.PausedFormat != "" {
				ticker.Reset(5 * time.Second)

				oscClient.Send(cfg.VRChat.PausedFormat, data)
				continue
			}

			ticker.Reset(500 * time.Millisecond)

			delta := time.Since(playbackState.UpdatedAt).Milliseconds()
			currentMS := playbackState.CurrentMS.Milliseconds() + delta
			line := spotify.GetCurrentWords(syncedLyrics, int(currentMS))

			if len(syncedLyrics) == 0 || line == "" {
				oscClient.Send(cfg.VRChat.NoLyricsFormat, data)
				continue
			}

			if lastWords == line {
				continue
			}

			lastWords = line
			data["line"] = line

			oscClient.Send(cfg.VRChat.LyricsFormat, data)
		}
	}
}
