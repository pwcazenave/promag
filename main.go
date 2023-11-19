package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
    "strconv"
	"strings"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AirGradient struct {
/*
	// Content of payload from the AirGradient firmware
	String payload = "{\"wifi\":" + String(WiFi.RSSI())
	+ (Co2 < 0 ? "" : ", \"rco2\":" + String(Co2))
	+ (pm25 < 0 ? "" : ", \"pm02\":" + String(pm25))
	+ ", \"atmp\":" + String(temp)
	+ (hum < 0 ? "" : ", \"rhum\":" + String(hum))
	+ "}";
	// Yields JSON something like:
	{
		"wifi": -51,
		"rco2": 517,
		"pm02": 4,
		"atmp": 15.5,
		"rhum": 78
	}

*/
	Wifi int     `json:"wifi"` // wifi signal strength (dB)
	Rco2 int     `json:"rco2"` // CO2 (ppm)
	Pm02 int     `json:"pm02"` // 2.0um particulate matter (ug/m^3)
	Atmp float64 `json:"atmp"` // atmospheric temperature (Celsius or Farenheit, user configured)
	Rhum int     `json:"rhum"` // relative humidity (%)
}

type prometheusMetrics struct {
	probeDuration prometheus.Gauge
	probeSuccess  prometheus.Gauge
	WiFi          prometheus.Gauge
	Rco2          prometheus.Gauge
	Pm02          prometheus.Gauge
	Atmp          prometheus.Gauge
	Rhum          prometheus.Gauge
}

func getEnv(key, fallback string) string {
    value, exists := os.LookupEnv(key)
    if !exists {
        value = fallback
    }
    return value
}

func initCollectors(reg *prometheus.Registry) *prometheusMetrics {
	m := new(prometheusMetrics)

	m.probeDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "probe_duration",
			Help: "How many seconds the probe took",
		},
	)
	reg.MustRegister(m.probeDuration)

	m.probeSuccess = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "probe_success",
			Help: "Whether or not the probe succeeded",
		},
	)
	reg.MustRegister(m.probeSuccess)

	m.WiFi = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "airgradient_wifi_strength",
			Help: "WiFi Signal Strength (dB)",
		},
	)
	reg.MustRegister(m.WiFi)

	m.Rco2 = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "airgradient_rco2",
			Help: "Relative CO2 concentration (ppm)",
		},
	)
	reg.MustRegister(m.Rco2)

	m.Pm02 = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "airgradient_pm02",
			Help: "2.0 ug particulate matter concentration (ug/m^3)",
		},
	)
	reg.MustRegister(m.Pm02)

	m.Atmp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "airgradient_atmp",
			Help: "Atmospheric Temperaeture (Celsius or Farenheit, user configured)",
		},
	)
	reg.MustRegister(m.Atmp)

	m.Rhum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "airgradient_rhum",
			Help: "Relative humidity (%)",
		},
	)
	reg.MustRegister(m.Rhum)

	return m

}

func parseAirGradientJSON(w http.ResponseWriter, r *http.Request, client *redis.Client, ctx context.Context) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// Use the ID as the prefix for the in-memory keys
	deviceParts := strings.Split(parts[len(parts) - 2], ":")
	deviceID := deviceParts[len(deviceParts) - 1]
	log.Println("Device ID:", deviceID)
	log.Println("URL Path:", path)
	log.Println("URL Parts:", parts)
	decoder := json.NewDecoder(r.Body)

	var jsonBody AirGradient
    err := decoder.Decode(&jsonBody)
    if err != nil {
        panic(err)
    }
    log.Println("wifi strength:", jsonBody.Wifi)
    log.Println("rco2:", jsonBody.Rco2)
    log.Println("pm02:", jsonBody.Pm02)
    log.Println("atmp:", jsonBody.Atmp)
    log.Println("rhum:", jsonBody.Rhum)
	log.Println("")

	log.Println(fmt.Sprintf("Setting Wifi for %s to %d", deviceID, jsonBody.Wifi))
	err = client.Set(ctx, fmt.Sprintf("%s_wifi", deviceID), jsonBody.Wifi, 0).Err()
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Setting Rco2 for %s to %d", deviceID, jsonBody.Rco2))
	err = client.Set(ctx, fmt.Sprintf("%s_rco2", deviceID), jsonBody.Rco2, 0).Err()
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Setting Pm02 for %s to %d", deviceID, jsonBody.Pm02))
	err = client.Set(ctx, fmt.Sprintf("%s_pm02", deviceID), jsonBody.Pm02, 0).Err()
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Setting atmp for %s to %f", deviceID, jsonBody.Atmp))
	err = client.Set(ctx, fmt.Sprintf("%s_atmp", deviceID), jsonBody.Atmp, 0).Err()
	if err != nil {
		panic(err)
	}
	log.Println(fmt.Sprintf("Setting rhum for %s to %d", deviceID, jsonBody.Rhum))
	err = client.Set(ctx, fmt.Sprintf("%s_rhum", deviceID), jsonBody.Rhum, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (m *prometheusMetrics) probeHandler(w http.ResponseWriter, r *http.Request, reg *prometheus.Registry, client *redis.Client, ctx context.Context) {
	// Fetch the values from the redis for the sensor name (given as the target)
	var success float64 = 1
	start := time.Now()

	params := r.URL.Query()
	deviceID := params.Get("target")
	if deviceID == "" {
		http.Error(w, "Target parameter missing or empty", http.StatusBadRequest)
		return
	}

	log.Println(fmt.Sprintf("Getting Wifi for %s", deviceID))
	Wifi, err := client.Get(ctx, fmt.Sprintf("%s_wifi", deviceID)).Result()
	if err != nil {
		success = 0
		log.Fatal(err)
		// panic(err)
	}
	log.Println(fmt.Sprintf("Got Wifi for %s of %s", deviceID, Wifi))

	log.Println(fmt.Sprintf("Getting Rco2 for %s", deviceID))
	Rco2, err := client.Get(ctx, fmt.Sprintf("%s_rco2", deviceID)).Result()
	if err != nil {
		success = 0
		log.Fatal(err)
		// panic(err)
	}
	log.Println(fmt.Sprintf("Got Rco2 for %s of %s", deviceID, Rco2))

	log.Println(fmt.Sprintf("Getting Pm02 for %s", deviceID))
	Pm02, err := client.Get(ctx, fmt.Sprintf("%s_pm02", deviceID)).Result()
	if err != nil {
		success = 0
		log.Fatal(err)
		// panic(err)
	}
	log.Println(fmt.Sprintf("Got Pm02 for %s of %s", deviceID, Pm02))

	log.Println(fmt.Sprintf("Getting Atmp for %s", deviceID))
	Atmp, err := client.Get(ctx, fmt.Sprintf("%s_atmp", deviceID)).Result()
	if err != nil {
		success = 0
		log.Fatal(err)
		// panic(err)
	}
	log.Println(fmt.Sprintf("Got Atmp for %s of %s", deviceID, Atmp))

	log.Println(fmt.Sprintf("Getting Rhum for %s", deviceID))
	Rhum, err := client.Get(ctx, fmt.Sprintf("%s_rhum", deviceID)).Result()
	if err != nil {
		success = 0
		log.Fatal(err)
		// panic(err)
	}
	log.Println(fmt.Sprintf("Got Rhum for %s of %s", deviceID, Rhum))

	if success == 1 {
		iWifi, err := strconv.ParseFloat(strings.TrimSpace(Rco2), 64)
		if err != nil {
			log.Fatal(err)
			// panic(err)
		}
		m.WiFi.Set(iWifi)
		iRco2, err := strconv.ParseFloat(strings.TrimSpace(Wifi), 64)
		if err != nil {
			log.Fatal(err)
			// panic(err)
		}
		m.Rco2.Set(iRco2)
		iPm02, err := strconv.ParseFloat(strings.TrimSpace(Pm02), 64)
		if err != nil {
			log.Fatal(err)
			// panic(err)
		}
		m.Pm02.Set(iPm02)
		iAtmp, err := strconv.ParseFloat(strings.TrimSpace(Atmp), 64)
		if err != nil {
			log.Fatal(err)
			// panic(err)
		}
		m.Atmp.Set(iAtmp)
		iRhum, err := strconv.ParseFloat(strings.TrimSpace(Rhum), 64)
		if err != nil {
			log.Fatal(err)
			// panic(err)
		}
		m.Rhum.Set(iRhum)
	}

	duration := time.Since(start).Seconds()
	m.probeSuccess.Set(success)
	m.probeDuration.Set(duration)
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
	h.ServeHTTP(w, r)

}

func main() {
	// Read config from environment variables
	promHost := getEnv("PROM_HOST", "")
	promPort := getEnv("PROM_PORT", "9000")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisDB := getEnv("REDIS_DB", "0")
	iredisPort, err := strconv.Atoi(redisPort)
	if err != nil {
		log.Fatal(err)
		// panic(err)
	}
	iredisDB, err := strconv.Atoi(redisDB)
	if err != nil {
		log.Fatal(err)
		// panic(err)
	}

	redisConnection := fmt.Sprintf("%s:%d", redisHost, iredisPort)
	log.Println("Connecting to redis on", redisConnection)

	// Connect to redis to store the latest incoming AirGradient data
    client := redis.NewClient(&redis.Options{
        Addr:	  redisConnection,
        Password: redisPassword,
        DB:		  iredisDB,
    })
	ctx := context.Background()

	// Set up the prometheus end of things
	registry := prometheus.NewRegistry()
	metrics := initCollectors(registry)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		metrics.probeHandler(w, r, registry, client, ctx)
	})
	// Listen on /sensors/ for the incoming AirGradient data. Use the sensor ID as the prefix for keys in redis.
	http.HandleFunc("/sensors/", func(w http.ResponseWriter, r *http.Request) {
		parseAirGradientJSON(w, r, client, ctx)
	})
	promConnection := fmt.Sprintf("%s:%s", promHost, promPort)
	log.Println("Listening on", promConnection)
	http.ListenAndServe(promConnection, nil)
}
