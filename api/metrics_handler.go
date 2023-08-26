package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
)

type Metrics struct {
	SMTPDelivered float64
	SMTPReceived  float64
	Users         int64
	Messages      int64
}

func (api *API) metricsJSONHandler(c echo.Context) error {
	// Gather all metrics
	promMetrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Errorf("Unable to gather metrics: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to gather metrics"})
	}

	// TODO: this should be done better
	metrics := Metrics{}

	smtpDelivered, err := getMetricByName(promMetrics, "smtp_delivered")
	if err == nil {
		if len(smtpDelivered.Metric) > 0 {
			metrics.SMTPDelivered = *smtpDelivered.Metric[0].Counter.Value
		}
	}

	smtpReceived, err := getMetricByName(promMetrics, "smtp_received")
	if err == nil {
		if len(smtpReceived.Metric) > 0 {
			metrics.SMTPReceived = *smtpReceived.Metric[0].Counter.Value
		}
	}

	usersCount, err := api.backend.UserRepo.GetTotalUsersCount()
	if err != nil {
		log.Errorf("Unable to gather metrics: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to gather user count metric"})
	}
	metrics.Users = usersCount

	messageCount, err := api.backend.MessageRepo.GetTotalMessagesCount()
	if err != nil {
		log.Errorf("Unable to gather metrics: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to gather message count metric"})
	}
	metrics.Messages = messageCount

	return c.JSON(http.StatusOK, metrics)
}

// getMetricByName searches for a metric by its name and returns it.
func getMetricByName(metrics []*io_prometheus_client.MetricFamily, nameToFind string) (*io_prometheus_client.MetricFamily, error) {
	for _, metric := range metrics {
		if *metric.Name == nameToFind {
			return metric, nil
		}
	}
	return nil, fmt.Errorf("metric not found: %s", nameToFind)
}
