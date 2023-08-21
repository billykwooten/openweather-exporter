// Copyright 2023 Artem Tarasov
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	DataUnits = map[string]string{"C": "metric", "F": "imperial", "K": "internal"}

	apiCallCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "openweather_api_calls_total",
		Help: "Number of API calls to openweathermap.org",
	}, []string{"location", "endpoint", "response_status"})
)

func init() {
	prometheus.MustRegister(apiCallCounter)
}

func CurrentByCoordinates(loc Location, client *http.Client, settings *Settings) (*OneCallCurrentData, error) {
	var onecall OneCallData

	units, ok := DataUnits[settings.DegreesUnit]
	if !ok {
		return nil, fmt.Errorf("unknown unit %s (must be C, F, or K)", settings.DegreesUnit)
	}

	endpoint := "https://api.openweathermap.org/data/3.0/onecall"

	q := url.Values{}
	q.Set("appid", settings.ApiKey)
	q.Set("lat", fmt.Sprint(loc.Latitude))
	q.Set("lon", fmt.Sprint(loc.Longitude))
	q.Set("units", units)
	q.Set("lang", settings.Language)
	q.Set("excludes", "minutely,hourly,daily,alerts")

	u, _ := url.Parse(endpoint)
	u.RawQuery = q.Encode()

	response, err := client.Get(u.String())

	if response != nil {
		apiCallCounter.WithLabelValues(loc.Location, endpoint, response.Status).Inc()
	}

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if bytes, err := io.ReadAll(response.Body); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &onecall); err != nil {
		return nil, err
	}

	return &onecall.Current, nil
}

func PollutionByCoordinates(loc Location, client *http.Client, settings *Settings) (*PollutionData, error) {
	var pollution Pollution

	endpoint := "https://api.openweathermap.org/data/2.5/air_pollution"

	q := url.Values{}
	q.Set("appid", settings.ApiKey)
	q.Set("lat", fmt.Sprint(loc.Latitude))
	q.Set("lon", fmt.Sprint(loc.Longitude))

	u, _ := url.Parse(endpoint)
	u.RawQuery = q.Encode()

	response, err := client.Get(u.String())

	if response != nil {
		apiCallCounter.WithLabelValues(loc.Location, endpoint, response.Status).Inc()
	}

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if bytes, err := io.ReadAll(response.Body); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &pollution); err != nil {
		return nil, fmt.Errorf("response: %s; error: %s", string(bytes), err.Error())
	}

	return &pollution.List[0], nil
}
