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

	temperatureMetric *prometheus.Desc
	humidity          *prometheus.Desc
	feelslike         *prometheus.Desc
	pressure          *prometheus.Desc
	windspeed         *prometheus.Desc
	rain1h            *prometheus.Desc
	snow1h            *prometheus.Desc
	winddegree        *prometheus.Desc
	cloudiness        *prometheus.Desc
	sunrise           *prometheus.Desc
	sunset            *prometheus.Desc
	currentconditions *prometheus.Desc
	uvi               *prometheus.Desc
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
func NewOpenweatherCollector(settings *Settings, locations string, cache *ttlcache.Cache) *OpenweatherCollector {

	return &OpenweatherCollector{
		Settings:  settings,
		Locations: resolveLocations(locations),
		Cache:     cache,
		temperatureMetric: prometheus.NewDesc("openweather_temperature",
			"Current temperature in degrees",
			[]string{"location"}, nil,
		),
		humidity: prometheus.NewDesc("openweather_humidity",
			"Current relative humidity",
			[]string{"location"}, nil,
		),
		feelslike: prometheus.NewDesc("openweather_feelslike",
			"Current feels_like temperature in degrees",
			[]string{"location"}, nil,
		),
		pressure: prometheus.NewDesc("openweather_pressure",
			"Current Atmospheric pressure hPa",
			[]string{"location"}, nil,
		),
		windspeed: prometheus.NewDesc("openweather_windspeed",
			"Current Wind Speed in mph or meters/sec if imperial",
			[]string{"location"}, nil,
		),
		rain1h: prometheus.NewDesc("openweather_rain1h",
			"Rain volume for last hour, in millimeters",
			[]string{"location"}, nil,
		),
		snow1h: prometheus.NewDesc("openweather_snow1h",
			"Snow volume for last hour, in millimeters",
			[]string{"location"}, nil,
		),
		winddegree: prometheus.NewDesc("openweather_winddegree",
			"Wind direction, degrees (meteorological)",
			[]string{"location"}, nil,
		),
		cloudiness: prometheus.NewDesc("openweather_cloudiness",
			"Cloudiness percentage",
			[]string{"location"}, nil,
		),
		sunrise: prometheus.NewDesc("openweather_sunrise",
			"Sunrise time, unix, UTC",
			[]string{"location"}, nil,
		),
		sunset: prometheus.NewDesc("openweather_sunset",
			"Sunset time, unix, UTC",
			[]string{"location"}, nil,
		),
		currentconditions: prometheus.NewDesc("openweather_currentconditions",
			"Current weather conditions",
			[]string{"location", "currentconditions"}, nil,
		),
		uvi: prometheus.NewDesc("openweather_ultraviolet_index",
			"Ultraviolet Index",
			[]string{"location"}, nil,
		),
	}
}

// Describe Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *OpenweatherCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the metric you create for a given collector
	ch <- collector.temperatureMetric
	ch <- collector.humidity
	ch <- collector.feelslike
	ch <- collector.pressure
	ch <- collector.windspeed
	ch <- collector.rain1h
	ch <- collector.winddegree
	ch <- collector.cloudiness
	ch <- collector.sunrise
	ch <- collector.sunset
	ch <- collector.currentconditions
	ch <- collector.uvi
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

		// Get Weather description out of Weather slice to pass as label
		var weatherDescription string
		for _, n := range w.Weather {
			weatherDescription = n.Description
		}

		// Write the latest value for each metric in the prometheus metric channel.
		// Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
		ch <- prometheus.MustNewConstMetric(collector.temperatureMetric, prometheus.GaugeValue, w.Temp, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.humidity, prometheus.GaugeValue, float64(w.Humidity), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.feelslike, prometheus.GaugeValue, w.FeelsLike, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.pressure, prometheus.GaugeValue, float64(w.Pressure), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.windspeed, prometheus.GaugeValue, w.WindSpeed, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.rain1h, prometheus.GaugeValue, w.Rain.OneH, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.winddegree, prometheus.GaugeValue, w.WindDeg, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.cloudiness, prometheus.GaugeValue, float64(w.Clouds), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.sunrise, prometheus.GaugeValue, float64(w.Sunrise), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.sunset, prometheus.GaugeValue, float64(w.Sunset), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.snow1h, prometheus.GaugeValue, w.Snow.OneH, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.currentconditions, prometheus.GaugeValue, 0, location.Location, weatherDescription)
		ch <- prometheus.MustNewConstMetric(collector.uvi, prometheus.GaugeValue, w.UVI, location.Location)
	}
}
