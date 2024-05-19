package utility

import (
	"errors"
	"io"
	"regexp"
)

var ErrNoScriptPath = errors.New("no script path found")

func Xor(buf []byte, key []byte) []byte {
	out := make([]byte, len(buf))
	m := len(key)
	var i int
	for j := 0; j < len(buf); j++ {
		out[j] = buf[j] ^ key[i]
		i++
		if i == m {
			i = 0
		}
	}
	return out
}

// ScrapeScriptPath tries to find Akamai script path in html document. You need to join site domain and script path yourself, eg. "https://www.example.com" + path
//
// Parameters:
//   - body: html document, eg. resp.Body
//
// Returns:
//   - string: Akamai script path
//   - error: nil if error didn't occur
func ScrapeScriptPath(body io.Reader) (string, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`<script type=\"text/javascript\"\s+(?:nonce=\".*\")?\s+src=\"([A-Za-z\d/\-_]+)\"></script>`).FindAllSubmatch(b, -1)
	if len(re) < 1 || len(re[0]) < 2 {
		return "", ErrNoScriptPath
	}

	return string(re[0][1]), nil
}
