package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/resources"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollectorAzureRmResources struct {
	CollectorProcessorGeneral

	prometheus struct {
		resource *prometheus.GaugeVec
	}
}

func (m *MetricsCollectorAzureRmResources) Setup(collector *CollectorGeneral) {
	m.CollectorReference = collector

	m.prometheus.resource = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "azurerm_resource_info",
			Help: "Azure Resource info",
		},
		append(
			[]string{
				"resourceID",
				"subscriptionID",
				"resourceGroup",
				"provider",
			},
			prefixSliceForPrometheusLabels(AZURE_RESOURCE_TAG_PREFIX, opts.AzureResourceTags)...
		),
	)

	prometheus.MustRegister(m.prometheus.resource)
}

func (m *MetricsCollectorAzureRmResources) Reset() {
	m.prometheus.resource.Reset()
}

func (m *MetricsCollectorAzureRmResources) Collect(ctx context.Context, callback chan<- func(), subscription subscriptions.Subscription) {
	client := resources.NewClient(*subscription.SubscriptionID)
	client.Authorizer = AzureAuthorizer

	list, err := client.ListComplete(ctx, "", "", nil)

	if err != nil {
		panic(err)
	}

	resourceMetric := MetricCollectorList{}

	for list.NotDone() {
		val := list.Value()

		infoLabels := prometheus.Labels{
			"subscriptionID": *subscription.SubscriptionID,
			"resourceID": *val.ID,
			"resourceGroup": extractResourceGroupFromAzureId(*val.ID),
			"provider": extractProviderFromAzureId(*val.ID),
		}
		infoLabels = addAzureResourceTags(infoLabels, val.Tags)
		resourceMetric.AddInfo(infoLabels)

		if list.NextWithContext(ctx) != nil {
			break
		}
	}

	callback <- func() {
		resourceMetric.GaugeSet(m.prometheus.resource)
	}
}
