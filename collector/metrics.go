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
	"github.com/prometheus/client_golang/prometheus"
)

type Gauge interface {
	Desc() *prometheus.Desc
	FromResponse(any) prometheus.Metric
}

type OCData = OneCallCurrentData

type OCGauge struct {
	desc               *prometheus.Desc
	extractValue       func(*OCData) float64
	extractLabelValues func(*OCData) []string
}

func (g *OCGauge) Desc() *prometheus.Desc {
	return g.desc
}

func (g *OCGauge) FromResponse(data any) prometheus.Metric {
	d := data.(*OCData)
	return prometheus.MustNewConstMetric(
		g.desc,
		prometheus.GaugeValue,
		g.extractValue(d),
		g.extractLabelValues(d)...,
	)
}

func OneCallGauges(location string) []Gauge {
	makeGauge := func(name, description string, extract func(*OCData) float64) *OCGauge {
		return &OCGauge{
			prometheus.NewDesc(name, description, []string{"location"}, nil),
			extract,
			func(*OCData) []string { return []string{location} },
		}
	}

	return []Gauge{
		makeGauge("openweather_temperature", "Current temperature in degrees",
			func(d *OCData) float64 { return d.Temp },
		),
		makeGauge("openweather_humidity", "Current relative humidity",
			func(d *OCData) float64 { return float64(d.Humidity) },
		),
		makeGauge("openweather_feelslike", "Current feels_like temperature in degrees",
			func(d *OCData) float64 { return d.FeelsLike },
		),
		makeGauge("openweather_pressure", "Current Atmospheric pressure hPa",
			func(d *OCData) float64 { return float64(d.Pressure) },
		),
		makeGauge("openweather_windspeed", "Current Wind Speed in mph or meters/sec if imperial",
			func(d *OCData) float64 { return d.WindSpeed },
		),
		makeGauge("openweather_rain1h", "Rain volume for last hour, in millimeters",
			func(d *OCData) float64 { return d.Rain.OneH },
		),
		makeGauge("openweather_snow1h", "Snow volume for last hour, in millimeters",
			func(d *OCData) float64 { return d.Snow.OneH },
		),
		makeGauge("openweather_winddegree", "Wind direction, degrees (meteorological)",
			func(d *OCData) float64 { return d.WindDeg },
		),
		makeGauge("openweather_cloudiness", "Cloudiness percentage",
			func(d *OCData) float64 { return float64(d.Clouds) },
		),
		makeGauge("openweather_sunrise", "Sunrise time, unix, UTC",
			func(d *OCData) float64 { return float64(d.Sunrise) },
		),
		makeGauge("openweather_sunset", "Sunset time, unix, UTC",
			func(d *OCData) float64 { return float64(d.Sunset) },
		),
		makeGauge("openweather_ultraviolet_index", "Ultraviolet Index",
			func(d *OCData) float64 { return d.UVI },
		),
		&OCGauge{
			prometheus.NewDesc("openweather_currentconditions",
				"Current weather conditions",
				[]string{"location", "currentconditions"}, nil,
			),
			func(*OCData) float64 { return 0 },
			func(d *OCData) []string {
				// Get Weather description out of Weather slice to pass as label
				var weatherDescription string
				for _, n := range d.Weather {
					weatherDescription = n.Description
				}
				return []string{location, weatherDescription}
			},
		},
	}
}
