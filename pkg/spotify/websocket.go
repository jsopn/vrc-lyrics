package spotify

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/valyala/fastjson"
)

// TODO: All of this is very dirty, needs a refactoring, i'll do it later.

type PlaybackState struct {
	TrackID string

	IsPlaying bool
	IsPaused  bool

	CurrentMS time.Duration
	Duration  time.Duration

	UpdatedAt time.Time
}

func (s *SpotifyClient) RegisterDevice() (err error) {
	resp, err := s.simpleBodyRequest("POST", "https://gew4-spclient.spotify.com/track-playback/v1/devices", map[string]interface{}{
		"device": map[string]interface{}{
			"brand": "spotify",
			"capabilities": map[string]interface{}{
				"change_volume":            false,
				"enable_play_token":        false,
				"supports_file_media_type": false,
				"disable_connect":          true,
				"audio_podcasts":           false,
				"video_playback":           false,
				"manifest_formats":         []map[string]interface{}{},
				"play_token_lost_behavior": "pause",
			},
			"device_id":           s.deviceID,
			"device_type":         "computer",
			"metadata":            map[string]interface{}{},
			"model":               "web_player",
			"name":                "Web Player (Firefox)",
			"platform_identifier": "web_player windows 10;firefox 82.0;desktop",
		},
		"connection_id":  s.connectionID,
		"client_version": "harmony:4.27.1-af7f4f3",
		"volume":         65535,
	})

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to register a device")
	}

	return nil
}

func (s *SpotifyClient) UpdateState() (err error) {
	resp, err := s.simpleBodyRequest("PUT", fmt.Sprintf("https://gew4-spclient.spotify.com/connect-state/v1/devices/hobs_%s", s.deviceID), map[string]interface{}{
		"member_type": "CONNECT_STATE",
		"device": map[string]interface{}{
			"device_info": map[string]interface{}{
				"capabilities": map[string]interface{}{
					"can_be_player":           false,
					"hidden":                  true,
					"needs_full_player_state": true,
				},
			},
		},
	})

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to update an state")
	}

	return nil
}

func (s *SpotifyClient) PingHandler(conn *websocket.Conn) {
	for {
		if err := conn.WriteJSON(map[string]interface{}{
			"type": "ping",
		}); err != nil {
			break
		}

		time.Sleep(10 * time.Second)
	}
}

func (s *SpotifyClient) WSHandler(conn *websocket.Conn, ch chan *PlaybackState) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("failed to read message:", err)
			return
		}

		msg, err := fastjson.Parse(string(message))
		if err != nil {
			log.Println("failed to parse message:", err)
			return
		}

		reason := string(msg.GetStringBytes("payloads", "0", "update_reason"))

		if string(msg.GetStringBytes("uri")) != "hm://connect-state/v1/cluster" || (reason != "DEVICE_STATE_CHANGED" && reason != "DEVICES_DISAPPEARED") {
			continue
		}

		playerState := msg.Get("payloads", "0", "cluster", "player_state")
		isPlaying := playerState.GetBool("is_playing")
		isPaused := playerState.GetBool("is_paused")

		if reason == "DEVICES_DISAPPEARED" {
			isPaused = true
		}

		contextURI := string(playerState.GetStringBytes("track", "uri"))

		position, err := strconv.Atoi(string(playerState.GetStringBytes("position_as_of_timestamp")))
		if err != nil || (position > 1500 && position < 3000) {
			continue
		}

		duration, err := strconv.Atoi(string(playerState.GetStringBytes("duration")))
		if err != nil {
			continue
		}

		timestamp, err := strconv.Atoi(string(msg.GetStringBytes("payloads", "0", "cluster", "timestamp")))
		if err != nil {
			continue
		}

		ch <- &PlaybackState{
			TrackID: strings.ReplaceAll(contextURI, "spotify:track:", ""),

			IsPlaying: isPlaying,
			IsPaused:  isPaused,
			CurrentMS: time.Duration(position) * time.Millisecond,
			Duration:  time.Duration(duration) * time.Millisecond,

			UpdatedAt: time.UnixMilli(int64(timestamp)),
		}
	}
}

func (s *SpotifyClient) ConnectWebsocket() (chan *PlaybackState, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	c, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://gew4-dealer.spotify.com/?access_token=%s", token.AccessToken), nil)
	if err != nil {
		return nil, err
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}

	playbackChan := make(chan *PlaybackState)
	go s.WSHandler(c, playbackChan)
	go s.PingHandler(c)

	s.connectionID = fastjson.GetString(message, "headers", "Spotify-Connection-Id")
	s.wsConn = c

	return playbackChan, nil
}

func (s *SpotifyClient) CloseWebsocket() error {
	return s.wsConn.Close()
}
