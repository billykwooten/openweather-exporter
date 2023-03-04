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

	"github.com/codingsince1985/geo-golang"
)

func GetCoords(geocoder geo.Geocoder, city string) (float64, float64, error) {
	location, err := geocoder.Geocode(city)
	if err != nil {
		return 0, 0, err
	}

	log.Infof("Latitude: %f Longitude: %f for %s found", location.Lat, location.Lng, city)

	return location.Lat, location.Lng, nil
}
