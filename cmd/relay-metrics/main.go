package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/puppetlabs/relay-core/pkg/metrics/opt"
	"github.com/puppetlabs/relay-core/pkg/metrics/service"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	cfg := opt.NewConfig()

	ctx := context.Background()
	var err error

	kc, err := cfg.KubernetesClient()
	if err != nil {
		panic(err.Error())
	}
	mc, err := cfg.MetricsClient(ctx)
	if err != nil {
		panic(err.Error())
	}

	statusName := "Pending Workflow Runs"
	oldestName := "Oldest Pending Workflow Run"
	statusType := "custom.googleapis.com/relay/workflowruns/status"
	oldestType := "custom.googleapis.com/relay/workflowruns/oldest_pending"

	if cfg.DeleteMetrics == true {
		err = service.DeleteMetric(os.Stdout, "projects/"+cfg.ProjectID+"/metricDescriptors/"+statusType)
		if err != nil {
			panic(err.Error())
		}
		err = service.DeleteMetric(os.Stdout, "projects/"+cfg.ProjectID+"/metricDescriptors/"+oldestType)
		if err != nil {
			panic(err.Error())
		}
		return
	}

	if cfg.PublishMetrics == true {
		_, err = service.CreateCustomMetric(ctx, cfg.ProjectID, mc, os.Stdout, statusName, statusType, "The number of Workflow Runs with a status of `pending`")
		if err != nil {
			panic(err.Error())
		}

		_, err = service.CreateCustomMetric(ctx, cfg.ProjectID, mc, os.Stdout, oldestName, oldestType, "The number of seconds since the oldest Workflow Runs with a status of `pending` was started")
		if err != nil {
			panic(err.Error())
		}
	}

	for {
		statuses := service.GetStatuses(ctx, kc)
		count := 0
		oldest := 0.0
		for _, status := range statuses {
			if status.Status == "pending" || status.Status == "" {
				count += 1
				if status.SecondsSinceStart > oldest {
					oldest = status.SecondsSinceStart
					fmt.Printf("Found older: %d\n", int(oldest))
				}
			}
		}
		statusSeries := service.MakeTimeSeriesValue(cfg.ProjectID, statusType, cfg.Environment, count)
		oldestSeries := service.MakeTimeSeriesValue(cfg.ProjectID, oldestType, cfg.Environment, int(oldest))

		if cfg.PublishMetrics == true {
			err = service.WriteTimeSeriesRequest(ctx, cfg.ProjectID, mc, []*monitoringpb.TimeSeries{statusSeries, oldestSeries})
			if err != nil {
				panic(err.Error())
			}
		} else {
			fmt.Println("not reporting")
		}

		// Stackdriver only wants points published every ten seconds or less
		time.Sleep(cfg.Interval * time.Second)
	}
}
