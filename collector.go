package main

import (
	"time"

	"github.com/Sirupsen/logrus"
)

// StartMetricCollector starts the metrics collector that fetchs Kafka metrics
// at each tick and send it to the OpenTSDB reporter
func (ctr *MetricsCtr) StartMetricCollector(interval time.Duration, openTSDB chan []Point) {
	intervalTicker := time.Tick(interval)

	ctr.collectAndSend(openTSDB)

	for {
		select {
		case <-intervalTicker:
			ctr.collectAndSend(openTSDB)
		}
	}
}

func (ctr *MetricsCtr) collectAndSend(openTSDB chan []Point) {
	metrics := ctr.collect()
	logrus.Infof("Ask to send %d metrics", len(metrics))
	points := metrics2Points(metrics)
	openTSDB <- points
}

func (ctr *MetricsCtr) collect() []map[string]map[string]float64 {
	allMetrics := make([]map[string]map[string]float64, len(metricsList))

	for i, metric := range metricsList {
		metrics, errors := ctr.getAllMetrics(metric)
		if len(metrics) == 0 && len(errors) > 1 {
			logrus.WithError(errors[0]).Error("Fail to get all metrics")
			continue
		}

		totalMetrics, err := ctr.calcTotal(metrics)
		if err != nil {
			logrus.WithError(err).Error("Fail to calculate metrics total")
			continue
		}

		allMetrics[i] = totalMetrics
	}

	return allMetrics
}

func metrics2Points(metrics []map[string]map[string]float64) []Point {
	var points []Point
	tags := map[string]string{}
	now := time.Now().Unix()

	for _, metric := range metrics {
		for name, kv := range metric {
			for key, value := range kv {
				points = append(points, Point{
					Timestamp: now,
					Tags:      tags,
					Metric:    name + "." + key,
					Value:     value,
				})
			}
		}
	}

	return points
}
