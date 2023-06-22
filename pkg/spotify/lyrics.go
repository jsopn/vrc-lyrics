package spotify

import (
	"fmt"
	"io"
	"strconv"

	"github.com/valyala/fastjson"
)

type LyricsLines struct {
	StartTime int
	Words     string
}

func (s *SpotifyClient) GetLyrics(trackID string) (lyrics []LyricsLines, err error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	req, err := s.newRequest("GET", fmt.Sprintf("https://spclient.wg.spotify.com/color-lyrics/v2/track/%s?format=json&market=from_token", trackID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", "Bearer "+token.AccessToken)
	req.Header.Set("App-platform", "WebPlayer")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO: dirty, i'll rewrite this later
	body, err := fastjson.Parse(string(b))
	if err != nil {
		return nil, err
	}

	if string(body.Get("lyrics", "syncType").GetStringBytes()) != "LINE_SYNCED" {
		return nil, nil
	}

	for _, v := range body.GetArray("lyrics", "lines") {
		startTime, err := strconv.Atoi(string(v.GetStringBytes("startTimeMs")))
		if err != nil {
			return nil, err
		}

		lyrics = append(lyrics, LyricsLines{
			StartTime: startTime,
			Words:     string(v.GetStringBytes("words")),
		})
	}

	return lyrics, nil
}

func GetCurrentWords(lyrics []LyricsLines, currentMs int) string {
	for i := len(lyrics) - 1; i >= 0; i-- {
		line := lyrics[i]
		if currentMs < line.StartTime {
			continue
		}

		return line.Words
	}

	return ""
}
