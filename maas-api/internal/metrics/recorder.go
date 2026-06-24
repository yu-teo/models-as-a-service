package metrics

import "time"

type MetricsRecorder interface {
	RecordRequestDuration(method, route, statusCode, tenant string, duration time.Duration)
	IncrementInFlight(method string)
	DecrementInFlight(method string)
}
