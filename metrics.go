package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// MetricsCtr is the metrics API controller that exposes Kafka JMX metrics
type MetricsCtr struct {
	JMXAdminPassword string
	Brokers          string
}

const (
	kafkaJMXPort = 7777
	jmxAdminUser = "zuperadmin"
)

// Metrics gathers all metrics from one broker
type Metrics map[string]map[string]interface{}

// GetMetrics returns all brokers metrics
func (ctr *MetricsCtr) GetMetrics(c *gin.Context) {
	metricsPath := c.Param("metrics")

	metrics, errors := ctr.getAllMetrics(metricsPath)
	if len(metrics) == 0 && len(errors) > 1 {
		c.JSON(500, gin.H{"error": errors[0].Error()})
		return
	}

	c.JSON(200, metrics)
}

// GetTotalMetrics returns the total for all brokers metrics
func (ctr *MetricsCtr) GetTotalMetrics(c *gin.Context) {
	metricsPath := c.Param("metrics")

	metrics, errors := ctr.getAllMetrics(metricsPath)
	if len(metrics) == 0 && len(errors) > 1 {
		c.JSON(500, gin.H{"error": errors[0].Error()})
		return
	}

	totalMetrics, err := ctr.calcTotal(metrics)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, totalMetrics)
}

func (ctr *MetricsCtr) getAllMetrics(metricsPath string) (map[string]Metrics, []error) {
	metrics := map[string]Metrics{}
	errors := []error{}
	var m sync.RWMutex

	brokers := strings.Split(ctr.Brokers, ",")
	nbBrokers := len(brokers)

	var wg sync.WaitGroup
	wg.Add(nbBrokers)

	// Get metrics for each broker
	for _, b := range brokers {
		go func(broker string) {
			defer wg.Done()

			url := fmt.Sprintf("http://%s:%s@%s:%d/jmx/read/%s", jmxAdminUser, ctr.JMXAdminPassword, broker, kafkaJMXPort, metricsPath)
			metric, err := jmxMetrics(url)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"broker":      broker,
					"metricsPath": metricsPath,
				}).Error("Fail to get JMX metrics")

				errors = append(errors, err)
				return
			}
			m.Lock()
			defer m.Unlock()
			metrics[broker] = metric

		}(b)
	}
	wg.Wait()

	return metrics, errors
}

var metricsList = []string{
	"kafka.server:name=BytesInPerSec,type=BrokerTopicMetrics",
	"kafka.server:name=BytesInPerSec,topic=*,type=BrokerTopicMetrics",
	"kafka.server:name=BytesOutPerSec,type=BrokerTopicMetrics",
	"kafka.server:name=BytesOutPerSec,topic=*,type=BrokerTopicMetrics",
	"kafka.server:name=MessagesInPerSec,type=BrokerTopicMetrics",
	"kafka.server:name=MessagesInPerSec,topic=*,type=BrokerTopicMetrics",
	"kafka.controller:name=ActiveControllerCount,type=KafkaController",
	"kafka.network:name=RequestQueueSize,type=RequestChannel",
	"kafka.network:name=RequestQueueTimeMs,request=*,type=RequestMetrics",
	"kafka.server:name=UnderReplicatedPartitions,type=ReplicaManager",
	"kafka.cluster:name=UnderReplicated,partition=*,topic=*,type=Partition",
	"kafka.log:name=LogFlushRateAndTimeMs,type=LogFlushStats",
	"kafka.network:name=TotalTimeMs,request=Produce,type=RequestMetrics",
	"kafka.network:name=TotalTimeMs,request=FetchConsumer,type=RequestMetrics",
	"kafka.network:name=TotalTimeMs,request=FetchFollower,type=RequestMetrics",
}

// ListMetrics lists collected metrics
func (ctr *MetricsCtr) ListMetrics(c *gin.Context) {
	c.JSON(200, metricsList)
}

func jmxMetrics(url string) (Metrics, error) {
	metrics := Metrics{}

	var mbeans map[string]interface{}
	err := Get(url, &mbeans)
	if err != nil {
		logrus.Error(err)
		return nil, fmt.Errorf("Internal server error")
	}

	// Handle error
	if mbeans["error"] != nil {
		return nil, fmt.Errorf("%v", mbeans["error"])
	}

	values := mbeans["value"].(map[string]interface{})
	mbeanName := mbeans["request"].(map[string]interface{})["mbean"].(string)

	// Detect if the level is a metric
	if strings.Contains(mbeanName, "*") {
		for name := range values {
			metrics[formatMetricName(name)] = values[name].(map[string]interface{})
		}
	} else {
		// Or If it is directly the values and pick the name in the request/mbean field
		name := mbeans["request"].(map[string]interface{})["mbean"].(string)
		metrics[formatMetricName(name)] = values
	}

	return metrics, nil
}

func formatMetricName(mbeanName string) string {
	//mbeanName = strings.Replace(mbeanName, ":name=", "_", -1)
	//mbeanName = strings.Replace(mbeanName, ",", "_", -1)
	return mbeanName
}

func (ctr *MetricsCtr) calcTotal(metrics map[string]Metrics) (map[string]map[string]float64, error) {
	totalMeters := map[string]map[string]float64{}

	var firstBroker string
	for k := range metrics {
		firstBroker = k
		break
	}
	var firstMetric string
	for k := range metrics[firstBroker] {
		firstMetric = k
		break
	}

	metricNames := extractMetricNames(metrics)
	metricMeters := extractMeterNames(metrics[firstBroker][firstMetric])

	// For each metric
	for _, metricName := range metricNames {
		meter := map[string]float64{}
		// For each broker metric
		for broker, metric := range metrics {

			logrus.WithFields(logrus.Fields{
				"metric": metricName,
				"broker": broker,
				"values": metric[metricName],
			}).Debug("metricNames")

			// A metric is not necessarily on all brokers
			if metric[metricName] == nil {
				continue
			}

			for _, k := range metricMeters {
				value := metric[metricName][k]
				switch t := value.(type) {
				case string:
					// Ignore
				case int32, int64, float32, float64:
					meter[k] += metric[metricName][k].(float64)
				default:
					logrus.WithField("metric", metric[metricName]).Errorf("Unknow type '%s'", t)
					return nil, errors.New("Internal server error")
				}

			}
		}
		totalMeters[metricName] = meter
	}

	return totalMeters, nil
}

func extractMetricNames(metrics map[string]Metrics) []string {
	metricNamesMap := map[string]string{}
	metricNames := []string{}
	for _, brokerMetrics := range metrics {
		for metricName := range brokerMetrics {
			metricNamesMap[metricName] = metricName
		}
	}
	for metricName := range metricNamesMap {
		metricNames = append(metricNames, metricName)
	}
	return metricNames
}

func extractMeterNames(metrics map[string]interface{}) []string {
	meterNames := []string{}
	for meterName := range metrics {
		meterNames = append(meterNames, meterName)
	}
	return meterNames
}

type jmxMeter struct {
	RateUnit          string  `json:"RateUnit"`
	OneMinuteRate     float64 `json:"OneMinuteRate"`
	EventType         string  `json:"EventType"`
	Count             float64 `json:"Count"`
	FifteenMinuteRate float64 `json:"FifteenMinuteRate"`
	FiveMinuteRate    float64 `json:"FiveMinuteRate"`
	MeanRate          float64 `json:"MeanRate"`
}

func toJmxMeter(meter interface{}) jmxMeter {
	if meter == nil {
		return jmxMeter{}
	}
	// Cast meter in map of (string, object)
	meterMap := meter.(map[string]interface{})

	return jmxMeter{
		Count:             meterMap["Count"].(float64),
		EventType:         meterMap["EventType"].(string),
		FifteenMinuteRate: meterMap["FifteenMinuteRate"].(float64),
		FiveMinuteRate:    meterMap["FiveMinuteRate"].(float64),
		MeanRate:          meterMap["MeanRate"].(float64),
		OneMinuteRate:     meterMap["OneMinuteRate"].(float64),
		RateUnit:          meterMap["RateUnit"].(string),
	}
}
