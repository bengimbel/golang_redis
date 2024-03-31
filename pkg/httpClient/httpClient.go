package httpClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	HTTPS string = "https"
	HOST  string = "api.openweathermap.org"
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

type HttpImplementor interface {
	MakeWeatherRequest(config *HttpConfig, responseStruct interface{}) error
}

// Create an instance of our client
func NewHttpClient() *HttpWeatherClient {
	return &HttpWeatherClient{
		Client: &http.Client{},
		URL: url.URL{
			Scheme: HTTPS,
			Host:   HOST,
		},
	}
}

// Custom HTTP client. This client instance has the host and scheme
// defined in it. We can dynamically pass in a config to define the path
// and query params. We also pass in a response struct for results to be applied to.
// Since we are passing in pointer to the response struct, that address in memory is
// filled in with results, and we don't need to return it. We only return an error
// if there is one.
func (hwc *HttpWeatherClient) MakeWeatherRequest(config *HttpConfig, responseStruct interface{}) error {
	query := url.Values{}

	// Loop over config query values and set them to url.Values{}
	for _, v := range config.Query {
		query.Set(v.Key, v.Value)
	}

	// encode path and query params onto the url
	hwc.URL.Path = config.Path
	hwc.URL.RawQuery = query.Encode()

	// Resolve the URI
	endpoint := hwc.URL.ResolveReference(&hwc.URL)

	// Make Request Object
	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	// Execute the request
	res, err := hwc.Client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Decode the json results to the response struct pointer we passed in.
	if err := json.NewDecoder(res.Body).Decode(&responseStruct); err != nil {
		return fmt.Errorf("Error decoding weather data: %s", err)
	}

	return nil
}
