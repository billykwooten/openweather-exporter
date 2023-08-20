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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v2"

	"github.com/codingsince1985/geo-golang/openstreetmap"
	log "github.com/sirupsen/logrus"

	"github.com/billykwooten/openweather-exporter/geo"
	"github.com/prometheus/client_golang/prometheus"
)

// OpenweatherCollector Define a struct for your collector that contains pointers
// to prometheus descriptors for each metric you wish to expose.
// Note you can also include fields of other types if they provide utility,
// but we just won't be exposing them as metrics.
var notFound = ttlcache.ErrNotFound

type Settings struct {
	ApiKey      string
	DegreesUnit string
	Language    string
	EnablePol   bool
}

type OpenweatherCollector struct {
	*Settings
	Cache     *ttlcache.Cache
	Locations []Location

	client *http.Client

	oneCallMetrics   map[string][]Metric
	pollutionMetrics map[string][]Metric
}

type Location struct {
	Location  string
	Latitude  float64
	Longitude float64
}

func resolveLocations(locations string) []Location {
	var res []Location

	for _, location := range strings.Split(locations, "|") {
		// Get Coords.
		latitude, longitude, err := geo.GetCoords(openstreetmap.Geocoder(), location)
		if err != nil {
			log.Fatal("failed to resolve location:", err)
		}
		res = append(res, Location{Location: location, Latitude: latitude, Longitude: longitude})
	}
	return res
}

// NewOpenweatherCollector You must create a constructor for your collector that
// initializes every descriptor and returns a pointer to the collector
func NewOpenweatherCollector(settings *Settings, locationsStr string, cache *ttlcache.Cache) *OpenweatherCollector {
	locations := resolveLocations(locationsStr)

	oneCallMetrics := make(map[string][]Metric)
	pollutionMetrics := make(map[string][]Metric)
	for _, loc := range locations {
		oneCallMetrics[loc.Location] = OneCallGauges(loc.Location)

		if settings.EnablePol {
			pollutionMetrics[loc.Location] = PollutionGauges(loc.Location)
		}
	}

	return &OpenweatherCollector{
		Settings:  settings,
		Locations: locations,
		Cache:     cache,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		oneCallMetrics:   oneCallMetrics,
		pollutionMetrics: pollutionMetrics,
	}
}

// Describe Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *OpenweatherCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metrics := range collector.oneCallMetrics {
		for _, metric := range metrics {
			ch <- metric.Desc()
		}
	}

	for _, metrics := range collector.pollutionMetrics {
		for _, metric := range metrics {
			ch <- metric.Desc()
		}
	}
}

// Collect implements required collect function for all prometheus collectors
func (collector *OpenweatherCollector) Collect(ch chan<- prometheus.Metric) {
	for _, location := range collector.Locations {
		collector.collectOneCall(location, ch)

		if collector.Settings.EnablePol {
			collector.collectPollution(location, ch)
		}
	}
}

func cachedHttpRequest[T any](collector *OpenweatherCollector, key string, request func() (T, error)) (T, error) {
	if val, err := collector.Cache.Get(key); err != notFound || val != nil {
		// Grab Metrics from cache
		return val.(T), nil
	} else {
		// Grab Metrics
		w, err := request()
		if err != nil {
			return w, err
		}
		err = collector.Cache.Set(key, w)
		if err != nil {
			return w, fmt.Errorf("Could not set cache data. %s", err.Error())
		}
		return w, nil
	}
}

func (collector *OpenweatherCollector) collectOneCall(location Location, ch chan<- prometheus.Metric) {
	w, err := cachedHttpRequest(collector, location.Location+":onecall",
		func() (*OneCallCurrentData, error) {
			return CurrentByCoordinates(location, collector.client, collector.Settings)
		},
	)

	if err != nil {
		log.Infof("Collecting metrics failed for %s: %s", location.Location, err.Error())
		return
	}

	// Write the latest value for each metric in the prometheus metric channel.
	for _, metric := range collector.oneCallMetrics[location.Location] {
		ch <- metric.FromResponse(w)
	}
}

func (collector *OpenweatherCollector) collectPollution(location Location, ch chan<- prometheus.Metric) {
	w, err := cachedHttpRequest(collector, location.Location+":pollution",
		func() (*PollutionData, error) {
			return PollutionByCoordinates(location, collector.client, collector.Settings)
		},
	)

	if err != nil {
		log.Infof("Collecting metrics failed for %s: %s", location.Location, err.Error())
		return
	}

	for _, metric := range collector.pollutionMetrics[location.Location] {
		ch <- metric.FromResponse(w)
	}
}
