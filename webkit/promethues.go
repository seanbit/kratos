package webkit

//
//import "github.com/prometheus/client_golang/prometheus"
//
//func init() {
//	prometheus.MustRegister(
//		MetricSeconds,
//		MetricRequests,
//		MetricIntercept,
//		MetricBotInterceptor,
//		MetricTurnstile,
//	)
//}
//
//var (
//	MetricSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
//		Namespace: "server",
//		Subsystem: "requests",
//		Name:      "duration_sec",
//		Help:      "server requests duration(sec).",
//		Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1, 3, 5},
//	}, []string{"kind", "operation"})
//
//	MetricRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "client",
//		Subsystem: "requests",
//		Name:      "code_total",
//		Help:      "The total number of processed requests",
//	}, []string{"kind", "operation", "code", "reason"})
//
//	MetricIntercept = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "server",
//		Subsystem: "requests",
//		Name:      "intercept",
//		Help:      "The interception strategy of requests",
//	}, []string{"server_name", "path", "has_sign", "verify", "block"})
//	MetricBotInterceptor = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "server",
//		Subsystem: "requests",
//		Name:      "bot_intercept",
//		Help:      "The bot interception strategy of requests",
//	}, []string{"operation", "block", "block_interceptor", "success", "header_exist", "verify_success", "intercept_if_verify_fail"})
//	MetricTurnstile = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "server",
//		Subsystem: "requests",
//		Name:      "turnstile",
//		Help:      "The turnstile strategy of requests",
//	}, []string{"operation", "success", "header_exist", "verify_success", "intercept_if_without_header", "intercept_if_verify_fail"})
//
//	AlarmStatsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "alarm",
//		Subsystem: "stats",
//		Name:      "service_function_total",
//		Help:      "The number of alarm stats",
//	}, []string{"service_name", "function"})
//	PlatformMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Namespace: "alarm",
//		Subsystem: "platform",
//		Name:      "api_total",
//		Help:      "The number of alarm platform api alert",
//	}, []string{"level", "platform", "title", "msg"})
//	MethodDurationMetric = prometheus.NewHistogramVec(prometheus.HistogramOpts{
//		Namespace: "alarm",
//		Subsystem: "platform",
//		Name:      "method_cost",
//		Buckets:   prometheus.DefBuckets,
//		//Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1, 3, 5},
//	}, []string{"service_name", "operation", "proc"})
//)
