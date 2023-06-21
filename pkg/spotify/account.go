package spotify

import (
	"fmt"
	"io"
	"time"

	"github.com/valyala/fastjson"
)

type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (s *SpotifyClient) RefreshToken() (*Token, error) {
	req, err := s.newRequest("GET", "https://open.spotify.com/get_access_token?reason=transport&productType=web_player", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	req.Header.Set("Cookie", "sp_dc="+s.cookie)
	req.Header.Set("App-platform", "WebPlayer0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body, err := fastjson.Parse(string(b))
	if err != nil {
		return nil, err
	}

	if body.GetBool("isAnonymous") {
		return nil, fmt.Errorf("provided cookie is not valid")
	}

	accessToken := string(body.GetStringBytes("accessToken"))
	expirationTimestamp := body.GetInt64("accessTokenExpirationTimestampMs")

	return &Token{
		AccessToken: accessToken,
		ExpiresAt:   time.UnixMilli(expirationTimestamp),
	}, nil
}

func (s *SpotifyClient) getToken() (*Token, error) {
	if s.token != nil && !s.token.IsExpired() {
		return s.token, nil
	}

	token, err := s.RefreshToken()
	if err != nil {
		return nil, err
	}

	s.token = token
	return token, nil
}
