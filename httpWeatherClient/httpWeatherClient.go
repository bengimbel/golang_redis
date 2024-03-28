package httpWeatherClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type QueryParams struct {
	Key   string
	Value string
}

type HttpConfig struct {
	Path  string
	Query []QueryParams
}

type HttpWeatherClient struct {
	Client *http.Client
	ApiKey string
	URL    url.URL
}

func NewHttpClient() *HttpWeatherClient {
	return &HttpWeatherClient{
		Client: &http.Client{},
		URL: url.URL{
			Scheme: "https",
			Host:   "api.openweathermap.org",
		},
	}
}

func (hwc *HttpWeatherClient) MakeWeatherRequest(config *HttpConfig, responseStruct interface{}) error {
	query := url.Values{}
	for _, v := range config.Query {
		query.Set(v.Key, v.Value)
	}
	hwc.URL.Path = config.Path
	hwc.URL.RawQuery = query.Encode()

	endpoint := hwc.URL.ResolveReference(&hwc.URL)

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	res, err := hwc.Client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&responseStruct); err != nil {
		return fmt.Errorf("Error decoding weather data: %s", err)
	}

	return nil
}
