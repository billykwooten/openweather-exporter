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

package main

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/billykwooten/openweather-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app         = kingpin.New("openweather-exporter", "Openweather Exporter for Openweather API").Author("Billy Wooten")
	addr        = app.Flag("listen-address", "HTTP port to listen on").Envar("OW_LISTEN_ADDRESS").Default(":9091").String()
	apiKey      = app.Flag("apikey", "Openweather API Key").Envar("OW_APIKEY").Required().String()
	city        = app.Flag("city", "City for Openweather to gather metrics from.").Envar("OW_CITY").Default("New York, NY").String()
	degreesUnit = app.Flag("degrees-unit", "The base unit for temperature output. Fahrenheit or Celsius").Envar("OW_DEGREES_UNIT").Default("F").String()
	language    = app.Flag("language", "The language for metric output").Envar("OW_LANGUAGE").Default("EN").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Setup better logging
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}

	log.SetFormatter(formatter)

	// Create a new instance of the weatherCollector and
	// register it with the prometheus client.
	weatherCollector := collector.NewOpenweatherCollector(*degreesUnit, *language, *apiKey, *city)
	prometheus.MustRegister(weatherCollector)

	// This section will start the HTTP server and expose
	// any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port " + *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
