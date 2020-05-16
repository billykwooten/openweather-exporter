// Copyright 2020 Billy Wooten
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
	"log"
	"net/http"
	"time"

	"github.com/billykwooten/openweather-exporter/geo"
	owm "github.com/billykwooten/openweathermap"
	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/prometheus/client_golang/prometheus"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type OpenweatherCollector struct {
	ApiKey            string
	DegreesUnit       string
	Language          string
	Location          string
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
}

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func NewOpenweatherCollector(degressUnit string, language string, apikey string, location string) *OpenweatherCollector {
	return &OpenweatherCollector{
		ApiKey:      apikey,
		DegreesUnit: degressUnit,
		Language:    language,
		Location:    location,
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
			[]string{"location", "currentconditions"}, nil,
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
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *OpenweatherCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
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

}

//Collect implements required collect function for all prometheus collectors
func (collector *OpenweatherCollector) Collect(ch chan<- prometheus.Metric) {
	// Get Coords
	latitude, longitude := geo.Get_coords(openstreetmap.Geocoder(), collector.Location)

	// Setup HTTP Client
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Grab Metrics
	w, err := owm.NewCurrent(collector.DegreesUnit, collector.Language, collector.ApiKey, owm.WithHttpClient(client))
	if err != nil {
		log.Fatalln(err)
	}

	w.CurrentByCoordinates(&owm.Coordinates{Longitude: longitude, Latitude: latitude})

	// Get Weather description out of Weather slice to pass as label
	var weather_description string
	for _, n := range w.Weather {
		weather_description = n.Description
	}

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(collector.temperatureMetric, prometheus.GaugeValue, w.Main.Temp, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.humidity, prometheus.GaugeValue, float64(w.Main.Humidity), collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.feelslike, prometheus.GaugeValue, w.Main.FeelsLike, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.pressure, prometheus.GaugeValue, w.Main.Pressure, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.windspeed, prometheus.GaugeValue, w.Wind.Speed, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.rain1h, prometheus.GaugeValue, w.Rain.OneH, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.winddegree, prometheus.GaugeValue, w.Wind.Deg, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.cloudiness, prometheus.GaugeValue, float64(w.Clouds.All), collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.sunrise, prometheus.GaugeValue, float64(w.Sys.Sunrise), collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.sunset, prometheus.GaugeValue, float64(w.Sys.Sunset), collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.snow1h, prometheus.GaugeValue, w.Snow.OneH, collector.Location)
	ch <- prometheus.MustNewConstMetric(collector.currentconditions, prometheus.GaugeValue, 0, collector.Location, weather_description)
}
