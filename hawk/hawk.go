package hawk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	api "github.com/xMajkel/akamai-api-kit"
)

const apiGenerateWebSensorEP = "https://ak01-eu.hwkapi.com/akamai/generate"
const apiScriptConfigEP = "https://ak-ppsaua.hwkapi.com/006180d12cf7"
const apiScriptConfigCacheEP = "https://ak-ppsaua.hwkapi.com/006180d12cf7/c"

var ErrHawk = errors.New("hawk")
var separatorBmsz = []byte("****")

// Static check to make sure struct implements all functions of the interface correctly
var _ api.Provider = (*HawkApi)(nil)

type HawkApi struct {
	*api.Config
	scriptUrl           string
	scriptBody          []byte
	dynamicScriptConfig string
	httpClient          *http.Client
}

type Option func(*HawkApi)

// WithHttpClient sets a custom http.Client that will be used for API requests
func WithHttpClient(client *http.Client) Option {
	return func(ap *HawkApi) {
		ap.httpClient = client
	}
}

func NewApi(conf *api.Config, options ...Option) *HawkApi {
	ap := &HawkApi{Config: conf}

	ap.httpClient = http.DefaultClient

	for _, option := range options {
		option(ap)
	}

	return ap
}

// SetScriptUrl sets the URL of the script to be used, has to be with domain, eg. "https://www.example.com/akamai_script_path"
func (ap *HawkApi) SetScriptUrl(url string) {
	ap.scriptUrl = url
}

// SetScriptBody sets the body of the script to be used and if dynamic is set in config it fetches script config from API provider
// Return error if it couldn't fetch dynamic script config
func (ap *HawkApi) SetScriptBody(body []byte) error {
	ap.scriptBody = body

	if ap.Dynamic {
		err := ap.getDynamicScriptConfig(body)
		if err != nil {
			return err
		}
	}
	return nil
}

type generateWebSensorJson struct {
	Site      string `json:"site"`
	Abck      string `json:"abck"`
	Events    string `json:"events"`
	UserAgent string `json:"user_agent"`
	Bmsz      string `json:"bm_sz,omitempty"`
	Config    string `json:"config,omitempty"`
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
func (ap *HawkApi) GenerateWebSensor(iteration int, abck string, bmsz string) (string, error) {

	payload := generateWebSensorJson{
		UserAgent: ap.UserAgent,
		Site:      ap.Site,
		Abck:      abck,
		Events:    "0,0",
	}

	if iteration > 0 {
		payload.Events = "1,0"
	}

	if bmsz != "" {
		payload.Bmsz = bmsz
	}

	if ap.Dynamic {
		payload.Config = ap.dynamicScriptConfig
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiGenerateWebSensorEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Api-Key", ap.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sec", "new")

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", api.ErrConnection
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if bmsz != "" {
		return string(bytes.Split(b, separatorBmsz)[0]), nil
	}
	return string(b), nil
}

func (ap *HawkApi) getDynamicScriptConfig(scriptBody []byte) error {
	payload := map[string]string{
		"hash": string(scriptBody),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiScriptConfigCacheEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", ap.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sec", "new")

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return api.ErrConnection
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	config := string(b)

	if config != "false" {
		ap.dynamicScriptConfig = config
		return nil
	}

	// If no script config in cache:

	payload = map[string]string{
		"body": base64.StdEncoding.EncodeToString(scriptBody),
	}
	jsonPayload, err = json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", apiScriptConfigEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", ap.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sec", "new")

	resp, err = ap.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ap.dynamicScriptConfig = string(b)

	return nil
}
