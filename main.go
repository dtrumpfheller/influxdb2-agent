package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dtrumpfheller/influxdb2-agent/helpers"
	"github.com/dtrumpfheller/influxdb2-agent/influxdb"
)

type WeatherHandler struct {
	config   helpers.Config
	endpoint helpers.Endpoint
}

var (
	configFile = flag.String("config", "config.yml", "configuration file")
)

func main() {

	// load arguments into variables
	flag.Parse()

	// load config file
	config := helpers.ReadConfig(*configFile)

	// configure endpoints
	for _, endpoint := range config.Endpoints {
		if endpoint.Type == "weather" {
			handler := WeatherHandler{config: config, endpoint: endpoint}
			http.HandleFunc(endpoint.Name, handler.weatherHandler)
		} else {
			log.Printf("Endpoint type [%s] not supported!\n", endpoint.Type)
		}
	}

	// start server
	log.Printf("Beginning to serve on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}

func (wh WeatherHandler) weatherHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Getting weather data... ")
	start := time.Now()

	// get weather data and write into response
	json, err := influxdb.GetWeather(wh.config, wh.endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}

	log.Printf("Finished in %s\n", time.Since(start))
}
