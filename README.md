# OpenWeather Exporter
Prometheus exporter for [openweather API](https://openweathermap.org/api)

# Requirements

* Linux / MacOSX, [`git bash`](https://git-scm.com/download/win) for Windows
* [docker](https://www.docker.com)

## Development

* [go](https://golang.org/dl)
* [openweathermap](https://github.com/briandowns/openweathermap)

## Configuration

Openweather exporter can be controlled by both ENV or CLI flags as described below.

| Environment        	       | CLI (`--flag`)              | Default                 	    | Description                                                                                                      |
|----------------------------|-----------------------------|---------------------------- |------------------------------------------------------------------------------------------------------------------|
| `OW_LISTEN_ADDRESS`           | `listen-address`            | `:9091`                     | The port for /metrics to listen on |
| `OW_APIKEY`                   | `apikey`                    | `<REQUIRED>`                | Your Openweather API key |
| `OW_CITY`                     | `city`                      | `New York, NY`              | City/Location in which to gather weather metrics |
| `OW_DEGREES_UNIT`             | `degrees-unit`              | `F`                         | Unit in which to show metrics (Kelvin, Fahrenheit or Celsius) |
| `OW_LANGUAGE`                 | `language`                  | `EN`                        | Language in which to show metrics |

## Usage

```
# Export weather metrics from Seattle using binary
./openweather-exporter --city "Seattle, WA" --apikey mi4o2n54i0510n4510

# Export weather metrics from Seattle using docker
docker run -d --restart on-failure --name=openweather-exporter -p 9091:9091 billykwooten/openweather-exporter --city "Seattle, WA" --apikey mi4o2n54i0510n4510
```

## Development

If you'd like to build this yourself you can clone this repo and run:

```
./script/cibuild
```