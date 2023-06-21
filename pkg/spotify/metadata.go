package spotify

import (
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/valyala/fastjson"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type TrackMetadata struct {
	Artists string
	Album   string
	Name    string
}

func ConvertB62(id string) []byte {
	base := big.NewInt(62)

	n := &big.Int{}
	for _, c := range []byte(id) {
		d := big.NewInt(int64(strings.IndexByte(alphabet, c)))
		n = n.Mul(n, base)
		n = n.Add(n, d)
	}

	nBytes := n.Bytes()
	if len(nBytes) < 16 {
		paddingBytes := make([]byte, 16-len(nBytes))
		nBytes = append(paddingBytes, nBytes...)
	}

	return nBytes
}

func (s *SpotifyClient) GetMetadata(trackID string) (metadata *TrackMetadata, err error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	req, err := s.newRequest("GET", fmt.Sprintf("https://spclient.wg.spotify.com/metadata/4/track/%s?market=from_token", fmt.Sprintf("%x", ConvertB62(trackID))), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", "Bearer "+token.AccessToken)
	req.Header.Set("App-platform", "WebPlayer")
	req.Header.Set("Accept", "application/json")

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

	artists := ""
	for _, v := range body.GetArray("artist") {
		artists += string(v.GetStringBytes("name")) + ", "
	}

	return &TrackMetadata{
		Name:    string(body.GetStringBytes("name")),
		Album:   string(body.GetStringBytes("album", "name")),
		Artists: strings.TrimSuffix(artists, ", "),
	}, nil
}
