package main

import (
	"flag"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

var (
	gitCommit = "dev"
	buildDate = "dev"

	bindPort = flag.Int("port", 4242, "HTTP port to listen")

	kafkaBrokers          = flag.String("kafka-brokers", "localhost", "The comma separated list of the Kafka brokers")
	kafkaJMXAdminPassword = flag.String("kafka-jmx-password", "", "JMX password")

	metricsInterval = flag.Int("metrics-interval", 30, "Interval (in seconds) for collecting metrics")
	openTSDBURL     = flag.String("opentsdb-url", "", "OpenTSDB URL to store metrics, e.g. http://url.to.opentsdb.com")
	openTSDBToken   = flag.String("opentsdb-token", "", "OpenTSDB Token to write metrics, e.g. user:password")
)

func main() {
	flag.Parse()

	metricsCtr := MetricsCtr{JMXAdminPassword: *kafkaJMXAdminPassword, Brokers: *kafkaBrokers}
	tick := time.Second * time.Duration(*metricsInterval)

	openTSDBChan := StartOpenTSDBReporter(*openTSDBURL, *openTSDBToken)
	go metricsCtr.StartMetricCollector(tick, openTSDBChan)

	serveAPI("kafka-metrics-collector", router(metricsCtr))
}

func router(mCtr MetricsCtr) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	//router.Use(cORSMiddleware())

	// Serve static
	router.Use(static.Serve("/", static.LocalFile("views", true)))

	// Base routes
	router.GET("/version", version)
	router.GET("/api", index)
	router.GET("/favicon.ico", favicon)

	// Authentication
	router.GET("/list", mCtr.ListMetrics)
	router.GET("/metrics/:metrics", mCtr.GetMetrics)
	router.GET("/metrics/:metrics/total", mCtr.GetTotalMetrics)

	return router
}
