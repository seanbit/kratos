package webkit

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	// Meter 实例
	meter metric.Meter

	// 指标定义
	_metricSeconds        metric.Float64Histogram
	_metricRequests       metric.Int64Counter
	_metricIntercept      metric.Int64Counter
	_metricBotInterceptor metric.Int64Counter
	_metricTurnstile      metric.Int64Counter
	_alarmStatsMetric     metric.Int64Counter
	_platformMetric       metric.Int64Counter
	_methodDurationMetric metric.Float64Histogram
)

// InitMetrics 初始化指标
func InitMetrics(name string) error {
	// 创建 Prometheus exporter
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	// 创建 MeterProvider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)

	// 获取 Meter
	meter = provider.Meter(name)

	// 初始化各个指标
	if err := initMetrics(); err != nil {
		return err
	}

	return nil
}

func initMetrics() error {
	var err error

	// 1. 请求时长直方图
	_metricSeconds, err = meter.Float64Histogram(
		"server_requests_duration",
		metric.WithDescription("server requests duration(sec)."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.250, 0.5, 1, 3, 5),
	)
	if err != nil {
		return err
	}

	// 2. 请求计数器
	_metricRequests, err = meter.Int64Counter(
		"client_requests_code_total",
		metric.WithDescription("The total number of processed requests"),
	)
	if err != nil {
		return err
	}

	// 3. 拦截策略计数器
	_metricIntercept, err = meter.Int64Counter(
		"server_requests_intercept",
		metric.WithDescription("The interception strategy of requests"),
	)
	if err != nil {
		return err
	}

	// 4. 机器人拦截计数器
	_metricBotInterceptor, err = meter.Int64Counter(
		"server_requests_bot_intercept",
		metric.WithDescription("The bot interception strategy of requests"),
	)
	if err != nil {
		return err
	}

	// 5. Turnstile 策略计数器
	_metricTurnstile, err = meter.Int64Counter(
		"server_requests_turnstile",
		metric.WithDescription("The turnstile strategy of requests"),
	)
	if err != nil {
		return err
	}

	// 6. 告警统计计数器
	_alarmStatsMetric, err = meter.Int64Counter(
		"alarm_stats_service_function_total",
		metric.WithDescription("The number of alarm stats"),
	)
	if err != nil {
		return err
	}

	// 7. 平台指标计数器
	_platformMetric, err = meter.Int64Counter(
		"alarm_platform_api_total",
		metric.WithDescription("The number of alarm platform api alert"),
	)
	if err != nil {
		return err
	}

	// 8. 方法时长直方图
	_methodDurationMetric, err = meter.Float64Histogram(
		"alarm_platform_method_cost",
		metric.WithDescription("Method execution cost"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return err
	}

	return nil
}

// 指标记录方法

func RecordMetricIntercept(serverName, path, hasSign, verify, block string) {
	RecordMetricInterceptWithCtx(nil, serverName, path, hasSign, verify, block)
}
func RecordMetricInterceptWithCtx(ctx context.Context, serverName, path, hasSign, verify, block string) {
	if ctx == nil {
		ctx = context.Background()
	}
	_metricIntercept.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("server_name", serverName),
			attribute.String("path", path),
			attribute.String("has_sign", hasSign),
			attribute.String("verify", verify),
			attribute.String("block", block),
		),
	)
}

func RecordMetricBotInterceptor(operation, block, blockInterceptor, success, headerExist, verifySuccess, interceptIfVerifyFail string) {
	RecordMetricBotInterceptorWithCtx(nil, operation, block, blockInterceptor, success, headerExist, verifySuccess, interceptIfVerifyFail)
}
func RecordMetricBotInterceptorWithCtx(ctx context.Context, operation, block, blockInterceptor, success, headerExist, verifySuccess, interceptIfVerifyFail string) {
	if ctx == nil {
		ctx = context.Background()
	}
	_metricBotInterceptor.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("block", block),
			attribute.String("block_interceptor", blockInterceptor),
			attribute.String("success", success),
			attribute.String("header_exist", headerExist),
			attribute.String("verify_success", verifySuccess),
			attribute.String("intercept_if_verify_fail", interceptIfVerifyFail),
		),
	)
}

func RecordMetricTurnstile(operation, success, headerExist, verifySuccess, interceptIfWithoutHeader, interceptIfVerifyFail string) {
	RecordMetricTurnstileWithCtx(nil, operation, success, headerExist, verifySuccess, interceptIfWithoutHeader, interceptIfVerifyFail)
}
func RecordMetricTurnstileWithCtx(ctx context.Context, operation, success, headerExist, verifySuccess, interceptIfWithoutHeader, interceptIfVerifyFail string) {
	if ctx == nil {
		ctx = context.Background()
	}
	_metricTurnstile.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("success", success),
			attribute.String("header_exist", headerExist),
			attribute.String("verify_success", verifySuccess),
			attribute.String("intercept_if_without_header", interceptIfWithoutHeader),
			attribute.String("intercept_if_verify_fail", interceptIfVerifyFail),
		),
	)
}

func RecordAlarmStatsMetric(serviceName, function string) {
	RecordAlarmStatsMetricWithCtx(nil, serviceName, function)
}
func RecordAlarmStatsMetricWithCtx(ctx context.Context, serviceName, function string) {
	if ctx == nil {
		ctx = context.Background()
	}
	_alarmStatsMetric.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("service_name", serviceName),
			attribute.String("function", function),
		),
	)
}

func RecordPlatformMetric(level, platform, title, msg string) {
	RecordPlatformMetricWithCtx(nil, level, platform, title, msg)
}
func RecordPlatformMetricWithCtx(ctx context.Context, level, platform, title, msg string) {
	if ctx == nil {
		ctx = context.Background()
	}
	_platformMetric.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("level", level),
			attribute.String("platform", platform),
			attribute.String("title", title),
			attribute.String("msg", msg),
		),
	)
}

func RecordMethodDurationMetric(serviceName, operation, proc string, duration float64) {
	RecordMethodDurationMetricWithCtx(nil, serviceName, operation, proc, duration)
}
func RecordMethodDurationMetricWithCtx(ctx context.Context, serviceName, operation, proc string, duration float64) {
	if ctx == nil {
		ctx = context.Background()
	}
	_methodDurationMetric.Record(
		ctx,
		duration,
		metric.WithAttributes(
			attribute.String("service_name", serviceName),
			attribute.String("operation", operation),
			attribute.String("proc", proc),
		),
	)
}
