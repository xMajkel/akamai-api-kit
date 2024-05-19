package api

type Provider interface {
	GenerateWebSensor(int, string, string) (string, error)
	SetScriptUrl(string)
	SetScriptBody([]byte) error
}

type Config struct {
	ApiKey     string
	UserAgent  string
	Site       string
	Dynamic    bool
	ScriptUrl  string
	ScriptBody []byte
}

// NewApiConfig returns new config for the API providers, it is required to create api provider implementing the Provider interface
//
// Parameters:
//   - apiKey: API provider key
//   - userAgent: User-Agent header value
//   - site: URL of Akamai protected site, eg. https://www.example.com/login/
//   - dynamic: If Akamai config is using dynamic scripts set this to true
//
// Returns:
//   - *Config: Config for the API providers
func NewApiConfig(apiKey string, userAgent string, site string, dynamic bool) *Config {
	c := &Config{
		ApiKey:    apiKey,
		UserAgent: userAgent,
		Site:      site,
		Dynamic:   dynamic,
	}

	return c
}
