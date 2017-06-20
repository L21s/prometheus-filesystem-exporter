package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fsnotify/fsnotify"
)

var (
	metricsDir  = flag.String("metrics-directory", "/metrics", "The directory to read metrics from")
	metricsPath = flag.String("metrics-path", "/metrics", "The http path under which metrics are exposed")
	listenAddr  = flag.String("listen-addr", ":8080", "The address to listen on for http requests")
	gauges      = make(map[string]prometheus.Gauge)
)

func main() {
	flag.Parse()

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Create:
					getOrCreateGaugeForPath(event.Name)
					updateGauge(event.Name)
				case fsnotify.Write:
					updateGauge(event.Name)
				case fsnotify.Remove:
					removeGauge(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(*metricsDir)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func getOrCreateGaugeForPath(path string) prometheus.Gauge {
	if pathIsDir(path) {
		return nil
	}
	gauge := gauges[path]
	if gauge != nil {
		return gauge
	}
	log.Println("Attempting to generate gauge from file: " + path)
	gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: filenameFromPath(path),
			Help: "Auto generated from filesystem path: " + path,
		})
	prometheus.MustRegister(gauge)
	gauges[path] = gauge
	return gauge
}

func updateGauge(path string) {
	if pathIsDir(path) {
		return
	}
	log.Println("Attempting to update gauge with value written to: " + path)
	gauge := getOrCreateGaugeForPath(path)
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(string(dat)), 64)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Setting gauge for " + path + " to " + strconv.FormatFloat(value, 'f', 10, 64))
	gauge.Set(value)
}

func removeGauge(path string) {
	log.Println("Attempting to remove gauge because of deleted file: " + path)
	gauge := gauges[path]
	if gauge == nil {
		return
	}
	prometheus.Unregister(gauge)
	delete(gauges, path)
}

func pathIsDir(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		return true
	}
	switch mode := file.Mode(); {
	case mode.IsDir():
		log.Println(path + " is a directory")
		return true
	default:
		return false
	}
}

func filenameFromPath(path string) string {
	var pathComponents = strings.Split(path, "/")
	return pathComponents[len(pathComponents)-1]
}
