package service

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	nebulav1 "github.com/puppetlabs/relay-core/pkg/apis/nebula.puppet.com/v1"
)

type WorkflowRunMetric struct {
	Name              string
	Status            string
	SecondsSinceStart float64
}

func GetStatuses(ctx context.Context, c client.Client) (ret []*WorkflowRunMetric) {
	now := time.Now().UTC()
	wrs := &nebulav1.WorkflowRunList{}
	err := c.List(ctx, wrs)
	if err != nil {
		panic(err.Error())
	}
	for _, item := range wrs.Items {
		m := WorkflowRunMetric{
			Name:              item.Name,
			Status:            item.Status.Status,
			SecondsSinceStart: now.Sub(item.ObjectMeta.CreationTimestamp.UTC()).Seconds(),
		}
		fmt.Printf("%s is %s, %d\n", m.Name, m.Status, int(m.SecondsSinceStart))
		ret = append(ret, &m)
	}
	return
}
