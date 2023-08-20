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
}

type OpenweatherCollector struct {
	*Settings
	Cache     *ttlcache.Cache
	Locations []Location

	gaugesByLocation map[string][]Gauge
}

type Location struct {
	Location    string
	Latitude    float64
	Longitude   float64
	CacheKeyOWM string
}

func resolveLocations(locations string) []Location {
	var res []Location

	for _, location := range strings.Split(locations, "|") {
		// Get Coords.
		latitude, longitude, err := geo.GetCoords(openstreetmap.Geocoder(), location)
		if err != nil {
			log.Fatal("failed to resolve location:", err)
		}
		cacheKeyOWM := fmt.Sprintf("OWM %s", location)
		res = append(res, Location{Location: location, Latitude: latitude, Longitude: longitude, CacheKeyOWM: cacheKeyOWM})
	}
	return res
}

// NewOpenweatherCollector You must create a constructor for your collector that
// initializes every descriptor and returns a pointer to the collector
func NewOpenweatherCollector(settings *Settings, locationsStr string, cache *ttlcache.Cache) *OpenweatherCollector {
	locations := resolveLocations(locationsStr)

	gauges := make(map[string][]Gauge)
	for _, loc := range locations {
		gauges[loc.Location] = OneCallGauges(loc.Location)
	}

	return &OpenweatherCollector{
		Settings:  settings,
		Locations: locations,
		Cache:     cache,

		gaugesByLocation: gauges,
	}
}

// Describe Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *OpenweatherCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, gauges := range collector.gaugesByLocation {
		for _, gauge := range gauges {
			ch <- gauge.Desc()
		}
	}
}

// Collect implements required collect function for all prometheus collectors
func (collector *OpenweatherCollector) Collect(ch chan<- prometheus.Metric) {
	for _, location := range collector.Locations {
		var w *OneCallCurrentData

		// Setup HTTP Client
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		if val, err := collector.Cache.Get(location.CacheKeyOWM); err != notFound || val != nil {
			// Grab Metrics from cache
			w = val.(*OneCallCurrentData)
		} else {
			// Grab Metrics
			w, err = CurrentByCoordinates(location, client, collector.Settings)
			if err != nil {
				log.Infof("Collecting metrics failed for %s: %s", location.Location, err.Error())
				continue
			}
			err = collector.Cache.Set(location.CacheKeyOWM, w)
			if err != nil {
				log.Infof("Could not set cache data. %s", err.Error())
				continue
			}
		}

		// Write the latest value for each metric in the prometheus metric channel.
		for _, gauge := range collector.gaugesByLocation[location.Location] {
			ch <- gauge.FromResponse(w)
		}
	}
}
