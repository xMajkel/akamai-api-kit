package hypersolutions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	api "github.com/xMajkel/akamai-api-kit"
)

const apiGenerateWebSensorEP = "https://akm.justhyped.dev/sensor"
const apiScriptConfigEP = "https://akm.justhyped.dev/dynamic"

var ErrHyperSolutions = errors.New("hyper-solutions")

// Static check to make sure type implements all functions of the interface correctly
var _ api.Provider = (*HyperSolutionsApi)(nil)

type HyperSolutionsApi struct {
	*api.Config
	dynamicScriptConfig string
	scriptBodyHash      string
	httpClient          *http.Client
}

type Option func(*HyperSolutionsApi)

// WithHttpClient sets a custom http.Client that will be used for API requests
func WithHttpClient(client *http.Client) Option {
	return func(ap *HyperSolutionsApi) {
		ap.httpClient = client
	}
}

func NewApi(conf *api.Config, options ...Option) *HyperSolutionsApi {
	ap := &HyperSolutionsApi{Config: conf}

	ap.httpClient = http.DefaultClient

	for _, option := range options {
		option(ap)
	}

	return ap
}

// SetScriptUrl sets the URL of the script to be used, has to be with domain, eg. "https://www.example.com/akamai_script_path"
func (ap *HyperSolutionsApi) SetScriptUrl(url string) {
	ap.ScriptUrl = url
}

// SetScriptBody sets the body of the script to be used, computes its hash and if dynamic is set in config it fetches script config from API provider
// Return error if it couldn't fetch dynamic script config
func (ap *HyperSolutionsApi) SetScriptBody(body []byte) error {
	ap.ScriptBody = body

	hash := sha256.Sum256(body)
	ap.scriptBodyHash = hex.EncodeToString(hash[:])

	if ap.Dynamic {
		err := ap.getDynamicScriptConfig(body)
		if err != nil {
			return errors.Join(ErrHyperSolutions, err)
		}
	}
	return nil
}

type generateWebSensorJson struct {
	UserAgent  string `json:"userAgent"`
	PageUrl    string `json:"pageUrl"`
	Version    string `json:"version"`
	Abck       string `json:"abck"`
	Bmsz       string `json:"bmsz,omitempty"`
	ScriptHash string `json:"scriptHash,omitempty"`
	Config     string `json:"dynamicValues,omitempty"`
}

type apiResponseJson struct {
	ErrrorMessage string `json:"errorMessage"`
	Payload       string `json:"payload"`
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
func (ap *HyperSolutionsApi) GenerateWebSensor(iteration int, abck string, bmsz string) (string, error) {
	payload := generateWebSensorJson{
		UserAgent:  ap.UserAgent,
		PageUrl:    ap.Site,
		Version:    "1",
		Abck:       abck,
		ScriptHash: ap.scriptBodyHash,
	}

	if bmsz != "" {
		payload.Bmsz = bmsz
		payload.Version = "2"
	}

	if ap.Dynamic {
		payload.Config = ap.dynamicScriptConfig
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Join(ErrHyperSolutions, err)
	}

	req, err := http.NewRequest(http.MethodPost, apiGenerateWebSensorEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", errors.Join(ErrHyperSolutions, err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", ap.ApiKey)

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return "", errors.Join(ErrHyperSolutions, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.Join(ErrHyperSolutions, errors.New(resp.Status))
	}

	var respJson apiResponseJson
	err = json.NewDecoder(resp.Body).Decode(&respJson)
	if err != nil {
		return "", errors.Join(ErrHyperSolutions, err)
	}

	if respJson.ErrrorMessage != "" {
		return "", errors.Join(ErrHyperSolutions, errors.New(respJson.ErrrorMessage))
	}

	return respJson.Payload, nil
}

func (ap *HyperSolutionsApi) getDynamicScriptConfig(scriptBody []byte) error {
	payload := map[string]string{
		"script": string(scriptBody),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiScriptConfigEP, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", ap.ApiKey)

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	var respJson apiResponseJson
	err = json.NewDecoder(resp.Body).Decode(&respJson)
	if err != nil {
		return err
	}

	if respJson.ErrrorMessage != "" {
		return errors.New(respJson.ErrrorMessage)
	}

	ap.dynamicScriptConfig = respJson.Payload

	return nil
}
