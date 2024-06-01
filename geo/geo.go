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

package geo

import (
	log "github.com/sirupsen/logrus"
)

var n = Nominatim{}

func GetCoords(city string) (float64, float64, error) {
	log.Info("Looking up: " + city)

	results, err := n.Search(SearchParameters{ // Check SearchResult struct for details
		Query:          city,
		IncludeAddress: true,
		IncludeGeoJSON: true,
	})
	if err != nil {
		return 0, 0, err
	}

	if results[0].Lat != 0 {
		log.Infof("Latitude: %f Longitude: %f for %s found", results[0].Lat, results[0].Lng, results[0].DisplayName)
		return results[0].Lat, results[0].Lng, nil
	} else {
		log.Fatalf("Could not get location data for %s", city)
		return 0, 0, err
	}

}
