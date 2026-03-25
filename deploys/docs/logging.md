# 日志收集配置指南

本文档介绍如何配置 PLG (Promtail + Loki + Grafana) 日志收集系统与 EVM-Scan 应用集成。

## 概述

```
┌─────────────────────────────────────────────────────────────────┐
│                        日志收集架构                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │   Pod    │───▶│ Promtail │───▶│   Loki   │───▶│ Grafana  │  │
│  │ (stdout) │    │ (收集)    │    │ (存储)    │    │ (查询)    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 应用配置

### Pod 标签

所有应用 Pod 已配置以下标签，用于日志过滤：

```yaml
labels:
  app: evm-scan-<component>           # 应用标识
  app.kubernetes.io/name: <component> # 组件名称 (web/scanner/job)
  app.kubernetes.io/component: <type> # 组件类型 (api/worker/job)
  app.kubernetes.io/part-of: evm-scan # 项目标识
  logging: enabled                     # 日志收集开关
```

### Pod 注解

```yaml
annotations:
  promtail.io/collect: "true"   # 启用日志收集
  promtail.io/parser: "json"    # 日志格式 (json/logfmt/raw)
```

---

## PLG 基础设施配置

> 以下配置用于独立部署的 PLG 基础设施项目

### 1. Promtail 配置

Promtail 需要配置 scrape_configs 来收集 EVM-Scan 应用的日志：

```yaml
# promtail-config.yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Kubernetes Pod 日志收集
  - job_name: kubernetes-pods
    kubernetes_sd_configs:
      - role: pod

    relabel_configs:
      # 只收集带有 logging: enabled 标签的 Pod
      - source_labels: [__meta_kubernetes_pod_label_logging]
        action: keep
        regex: enabled

      # 检查 promtail.io/collect 注解
      - source_labels: [__meta_kubernetes_pod_annotation_promtail_io_collect]
        action: keep
        regex: true

      # 设置命名空间标签
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace

      # 设置 Pod 名称标签
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod

      # 设置应用标签
      - source_labels: [__meta_kubernetes_pod_label_app]
        target_label: app

      # 设置组件标签
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        target_label: component

      # 设置组件类型标签
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_component]
        target_label: component_type

      # 设置项目标签
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_part_of]
        target_label: project

      # Job 名称标签 (仅 Job 类型)
      - source_labels: [__meta_kubernetes_pod_label_job_name]
        target_label: job_name

      # 容器名称
      - source_labels: [__meta_kubernetes_pod_container_name]
        target_label: container

    pipeline_stages:
      # JSON 日志解析
      - match:
          selector: '{project="evm-scan"}'
          stages:
            - json:
                expressions:
                  level: level
                  msg: msg
                  ts: ts
                  caller: caller
            - labels:
                level:
            - timestamp:
                source: ts
                format: RFC3339Nano
```

### 2. Loki 配置

```yaml
# loki-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://alertmanager:9093

limits_config:
  retention_period: 720h  # 30 天
  ingestion_rate_mb: 10
  ingestion_burst_size_mb: 20
  max_streams_per_user: 10000
  max_entries_limit_per_query: 5000
```

### 3. Grafana 数据源配置

```yaml
# grafana-datasource.yaml
apiVersion: 1
datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
    jsonData:
      maxLines: 1000
```

---

## Grafana 日志查询

### 常用 LogQL 查询

```logql
# 查询所有 evm-scan 日志
{project="evm-scan"}

# 按环境过滤
{project="evm-scan", namespace="evm-scan-dev"}
{project="evm-scan", namespace="evm-scan-prod"}

# 按组件过滤
{project="evm-scan", component="web"}
{project="evm-scan", component="scanner"}
{project="evm-scan", component="job"}

# 按日志级别过滤
{project="evm-scan"} | json | level="error"
{project="evm-scan"} | json | level=~"error|warn"

# 搜索特定内容
{project="evm-scan"} |= "database connection"
{project="evm-scan"} |~ "error.*timeout"

# 查询特定 Job
{project="evm-scan", job_name="event-re-dispatch"}

# 统计错误数量
sum(rate({project="evm-scan"} | json | level="error" [5m])) by (component)
```

### 推荐 Dashboard 面板

1. **日志流面板**: 实时日志流
2. **错误统计面板**: 按组件统计错误数量
3. **日志级别分布**: 各级别日志占比
4. **Top 错误**: 最常见的错误信息

---

## 应用日志格式

### 推荐的 JSON 日志格式

应用应输出以下格式的 JSON 日志：

```json
{
  "ts": "2024-01-09T10:30:00.123456789Z",
  "level": "info",
  "msg": "Request processed successfully",
  "caller": "handler/user.go:42",
  "request_id": "abc123",
  "method": "GET",
  "path": "/api/v1/users",
  "duration": "15.2ms",
  "status": 200
}
```

### Go 应用日志配置

使用 zerolog (已在项目中使用):

```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func init() {
    // JSON 格式输出
    zerolog.TimeFieldFormat = time.RFC3339Nano

    // 设置日志级别
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// 使用示例
log.Info().
    Str("request_id", requestID).
    Str("method", r.Method).
    Str("path", r.URL.Path).
    Dur("duration", duration).
    Int("status", status).
    Msg("Request processed")
```

---

## 告警配置

### Loki 告警规则

```yaml
# loki-alerts.yaml
groups:
  - name: evm-scan-alerts
    rules:
      # 错误率告警
      - alert: HighErrorRate
        expr: |
          sum(rate({project="evm-scan"} | json | level="error" [5m])) by (component) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in {{ $labels.component }}"
          description: "Component {{ $labels.component }} has error rate > 0.1/s for 5 minutes"

      # 无日志告警
      - alert: NoLogs
        expr: |
          absent(sum(rate({project="evm-scan", component="web"} [5m])))
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "No logs from web component"
          description: "No logs received from web component for 10 minutes"

      # 致命错误告警
      - alert: FatalError
        expr: |
          count_over_time({project="evm-scan"} | json | level="fatal" [1m]) > 0
        labels:
          severity: critical
        annotations:
          summary: "Fatal error in {{ $labels.component }}"
          description: "Fatal error detected in {{ $labels.component }}"
```

---

## 标签说明

| 标签 | 说明 | 示例值 |
|-----|------|-------|
| `namespace` | K8s 命名空间 | `evm-scan-dev`, `evm-scan-prod` |
| `pod` | Pod 名称 | `dev-evm-scan-web-xxx` |
| `container` | 容器名称 | `web`, `scanner`, `job` |
| `app` | 应用标识 | `evm-scan-web`, `evm-scan-scanner` |
| `component` | 组件名称 | `web`, `scanner`, `job` |
| `component_type` | 组件类型 | `api`, `worker`, `job` |
| `project` | 项目标识 | `evm-scan` |
| `job_name` | Job 名称 (仅 Job) | `event-re-dispatch` |
| `level` | 日志级别 (解析后) | `debug`, `info`, `warn`, `error`, `fatal` |

---

## 故障排除

### 日志未收集

1. 检查 Pod 标签:
   ```bash
   kubectl get pod <pod-name> -n evm-scan-dev -o yaml | grep -A10 "labels:"
   ```

2. 检查 Promtail 日志:
   ```bash
   kubectl logs -f deployment/promtail -n logging
   ```

3. 确认 Pod 注解:
   ```bash
   kubectl get pod <pod-name> -n evm-scan-dev -o yaml | grep -A5 "annotations:"
   ```

### 日志解析失败

1. 检查应用日志格式是否为 JSON
2. 验证 JSON 格式:
   ```bash
   kubectl logs <pod-name> -n evm-scan-dev | head -1 | jq .
   ```

### Loki 查询慢

1. 添加更多标签过滤条件
2. 缩小时间范围
3. 检查 Loki 资源配置

---

## PLG 部署参考

推荐使用 Helm 部署 PLG 栈：

```bash
# 添加 Grafana Helm 仓库
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# 部署 Loki Stack (包含 Loki + Promtail + Grafana)
helm install loki grafana/loki-stack \
  --namespace logging \
  --create-namespace \
  --set grafana.enabled=true \
  --set promtail.enabled=true \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=50Gi
```

或使用官方 Kustomize 配置:
- Loki: https://github.com/grafana/loki/tree/main/production/kustomize
- Promtail: https://github.com/grafana/loki/tree/main/clients/cmd/promtail
