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
	owm "github.com/briandowns/openweathermap"
	"github.com/prometheus/client_golang/prometheus"
)

// OpenweatherCollector Define a struct for your collector that contains pointers
// to prometheus descriptors for each metric you wish to expose.
// Note you can also include fields of other types if they provide utility,
// but we just won't be exposing them as metrics.
var notFound = ttlcache.ErrNotFound

type OpenweatherCollector struct {
	ApiKey            string
	Cache             *ttlcache.Cache
	enablePol         bool
	enableUV          bool
	DegreesUnit       string
	Language          string
	Locations         []Location
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
	aqi               *prometheus.Desc
	Co                *prometheus.Desc
	No                *prometheus.Desc
	No2               *prometheus.Desc
	O3                *prometheus.Desc
	So2               *prometheus.Desc
	Pm25              *prometheus.Desc
	Pm10              *prometheus.Desc
	Nh3               *prometheus.Desc
	UVI               *prometheus.Desc
}

type Location struct {
	Location      string
	Latitude      float64
	Longitude     float64
	CacheKeyOWM   string
	CacheKeyPOWM  string
	CacheKeyUVOWM string
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
		cacheKeyPOWM := fmt.Sprintf("POWM %s", location)
		cacheKeyUVOWM := fmt.Sprintf("UVOWM %s", location)
		res = append(res, Location{Location: location, Latitude: latitude, Longitude: longitude, CacheKeyOWM: cacheKeyOWM, CacheKeyPOWM: cacheKeyPOWM, CacheKeyUVOWM: cacheKeyUVOWM})
	}
	return res
}

// NewOpenweatherCollector You must create a constructor for your collector that
// initializes every descriptor and returns a pointer to the collector
func NewOpenweatherCollector(degreesUnit string, language string, apikey string, locations string, cache *ttlcache.Cache, enablePol bool, enableUV bool) *OpenweatherCollector {

	return &OpenweatherCollector{
		ApiKey:      apikey,
		DegreesUnit: degreesUnit,
		Language:    language,
		Locations:   resolveLocations(locations),
		Cache:       cache,
		enablePol:   enablePol,
		enableUV:    enableUV,
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
		aqi: prometheus.NewDesc("openweather_pollution_airqualityindex",
			"Air Quality Index. 1 = Good, 2 = Fair, 3 = Moderate, 4 = Poor, 5 = Very Poor.",
			[]string{"location"}, nil,
		),
		Co: prometheus.NewDesc("openweather_pollution_carbonmonoxide",
			"Concentration of CO (Carbon Monoxide) μg/m3",
			[]string{"location"}, nil,
		),
		No: prometheus.NewDesc("openweather_pollution_nitrogenmonoxide",
			"Concentration of NO (Nitrogen Monoxide) μg/m3",
			[]string{"location"}, nil,
		),
		No2: prometheus.NewDesc("openweather_pollution_nitrogendioxide",
			"Concentration of NO2 (Nitrogen Dioxide) μg/m3",
			[]string{"location"}, nil,
		),
		O3: prometheus.NewDesc("openweather_pollution_ozone",
			"Concentration of O3 (Ozone) μg/m3",
			[]string{"location"}, nil,
		),
		So2: prometheus.NewDesc("openweather_pollution_sulphurdioxide",
			"Concentration of SO2 (Sulphur Dioxide) μg/m3",
			[]string{"location"}, nil,
		),
		Pm25: prometheus.NewDesc("openweather_pollution_pm25",
			"Concentration of PM2.5 (Fine particles matter) μg/m3",
			[]string{"location"}, nil,
		),
		Pm10: prometheus.NewDesc("openweather_pollution_pm10",
			"Concentration of PM10 (Coarse particles matter) μg/m3",
			[]string{"location"}, nil,
		),
		Nh3: prometheus.NewDesc("openweather_pollution_nh3",
			"Concentration of NH3 (Ammonia) μg/m3",
			[]string{"location"}, nil,
		),
		UVI: prometheus.NewDesc("openweather_ultraviolet_index",
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
	ch <- collector.aqi
	ch <- collector.Co
	ch <- collector.No
	ch <- collector.No2
	ch <- collector.O3
	ch <- collector.So2
	ch <- collector.Pm25
	ch <- collector.Pm10
	ch <- collector.Nh3
	ch <- collector.UVI

}

// Collect implements required collect function for all prometheus collectors
func (collector *OpenweatherCollector) Collect(ch chan<- prometheus.Metric) {
	for _, location := range collector.Locations {
		var w *owm.CurrentWeatherData
		var pd *owm.Pollution
		var uuv *owm.UV

		// Setup HTTP Client
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		if val, err := collector.Cache.Get(location.CacheKeyOWM); err != notFound || val != nil {
			// Grab Metrics from cache
			w = val.(*owm.CurrentWeatherData)
			// Grab pollution metrics from cache if enabled
			if collector.enablePol == true {
				if pval, err := collector.Cache.Get(location.CacheKeyPOWM); err != notFound || pval != nil {
					pd = pval.(*owm.Pollution)
				}
			}
			if collector.enableUV == true {
				if uvval, err := collector.Cache.Get(location.CacheKeyUVOWM); err != notFound || uvval != nil {
					uuv = uvval.(*owm.UV)
				}
			}
		} else {
			// Grab Metrics
			w, err = owm.NewCurrent(collector.DegreesUnit, collector.Language, collector.ApiKey, owm.WithHttpClient(client))
			if err != nil {
				log.Fatal("invalid openweather API configuration:", err)
			}
			err = w.CurrentByCoordinates(&owm.Coordinates{Latitude: location.Latitude, Longitude: location.Longitude})
			if err != nil {
				log.Infof("Collecting metrics failed for %s: %s", location.Location, err.Error())
				continue
			}
			err = collector.Cache.Set(location.CacheKeyOWM, w)
			if err != nil {
				log.Infof("Could not set cache data. %s", err.Error())
				continue
			}
			if collector.enablePol == true {
				pd, err = owm.NewPollution(collector.ApiKey, owm.WithHttpClient(client))
				if err != nil {
					log.Warnf("Collecting pollution metrics failed for %s: %s", location.Location, err.Error())
					continue
				}
				err = pd.PollutionByParams(&owm.PollutionParameters{Location: owm.Coordinates{Latitude: location.Latitude, Longitude: location.Longitude}})
				if err != nil {
					log.Infof("Collecting pollution metrics failed for %s: %s", location.Location, err.Error())
					continue
				}
				err = collector.Cache.Set(location.CacheKeyPOWM, pd)
				if err != nil {
					log.Infof("Could not set pollution cache data. %s", err.Error())
					continue
				}
			}
			if collector.enableUV == true {
				uuv, err = owm.NewUV(collector.ApiKey, owm.WithHttpClient(client))
				if err != nil {
					log.Warnf("Collecting UV metrics failed for %s: %s", location.Location, err.Error())
					continue
				}
				err = uuv.Current(&owm.Coordinates{Latitude: location.Latitude, Longitude: location.Longitude})
				if err != nil {
					log.Infof("Collecting UV metrics failed for %s: %s", location.Location, err.Error())
					continue
				}
				err = collector.Cache.Set(location.CacheKeyUVOWM, uuv)
				if err != nil {
					log.Infof("Could not set UV cache data. %s", err.Error())
					continue
				}
			}
		}

		// Get Weather description out of Weather slice to pass as label
		var weatherDescription string
		for _, n := range w.Weather {
			weatherDescription = n.Description
		}

		// Write the latest value for each metric in the prometheus metric channel.
		// Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
		ch <- prometheus.MustNewConstMetric(collector.temperatureMetric, prometheus.GaugeValue, w.Main.Temp, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.humidity, prometheus.GaugeValue, float64(w.Main.Humidity), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.feelslike, prometheus.GaugeValue, w.Main.FeelsLike, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.pressure, prometheus.GaugeValue, w.Main.Pressure, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.windspeed, prometheus.GaugeValue, w.Wind.Speed, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.rain1h, prometheus.GaugeValue, w.Rain.OneH, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.winddegree, prometheus.GaugeValue, w.Wind.Deg, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.cloudiness, prometheus.GaugeValue, float64(w.Clouds.All), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.sunrise, prometheus.GaugeValue, float64(w.Sys.Sunrise), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.sunset, prometheus.GaugeValue, float64(w.Sys.Sunset), location.Location)
		ch <- prometheus.MustNewConstMetric(collector.snow1h, prometheus.GaugeValue, w.Snow.OneH, location.Location)
		ch <- prometheus.MustNewConstMetric(collector.currentconditions, prometheus.GaugeValue, 0, location.Location, weatherDescription)
		if collector.enablePol == true {
			ch <- prometheus.MustNewConstMetric(collector.aqi, prometheus.GaugeValue, pd.List[0].Main.Aqi, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.Co, prometheus.GaugeValue, pd.List[0].Components.Co, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.No, prometheus.GaugeValue, pd.List[0].Components.No, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.No2, prometheus.GaugeValue, pd.List[0].Components.No2, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.O3, prometheus.GaugeValue, pd.List[0].Components.O3, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.So2, prometheus.GaugeValue, pd.List[0].Components.So2, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.Pm25, prometheus.GaugeValue, pd.List[0].Components.Pm25, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.Pm10, prometheus.GaugeValue, pd.List[0].Components.Pm10, location.Location)
			ch <- prometheus.MustNewConstMetric(collector.Nh3, prometheus.GaugeValue, pd.List[0].Components.Nh3, location.Location)
		}
		if collector.enableUV == true {
			ch <- prometheus.MustNewConstMetric(collector.UVI, prometheus.GaugeValue, uuv.Value, location.Location)
		}
	}
}
