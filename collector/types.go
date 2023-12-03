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

// Documentation: https://openweathermap.org/api/one-call-3#current

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

type Rain struct {
	OneH float64 `json:"1h,omitempty"`
}

type Snow struct {
	OneH float64 `json:"1h,omitempty"`
}

// Weather https://openweathermap.org/weather-conditions#Weather-Condition-Codes-2
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// OneCallData the API should be called with exclude=minutely,hourly,daily,alerts
type OneCallData struct {
	Latitude       float64            `json:"lat"`
	Longitude      float64            `json:"lon"`
	Timezone       string             `json:"timezone"`
	TimezoneOffset int                `json:"timezone_offset"`
	Current        OneCallCurrentData `json:"current,omitempty"`
}

type OneCallCurrentData struct {
	Dt         int       `json:"dt"`
	Sunrise    int       `json:"sunrise"`
	Sunset     int       `json:"sunset"`
	Temp       float64   `json:"temp"`
	FeelsLike  float64   `json:"feels_like"`
	Pressure   int       `json:"pressure"`
	Humidity   int       `json:"humidity"`
	DewPoint   float64   `json:"dew_point"`
	Clouds     int       `json:"clouds"`
	UVI        float64   `json:"uvi"`
	Visibility int       `json:"visibility"`
	WindSpeed  float64   `json:"wind_speed"`
	WindGust   float64   `json:"wind_gust,omitempty"`
	WindDeg    float64   `json:"wind_deg"`
	Rain       Rain      `json:"rain,omitempty"`
	Snow       Snow      `json:"snow,omitempty"`
	Weather    []Weather `json:"weather"`
}

// Pollution API: https://openweathermap.org/api/air-pollution#current

type Pollution struct {
	Location struct {
		Longitude float64 `json:"lon"`
		Latitude  float64 `json:"lat"`
	} `json:"coord"`
	List []PollutionData `json:"list"`
}

type PollutionData struct {
	Dt   int `json:"dt"`
	Main struct {
		Aqi float64 `json:"aqi"`
	} `json:"main"`
	Components struct {
		Co   float64 `json:"co"`
		No   float64 `json:"no"`
		No2  float64 `json:"no2"`
		O3   float64 `json:"o3"`
		So2  float64 `json:"so2"`
		Pm25 float64 `json:"pm2_5"`
		Pm10 float64 `json:"pm10"`
		Nh3  float64 `json:"nh3"`
	} `json:"components"`
}
