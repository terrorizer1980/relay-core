package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/label"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// createCustomMetric creates a custom metric specified by the metric type.
func CreateCustomMetric(ctx context.Context, projectID string, mc *monitoring.MetricClient, w io.Writer, name, metricType, description string) (*metricpb.MetricDescriptor, error) {
	md := &metric.MetricDescriptor{
		Name:        name,
		DisplayName: name,
		Type:        metricType,
		Labels: []*label.LabelDescriptor{{
			Key:         "environment",
			ValueType:   label.LabelDescriptor_STRING,
			Description: "The environment of the reported metric",
		}},
		MetricKind:  metric.MetricDescriptor_GAUGE,
		ValueType:   metric.MetricDescriptor_INT64,
		Unit:        "1",
		Description: description,
	}
	req := &monitoringpb.CreateMetricDescriptorRequest{
		Name:             "projects/" + projectID,
		MetricDescriptor: md,
	}
	m, err := mc.CreateMetricDescriptor(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not create custom metric: %v", err)
	}

	_, _ = fmt.Fprintf(w, "Created %s custom metric\n", m.GetName())
	return m, nil
}

// makeTimeSeriesValue constructs a a value for the custom metric created
func MakeTimeSeriesValue(projectID, metricType, environment string, value int) *monitoringpb.TimeSeries {
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	return &monitoringpb.TimeSeries{
		Metric: &metricpb.Metric{
			Type: metricType,
			Labels: map[string]string{
				"environment": environment,
			},
		},
		Resource: &monitoredres.MonitoredResource{
			Type: "global",
			Labels: map[string]string{
				"project_id": projectID,
			},
		},
		// Only a single point seems to be possible.
		Points: []*monitoringpb.Point{{
			Interval: &monitoringpb.TimeInterval{
				EndTime: now,
			},
			Value: &monitoringpb.TypedValue{
				Value: &monitoringpb.TypedValue_Int64Value{
					Int64Value: int64(value),
				},
			},
		}},
	}
}

// writeTimeSeriesRequest publishes the timeseries datapoint. The series must all be the same metric type
func WriteTimeSeriesRequest(ctx context.Context, projectID string, mc *monitoring.MetricClient, series []*monitoringpb.TimeSeries) error {
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       "projects/" + projectID,
		TimeSeries: series,
	}
	log.Printf("writeTimeseriesRequest: %+v\n", req)

	err := mc.CreateTimeSeries(ctx, req)
	if err != nil {
		return fmt.Errorf("could not write time series value, %v ", err)
	}
	return nil
}

// deleteMetric deletes the given metric. name should be of the form
// "projects/PROJECT_ID/metricDescriptors/METRIC_TYPE".
func DeleteMetric(w io.Writer, name string) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	req := &monitoringpb.DeleteMetricDescriptorRequest{
		Name: name,
	}

	if err := c.DeleteMetricDescriptor(ctx, req); err != nil {
		return fmt.Errorf("could not delete metric: %v", err)
	}
	fmt.Fprintf(w, "Deleted metric: %q\n", name)
	return nil
}
