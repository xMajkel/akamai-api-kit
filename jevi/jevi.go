package jevi

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	api "github.com/xMajkel/akamai-api-kit"
	"github.com/xMajkel/akamai-api-kit/utility"
)

const apiGenerateSensorEP = "https://www.jevi.dev/Akamai"

var ErrJevi = errors.New("jevi")

// Static check to make sure type implements all functions of the interface correctly
var _ api.Provider = (*JeviApi)(nil)

type JeviApi struct {
	*api.Config
	decryptionKey  string
	scriptBodyHash string
	httpClient     *http.Client
}

type Option func(*JeviApi)

// WithDecryptor sets a decryption key for decrypying responses the API
func WithDecryptor(decryptionKey string) Option {
	return func(ap *JeviApi) {
		ap.decryptionKey = decryptionKey
	}
}

// WithHttpClient sets a custom http.Client that will be used for API requests
func WithHttpClient(client *http.Client) Option {
	return func(ap *JeviApi) {
		ap.httpClient = client
	}
}

func NewApi(conf *api.Config, options ...Option) *JeviApi {
	ap := &JeviApi{Config: conf}

	ap.httpClient = http.DefaultClient

	for _, option := range options {
		option(ap)
	}

	return ap
}

// SetScriptUrl sets the URL of the script to be used, has to be with domain, eg. "https://www.example.com/akamai_script_path"
func (ap *JeviApi) SetScriptUrl(url string) {
	ap.ScriptUrl = url
}

// SetScriptBody sets the body of the script to be used and computes its hash if dynamic is set in config. It doesn't return any errors
func (ap *JeviApi) SetScriptBody(body []byte) error {
	ap.ScriptBody = body
	if ap.Dynamic {
		checksumBytes := md5.Sum(body)
		ap.scriptBodyHash = hex.EncodeToString(checksumBytes[:])
	}
	return nil
}

// NewApiConfig returns new config for the API providers, it is required to create api provider implementing the Provider interface
//
// Parameters:
//   - iteration: which time sensor is generated
//   - abck: _abck cookie value
//   - bmsz: bm_sz cookie value, optional for Akamai before verison 2
//
// Returns:
//   - string: sensor string that has to be posted to the script URL
//   - error: nil if error didn't occur
func (ap *JeviApi) GenerateWebSensor(iteration int, abck string, bmsz string) (string, error) {

	form := url.Values{
		"key":  {ap.ApiKey},
		"site": {ap.Site},
		"mode": {"API"},
		"ua":   {ap.UserAgent},
		"abck": {abck},
	}

	if bmsz != "" {
		form.Add("bmsz", bmsz)
	}

	if ap.Dynamic {
		form.Add("etag", ap.scriptBodyHash)
	}

	req, err := http.NewRequest(http.MethodPost, apiGenerateSensorEP, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Join(ErrJevi, err)
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", errors.Join(ErrJevi, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.Join(ErrJevi, errors.New(resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Join(ErrJevi, err)
	}

	if ap.decryptionKey != "" {
		decryptedBody := utility.Xor(body, []byte(ap.decryptionKey))
		return string(decryptedBody), nil
	}

	return string(body), nil
}
