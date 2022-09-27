# InfluxDB2 Agent
The InfluxDB2 Agent was build for my [LilyGo Weather Station]() to move complexity out of the station itself. Less processing means less energy consumption with is key for a battery operated device. The agent provides http endpoints that returns cumulated data that was extracted with 1..n flux queries from an InfluxDB2 database.

## Endpoints
The endpoints are fully configurable. Based on the type a different mapping is used.
Currently only 'weather' is supported.

### Weather Type
Returns weather data for various locations.

The **location** is important as the weather data is grouped by it.
This was done to allow loading different data like temperature or humidity for the same location with
different flux queries and still have them in the same data set.
Note that the data is simply added onto the array, there is no time check to make sure everything is in order!

The data is loaded from InfluxDB2 via one or multiple flux queries that need to be configured in the config file.
Example query:
```
  from(bucket: "weather")
  |> range(start: -72h)
  |> filter(fn: (r) => r["_measurement"] == "openweathermap")
  |> filter(fn: (r) => r["_field"] == "main_temp")
  |> aggregateWindow(every: 108m, fn: mean)
  |> map(fn: (r) => ({ _value:r._value, _time:r._time, _field:"temperature", location:"Outside" }))'
```
This query will load the temperature values for the last 3 days and aggregate them into 108m intervals. My LilyGo weather station uses 40 measurments and thus the 108m. Always use full days.

Currently following attributes in the flux response are extracted:
| Name        | Mandatory | Comment        |
|-------------|-----------|----------------|
| location    | y         |                |
| temperature | n         | rounded to x.y |
| humidity    | n         | rounded to int |
| co2         | n         | rounded to int |

Example response:
```
[
    {
        "location": "Living Room",
        "temperature": [
            21.2,
            20.8
        ],
        "humidity": [
            54,
            53
        ],
        "co2": [
            799,
            777
        ]
    }
]
```

## Configuration
The configuration file must be a valid YAML file. Its path can be passed into the application as argument, else **config.yml** is assumed.

Example **config.yml** file;
```
  port: 9098
  influxDB2: http://127.0.0.1:9086
  token: "abcd"
  organization: "home"
  endpoints:
    - name: "/weather"
      type: "weather"
      queries:
        - 'from(bucket: "weather") |> range(start: -72h) |> filter(fn: (r) => r["_measurement"] == "openweathermap") |> filter(fn: (r) => r["_field"] == "main_temp") |> aggregateWindow(every: 108m, fn: mean) |> map(fn: (r) => ({ _value:r._value, _time:r._time, _field:"temperature", location:"Outside" }))'
```

| Name              | Description                           |
|-------------------|---------------------------------------|
| port              | port on which this agent will listen  |
| influxDB2         | address of InfluxDB2 server           |
| token             | auth token to access InfluxDB2 server |
| organization      | organization of InfluxDB2 server      |
| endpoints.name    | url path                              |
| endpoints.type    | type of endpoint, needed for mapping  |
| endpoints.queries | 1..n flux queries to gather data      |

## Docker
The agent was written with the intent of running it in docker. You can also run it directly if this is preferred.

### Build Image
Execute following statement, then either start via docker or docker compose.
```
docker build -t influxdb2-agent .
```

### Docker
```
docker run -d --restart unless-stopped --name=influxdb2-agent -v ./config.yml:/config.yml -p 9098:9098 influxdb2-agent
```

### Docker Compose
```
version: "3.4"
services:
  influxdb2-agent:
    image: influxdb2-agent
    container_name: influxdb2-agent
    restart: unless-stopped
    ports:
      - "9098:9098"
    volumes:
      - ./config.yml:/config.yml
```