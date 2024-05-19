# Akamai API kit
GO package to help with Akamai Bot Manager (web) using most popular API providers.
## Disclaimer
This repository is intended for educational and research purposes only. Please use it responsibly and ethically. While you're free to use this code as you wish under the MIT License, I strongly advise against using it for any malicious activities or violating any website's terms of service.
## Getting started
Before you start make sure you are using proper **TLS** client.
I recommend [bogdanfinn's TLS](https://www.github.com/bogdanfinn/tls-client)

Another very important thing is headers order. If the site is protected by Akamai, it pays close a close attention to headers and their order. You cannot rely on Chrome's "DevTools" as they don't show you the true order of the headers sent, for this you'll need to pass your network traffic through a reverse proxy. Here I can suggest [Charles](https://www.charlesproxy.com/)
## How to solve web Akamai
 1. Fetch site protected by Akamai
 2. Create api.Config
 ```go
apiConf:=api.NewApiConfig(
	"api-key", // api key from one of the providers
	"user-agent", // User-Agent header
	"https://www.example.com/login/", 
	false, // true if site uses dynamic script body
)
 ```
 3. Select API provider
 ```go
provider := jevi.NewApi(apiConf)
//or
provider := fdis.NewApi(apiConf)
//or
provider := hawk.NewApi(apiConf)
//or
provider := hypersolutions.NewApi(apiConf)
```
 4. Scrape script path using response body from point number 1
```go
scriptPath,err := utility.ScrapeScriptPath(resp.Body)
if err != nil {
  //No script path found
  return
}

provider.SetScriptUrl(scriptUrl)
```
 5. Fetch script body using `https://www.example.com + scriptPath`
```go
body, _ := io.ReadAll(resp.Body)
provider.SetScriptBody(body)
```
 6. Generate sensor and post it to `scriptPath`
 ```go
for i:=0;i<3;i++{
  sensor, err := provider.GenerateWebSensor(i, "_abck cookie value", "bm_sz cookie value")
  if err !=  nil {
    return
  }
  // POST `{"sensor_data":sensor}` to scriptUrl
  
  // Stop if cookie is valid
  if validator.IsCookieValid("new _abck cookie value"){
    break
  }
}
if !validator.IsCookieValid("new _abck cookie value"){
  return
}
```

> Note: Checking if _abck cookie is valid doesn't work on every website. If you observe single digit after first `~` in _abck change from -1 to 0 in browser, you can most probably use the method above.

 7. Now you should be able to successfully interact with protected endpoint
## _abck cookie invalidation
SometSometimes, protected endpoints return a new _abck cookie that will no longer be valid.
Use `validator.IsCookieNoLongerValid("_abck cookie value)` to check if that is the case, if yes, then you should be able to get a new valid one after generating and posting one sensor.