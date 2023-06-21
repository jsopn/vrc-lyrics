package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type SpotifyClient struct {
	cookie       string
	deviceID     string
	connectionID string

	token      *Token
	httpClient *http.Client
	wsConn     *websocket.Conn
}

func New(token *Token, cookie string) *SpotifyClient {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 22)
	r.Read(b)

	return &SpotifyClient{
		token:    token,
		cookie:   cookie,
		deviceID: fmt.Sprintf("%x", b)[2:24],

		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *SpotifyClient) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0")

	return req, nil
}

func (s *SpotifyClient) simpleBodyRequest(method string, url string, body any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", "Bearer "+token.AccessToken)

	if s.connectionID != "" {
		req.Header.Set("x-spotify-connection-id", s.connectionID)
	}

	return s.httpClient.Do(req)
}
