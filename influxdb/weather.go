package influxdb

import (
	"context"
	"encoding/json"
	"log"
	"math"

	"github.com/dtrumpfheller/influxdb2-agent/helpers"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Location struct {
	Location    string    `json:"location"`
	Temperature []float64 `json:"temperature,omitempty"`
	Humidity    []float64 `json:"humidity,omitempty"`
	Co2         []float64 `json:"co2,omitempty"`
}

func GetWeather(config helpers.Config) ([]byte, error) {

	locations := make(map[string]Location)

	// create client objects
	client := influxdb2.NewClient(config.InfluxDB2, config.Token)
	queryAPI := client.QueryAPI(config.Organization)

	for _, query := range config.Weather.Queries {
		// execute flux query
		result, err := queryAPI.Query(context.Background(), query)
		if err != nil {
			log.Printf("Error calling InfluxDB [%s]!\n", err.Error())
			return nil, err
		}

		// process response
		for result.Next() {
			var name = result.Record().ValueByKey("location").(string)

			var location Location
			if val, ok := locations[name]; ok {
				location = val
			} else {
				location.Location = name
				location.Temperature = make([]float64, 0)
				location.Humidity = make([]float64, 0)
				location.Co2 = make([]float64, 0)
			}

			if result.Record().Field() == "temperature" {
				value := getFloat(result.Record().Value(), 1, location.Temperature)
				location.Temperature = append(location.Temperature, value)

			} else if result.Record().Field() == "humidity" {
				value := getFloat(result.Record().Value(), 0, location.Humidity)
				location.Humidity = append(location.Humidity, value)

			} else if result.Record().Field() == "co2" {
				value := getFloat(result.Record().Value(), 0, location.Co2)
				location.Co2 = append(location.Co2, value)
			}

			locations[name] = location
		}
	}

	// ensures background processes finishes
	client.Close()

	// convert map to slice of values for marhsalling
	values := []Location{}
	for _, value := range locations {
		values = append(values, value)
	}

	// marshall into json
	json, err := json.Marshal(values)
	if err != nil {
		log.Printf("Error processing InfluxDB response [%s]!\n", err.Error())
		return nil, err
	}

	return json, nil
}

func getFloat(value interface{}, precision int, array []float64) float64 {

	if value == nil {
		// value is null, use last value from the array or zero if empty
		if len(array) > 0 {
			return array[len(array)-1]
		} else {
			return 0
		}
	}

	float := value.(float64)

	// round value to desired precision
	if precision == 0 {
		float = math.Round(float)
	} else if precision == 1 {
		float = float64(int(float*10)) / 10
	} else if precision == 2 {
		float = float64(int(float*100)) / 100
	} else {
		log.Panicf("Unsupported precision [%d]!\n", precision)
	}

	return float
}
