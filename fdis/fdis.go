package fdis

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	api "github.com/xMajkel/akamai-api-kit"
)

const apiGenerateWebSensorEP = "https://akamai.fdisservices.co/v2/web"

var ErrFdis = errors.New("fdis")

// Static check to make sure type implements all functions of the interface correctly
var _ api.Provider = (*FdisApi)(nil)

type FdisApi struct {
	*api.Config
	scriptBodyHash string
	httpClient     *http.Client
}

type Option func(*FdisApi)

// WithHttpClient sets a custom http.Client that will be used for API requests
func WithHttpClient(client *http.Client) Option {
	return func(ap *FdisApi) {
		ap.httpClient = client
	}
}

func NewApi(conf *api.Config, options ...Option) *FdisApi {
	ap := &FdisApi{Config: conf}

	ap.httpClient = http.DefaultClient

	for _, option := range options {
		option(ap)
	}

	return ap
}

// SetScriptUrl sets the URL of the script to be used, has to be with domain, eg. "https://www.example.com/akamai_script_path"
func (ap *FdisApi) SetScriptUrl(url string) {
	ap.ScriptUrl = url
}

// SetScriptBody sets the body of the script to be used and computes its hash if dynamic is set in config. It doesn't return any errors
func (ap *FdisApi) SetScriptBody(body []byte) error {
	ap.ScriptBody = body
	if ap.Dynamic {
		hash := md5.Sum(body)
		ap.scriptBodyHash = hex.EncodeToString(hash[:])
	}
	return nil
}

type generateWebSensorJson struct {
	Url         string `json:"url"`
	Abck        string `json:"abck"`
	Bmsz        string `json:"bm_sz,omitempty"`
	ScriptUrl   string `json:"scriptUrl"`
	Type        int    `json:"type"`
	UserAgent   string `json:"userAgent"`
	Keyboard    bool   `json:"keyboard"`
	Dynamic     bool   `json:"dynamic"`
	DynamicHash string `json:"dynamicHash,omitempty"`
}

type apiResponseJson struct {
	Error  string `json:"error"`
	Sensor string `json:"sensor"`
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
func (ap *FdisApi) GenerateWebSensor(iteration int, abck string, bmsz string) (string, error) {
	payload := generateWebSensorJson{
		UserAgent: ap.UserAgent,
		Url:       ap.Site,
		Type:      1,
		Abck:      abck,
		ScriptUrl: ap.ScriptUrl,
		Keyboard:  false,
	}

	if bmsz != "" {
		payload.Bmsz = bmsz
		payload.Type = 2
	}

	if ap.Dynamic {
		payload.Dynamic = true
		payload.DynamicHash = ap.scriptBodyHash
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Join(ErrFdis, err)
	}

	req, err := http.NewRequest(http.MethodPost, apiGenerateWebSensorEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", errors.Join(ErrFdis, err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", ap.ApiKey)

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", errors.Join(ErrFdis, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.Join(ErrFdis, errors.New(resp.Status))
	}

	var respJson apiResponseJson
	err = json.NewDecoder(resp.Body).Decode(&respJson)
	if err != nil {
		return "", errors.Join(ErrFdis, err)
	}

	if respJson.Error != "" {
		return "", errors.Join(ErrFdis, errors.New(respJson.Error))
	}

	return respJson.Sensor, nil
}
