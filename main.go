package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	metrics     = make(map[string]*prometheus.GaugeVec)
)

func main() {
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*metricsDir, "*"))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		updateMetric(file)
	}

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
					updateMetric(event.Name)
				case fsnotify.Write:
					updateMetric(event.Name)
				case fsnotify.Remove:
					removeMetric(event.Name)
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

	http.Handle(*metricsPath, promhttp.Handler())
	log.Println("Watching " + *metricsDir + " for metrics")
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func getOrCreateMetricForPath(path string) *prometheus.GaugeVec {
	metricName := metricNameFromPath(path)
	metric, ok := metrics[metricName]
	if ok {
		return metric
	}
	log.Println("Attempting to generate metric from file: " + path)
	labels := labelsFromPath(path)
	labelKeys := make([]string, 0, len(labels))
	for k := range labels {
		labelKeys = append(labelKeys, k)
	}
	metric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricName,
			Help: "Auto generated from filesystem path: " + *metricsDir + "/" + metricName,
		},
		labelKeys)
	prometheus.MustRegister(metric)
	metrics[metricName] = metric
	return metric
}

func updateMetric(path string) {
	if pathIsDir(path) {
		return
	}
	log.Println("Attempting to update metric with value written to: " + path)
	metric := getOrCreateMetricForPath(path)
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
	log.Println("Setting metric for " + path + " to " + strconv.FormatFloat(value, 'f', 10, 64))
	metric.With(labelsFromPath(path)).Set(value)
}

func removeMetric(path string) {
	log.Println("Attempting to remove metric because of deleted file: " + path)
	metric := metrics[path]
	if metric == nil {
		return
	}
	prometheus.Unregister(metric)
	delete(metrics, path)
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

func metricNameFromPath(path string) string {
	return strings.Split(filenameFromPath(path), ";")[0]
}

func labelsFromPath(path string) map[string]string {
	labels := make(map[string]string)
	labelsFromFilename := strings.Split(filenameFromPath(path), ";")
	//remove the first element - it's the metrics name
	labelsFromFilename = labelsFromFilename[1:]
	for _, labelPair := range labelsFromFilename {
		splittedLabelPair := strings.Split(labelPair, "=")
		if len(splittedLabelPair) != 2 {
			log.Println(labelPair + " in file" + path + " is invalid - please make sure all filenames have the following format: metricName;label=value;label=value")
			continue
		}
		labels[splittedLabelPair[0]] = splittedLabelPair[1]
	}
	return labels
}
