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

package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/billykwooten/openweather-exporter/collector"
	"github.com/jellydator/ttlcache/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Default App Flags
	app         = kingpin.New("openweather-exporter", "Openweather Exporter for Openweather API").Author("Billy Wooten")
	addr        = app.Flag("listen-address", "HTTP port to listen on. (Default 9091)").Envar("OW_LISTEN_ADDRESS").Default(":9091").String()
	apiKey      = app.Flag("apikey", "Openweather API Key").Envar("OW_APIKEY").Required().String()
	city        = app.Flag("city", "City for Openweather to gather metrics from.").Envar("OW_CITY").Default("New York, NY").String()
	degreesUnit = app.Flag("degrees-unit", "The base unit for temperature output. Fahrenheit or Celsius. (Default: F)").Envar("OW_DEGREES_UNIT").Default("F").String()
	language    = app.Flag("language", "The language for metric output. (Default: EN)").Envar("OW_LANGUAGE").Default("EN").String()
	cacheTTL    = app.Flag("cache-ttl", "Cache time-to-live in seconds. (Default: 300)").Envar("OW_CACHE_TTL").Default("300").String()

	// Extra App Flags
	enablePol = app.Flag("enable-pol", "Enable Pollution Metrics. (Default: false)").Envar("OW_ENABLE_POL").Default("false").Bool()
	enableUV  = app.Flag("enable-uv", "Enable Ultraviolet Index Metrics. (Default: false)").Envar("OW_ENABLE_UV").Default("false").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Setup better logging
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}

	log.SetFormatter(formatter)

	// Create a new instance of the weatherCollector with caching and
	// register it with the prometheus client.
	cache := ttlcache.NewCache()
	ttl, err := strconv.ParseUint(*cacheTTL, 10, 64)
	if err != nil {
		log.Fatal("Invalid TTL value: ", err)
	}
	cache.SetTTL(time.Duration(ttl) * time.Second)
	cache.SkipTTLExtensionOnHit(true)

	// Add some logging for extra collectors
	if *enablePol {
		log.Info("Pollution metrics enabled, this will call the API more than once per call.")
	}
	if *enableUV {
		log.Info("Ultraviolet Index metrics enabled, this will call the API more than once per call.")
	}

	settings := collector.Settings{
		DegreesUnit: *degreesUnit, Language: *language, ApiKey: *apiKey,
	}
	weatherCollector := collector.NewOpenweatherCollector(&settings, *city, cache)
	prometheus.MustRegister(weatherCollector)

	// This section will start the HTTP server and expose
	// any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port " + *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
