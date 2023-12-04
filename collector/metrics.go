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
	"github.com/prometheus/client_golang/prometheus"
)

type Metric interface {
	Desc() *prometheus.Desc
	FromResponse(any) prometheus.Metric
}

type ApiResponse interface {
	*OneCallCurrentData | *PollutionData
}

type Gauge[T ApiResponse] struct {
	desc               *prometheus.Desc
	extractValue       func(T) float64
	extractLabelValues func(T) []string
}

func (g *Gauge[T]) Desc() *prometheus.Desc {
	return g.desc
}

func (g *Gauge[T]) FromResponse(data any) prometheus.Metric {
	d := data.(T)
	return prometheus.MustNewConstMetric(
		g.desc,
		prometheus.GaugeValue,
		g.extractValue(d),
		g.extractLabelValues(d)...,
	)
}

func OneCallGauges(location string) []Metric {
	makeGauge := func(name, description string, extract func(*OneCallCurrentData) float64) *Gauge[*OneCallCurrentData] {
		return &Gauge[*OneCallCurrentData]{
			prometheus.NewDesc(name, description, []string{"location"}, nil),
			extract,
			func(*OneCallCurrentData) []string { return []string{location} },
		}
	}

	return []Metric{
		makeGauge("openweather_temperature", "Current temperature in degrees",
			func(d *OneCallCurrentData) float64 { return d.Temp },
		),
		makeGauge("openweather_humidity", "Current relative humidity",
			func(d *OneCallCurrentData) float64 { return float64(d.Humidity) },
		),
		makeGauge("openweather_feelslike", "Current feels_like temperature in degrees",
			func(d *OneCallCurrentData) float64 { return d.FeelsLike },
		),
		makeGauge("openweather_pressure", "Current Atmospheric pressure hPa",
			func(d *OneCallCurrentData) float64 { return float64(d.Pressure) },
		),
		makeGauge("openweather_windspeed", "Current Wind Speed in mph or meters/sec if imperial",
			func(d *OneCallCurrentData) float64 { return d.WindSpeed },
		),
		makeGauge("openweather_rain1h", "Rain volume for last hour, in millimeters",
			func(d *OneCallCurrentData) float64 { return d.Rain.OneH },
		),
		makeGauge("openweather_snow1h", "Snow volume for last hour, in millimeters",
			func(d *OneCallCurrentData) float64 { return d.Snow.OneH },
		),
		makeGauge("openweather_winddegree", "Wind direction, degrees (meteorological)",
			func(d *OneCallCurrentData) float64 { return d.WindDeg },
		),
		makeGauge("openweather_cloudiness", "Cloudiness percentage",
			func(d *OneCallCurrentData) float64 { return float64(d.Clouds) },
		),
		makeGauge("openweather_sunrise", "Sunrise time, unix, UTC",
			func(d *OneCallCurrentData) float64 { return float64(d.Sunrise) },
		),
		makeGauge("openweather_sunset", "Sunset time, unix, UTC",
			func(d *OneCallCurrentData) float64 { return float64(d.Sunset) },
		),
		makeGauge("openweather_ultraviolet_index", "Ultraviolet Index",
			func(d *OneCallCurrentData) float64 { return d.UVI },
		),
		&Gauge[*OneCallCurrentData]{
			prometheus.NewDesc("openweather_currentconditions",
				"Current weather conditions",
				[]string{"location", "currentconditions"}, nil,
			),
			func(*OneCallCurrentData) float64 { return 0 },
			func(d *OneCallCurrentData) []string {
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

func PollutionGauges(location string) []Metric {
	makeGauge := func(name, description string, extract func(*PollutionData) float64) *Gauge[*PollutionData] {
		return &Gauge[*PollutionData]{
			prometheus.NewDesc(name, description, []string{"location"}, nil),
			extract,
			func(*PollutionData) []string { return []string{location} },
		}
	}

	return []Metric{
		makeGauge("openweather_pollution_airqualityindex", "Air Quality Index. 1 = Good, 2 = Fair, 3 = Moderate, 4 = Poor, 5 = Very Poor.",
			func(pd *PollutionData) float64 { return pd.Main.Aqi },
		),
		makeGauge("openweather_pollution_carbonmonoxide", "Concentration of CO (Carbon Monoxide) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.Co },
		),
		makeGauge("openweather_pollution_nitrogenmonoxide", "Concentration of NO (Nitrogen Monoxide) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.No },
		),
		makeGauge("openweather_pollution_nitrogendioxide", "Concentration of NO2 (Nitrogen Dioxide) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.No2 },
		),
		makeGauge("openweather_pollution_ozone", "Concentration of O3 (Ozone) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.O3 },
		),
		makeGauge("openweather_pollution_sulphurdioxide", "Concentration of SO2 (Sulphur Dioxide) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.So2 },
		),
		makeGauge("openweather_pollution_pm25", "Concentration of PM2.5 (Fine particles matter) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.Pm25 },
		),
		makeGauge("openweather_pollution_pm10", "Concentration of PM10 (Coarse particles matter) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.Pm10 },
		),
		makeGauge("openweather_pollution_nh3", "Concentration of NH3 (Ammonia) μg/m3",
			func(pd *PollutionData) float64 { return pd.Components.Nh3 },
		),
	}
}
