package protobufmessage

import (
	"fmt"

	event "github.com/prometheus/statsd_exporter/pkg/event"
)

func buildEvent(statType, metric string, value float64, relative bool, labels map[string]string) (event.Event, error) {
	switch statType {
	case "c":
		return &event.CounterEvent{
			CMetricName: metric,
			CValue:      float64(value),
			CLabels:     labels,
		}, nil
	case "g":
		return &event.GaugeEvent{
			GMetricName: metric,
			GValue:      float64(value),
			GRelative:   relative,
			GLabels:     labels,
		}, nil
	case "ms", "h", "d":
		return &event.TimerEvent{
			TMetricName: metric,
			TValue:      float64(value),
			TLabels:     labels,
		}, nil
	case "s":
		return nil, fmt.Errorf("no support for StatsD sets")
	default:
		return nil, fmt.Errorf("bad stat type %s", statType)
	}
}

func MessageToEvent(metric TraceMetric) event.Events {
	events := event.Events{}
	labels := map[string]string{}

	for _, tag := range metric.Tags {
		labels[*tag.Name] = *tag.Value
	}

	event, _ := buildEvent(*metric.Type, *metric.Name, *metric.Value, false, labels)

	events = append(events, event)

	return events
}
