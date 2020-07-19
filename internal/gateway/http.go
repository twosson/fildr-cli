package gateway

import (
	"fildr-cli/internal/config"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	httpClient *http.Client
)

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 30
)

func postGateway(data *MetricData) {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: MaxIdleConnections,
			},
			Timeout: time.Duration(RequestTimeout) * time.Second,
		}
	}

	cfg := config.Get()
	url := cfg.Gateway.Url
	if url[len(url)-1] == '/' {
		url = url + "metrics/job/" + data.job + "/instance/" + data.instance
	} else {
		url = url + "/metrics/job/" + data.job + "/instance/" + data.instance
	}

	req, err := http.NewRequest(http.MethodPost, url, data.data)
	if err != nil {
		return
	}
	req.Header.Add("blade-auth", "Bearer "+cfg.Gateway.Token)
	req.Header.Add("Content-Type", "text/plain")
	resp, err := httpClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			logger.Warnf("%d push gateway unauthorized.", resp.StatusCode)
		}
	}
	if err != nil {
		logger.Warnf("push gateway err: %v", err)
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}
