// Copyright 2023 Billy Wooten
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
	log "github.com/sirupsen/logrus"
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

	log.Infof("Gathering Metrics from Openweather API 3.0 for %s, Lat:%f, Lon:%f", loc.Location, loc.Latitude, loc.Longitude)
	response, err := client.Get(u.String())

	// Success is indicated with 2xx status codes:
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		log.Fatal("Non-OK HTTP status when hitting openweather API, is your API Key correct and did you sign up for the 3.0 API plan?"+" Response given was: ", response.Status+". If you have not signed up for the free or paid subscription for the 3.0 API, please see https://openweathermap.org/price, after activation it might take 1-4 hours for their API to accept your API key, there is nothing I can do about this as it's server-side.")
	}

	if response != nil {
		apiCallCounter.WithLabelValues(loc.Location, endpoint, response.Status).Inc()
	}

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)
	if bytes, err := io.ReadAll(response.Body); err != nil {
		return nil, err
	} else if err := json.Unmarshal(bytes, &pollution); err != nil {
		return nil, fmt.Errorf("response: %s; error: %s", string(bytes), err.Error())
	}

	return &pollution.List[0], nil
}
