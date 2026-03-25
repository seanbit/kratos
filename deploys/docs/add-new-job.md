# 添加新 Job 指南

本文档介绍如何添加一个新的一次性 Job 任务。

## 概述

添加新 Job 需要完成三个部分：
1. **Go 代码**: 实现 Job 业务逻辑
2. **K8s 配置**: 创建 Job 部署清单
3. **GitHub Actions**: 添加 Job 选项

## 完整示例

以添加一个名为 `data-migration` 的 Job 为例。

---

## 第一步：Go 代码实现

### 1.1 创建 Job 文件

创建文件 `cmd/job/jobs/data_migration.go`:

```go
package jobs

import (
    "fmt"

    "github.com/go-kratos/kratos/v2/log"
    "github.com/spf13/cobra"
)

// DataMigration job 的参数
var (
    dataMigrationBatchSize int
    dataMigrationDryRun    bool
)

// RegisterDataMigrationFlags 注册 data-migration 命令的参数
func RegisterDataMigrationFlags(cmd *cobra.Command) {
    cmd.Flags().IntVar(&dataMigrationBatchSize, "batch-size", 1000, "每批处理的数据量")
    cmd.Flags().BoolVar(&dataMigrationDryRun, "dry-run", false, "试运行模式，不实际执行")
}

// RunDataMigration 执行 data-migration job
func RunDataMigration(cmd *cobra.Command, app *App) error {
    log.Info("Starting data-migration job...")
    log.Infof("Batch size: %d, Dry run: %v", dataMigrationBatchSize, dataMigrationDryRun)

    // TODO: 实现数据迁移逻辑
    // 可通过 app 访问依赖注入的组件:
    // - app.ScanProgressService
    // - 其他注入的服务...

    if dataMigrationDryRun {
        log.Info("Dry run mode, no changes made")
        return nil
    }

    // 实际迁移逻辑
    // ...

    log.Info("Data migration completed successfully")
    return nil
}
```

### 1.2 在 main.go 中注册命令

编辑 `cmd/job/main.go`:

```go
package main

import (
    // ... 现有 import
)

var (
    // ... 现有变量

    // 添加新命令变量
    dataMigrationCmd = &cobra.Command{
        Use:   "data-migration",
        Short: "Migrate data between tables or databases",
        Long:  "Perform data migration with configurable batch size and dry-run mode",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 初始化配置
            cleanConfig := global.InitConfig("file", configFile, secretFile)
            defer cleanConfig()
            cfg := global.GetConfig()

            // 初始化logger
            webkit.InitLogger(rootCmd.Use, versionCmd.Version, int(cfg.LogLevel))

            // 初始化依赖注入
            app, cleanup, err := initApp(cfg.Server, cfg.Data, cfg.Blockchain, cfg.Scanner, log.DefaultLogger)
            if err != nil {
                return fmt.Errorf("failed to init app: %w", err)
            }
            defer cleanup()

            return jobs.RunDataMigration(cmd, app)
        },
    }
)

func init() {
    // ... 现有代码

    // 注册 data-migration 命令的参数
    jobs.RegisterDataMigrationFlags(dataMigrationCmd)

    // 添加子命令
    rootCmd.AddCommand(dataMigrationCmd)
}
```

### 1.3 验证 Go 代码

```bash
# 编译检查
go build ./cmd/job/...

# 本地测试
go run ./cmd/job/... data-migration --help
go run ./cmd/job/... data-migration --batch-size=100 --dry-run --config=configs/config.yaml
```

---

## 第二步：创建 K8s Job 配置

### 2.1 复制模板

```bash
cp deploys/jobs/job-template.yaml deploys/jobs/data-migration.yaml
```

### 2.2 修改配置

编辑 `deploys/jobs/data-migration.yaml`:

```yaml
# Job: data-migration
# 功能: 数据迁移任务
apiVersion: batch/v1
kind: Job
metadata:
  name: evm-scan-job-data-migration
  namespace: evm-scan
  labels:
    app: evm-scan-job
    job-name: "data-migration"
    app.kubernetes.io/name: job
    app.kubernetes.io/component: job
spec:
  ttlSecondsAfterFinished: 86400
  backoffLimit: 3
  template:
    metadata:
      labels:
        app: evm-scan-job
        job-name: "data-migration"
    spec:
      restartPolicy: Never
      containers:
        - name: job
          image: evm-scan-job
          imagePullPolicy: Always
          command:
            - "./server"
          args:
            - "-conf"
            - "/data/conf"
            - "-run-type"
            - "job"
            - "-job-name"
            - "data-migration"
            # 可添加 Job 特定参数
            - "--batch-size"
            - "1000"
          env:
            - name: ENV_RUN_TYPE
              value: "job"
            - name: ENV_JOB_NAME
              value: "data-migration"
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
          volumeMounts:
            - name: config
              mountPath: /data/conf/config.yaml
              subPath: config.yaml
              readOnly: true
            - name: secret
              mountPath: /data/conf/secret.yaml
              subPath: secret.yaml
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: evm-scan-config
        - name: secret
          secret:
            secretName: evm-scan-secret
      imagePullSecrets:
        - name: harbor-secret
```

### 2.3 自定义资源配额（可选）

根据 Job 的需求调整资源：

```yaml
resources:
  requests:
    cpu: "200m"      # 增加 CPU
    memory: "256Mi"  # 增加内存
  limits:
    cpu: "1000m"
    memory: "1Gi"
```

---

## 第三步：更新 GitHub Actions

### 3.1 添加到 Run Job 工作流

编辑 `.github/workflows/run-job.yaml`:

```yaml
on:
  workflow_dispatch:
    inputs:
      job_name:
        description: 'Job 名称'
        required: true
        type: choice
        options:
          - event-re-dispatch
          - data-migration        # 添加新选项
          # 新增 job 时在此添加选项
```

---

## 第四步：测试和部署

### 4.1 本地测试

```bash
# 1. 编译
make build

# 2. 本地运行测试
./bin/server job data-migration --dry-run --config=configs/config.yaml
```

### 4.2 提交代码

```bash
# 1. 添加所有更改
git add cmd/job/jobs/data_migration.go
git add cmd/job/main.go
git add deploys/jobs/data-migration.yaml
git add .github/workflows/run-job.yaml

# 2. 提交
git commit -m "feat(job): add data-migration job

- Add data migration job with batch processing
- Support dry-run mode for testing
- Add K8s job configuration"

# 3. 推送
git push origin dev
```

### 4.3 构建镜像

1. 进入 GitHub Actions
2. 运行 **Build and Deploy**:
   - app_type: `job`
   - environment: `dev`
   - job_name: `data-migration`

### 4.4 运行 Job

1. 进入 GitHub Actions
2. 运行 **Run K8s Job**:
   - job_name: `data-migration`
   - environment: `dev`

---

## Job 参数传递

### 方法1：通过 args 传递

在 Job YAML 中：

```yaml
args:
  - "-conf"
  - "/data/conf"
  - "-run-type"
  - "job"
  - "-job-name"
  - "data-migration"
  - "--batch-size"
  - "1000"
  - "--dry-run"
```

### 方法2：通过环境变量传递

```yaml
env:
  - name: BATCH_SIZE
    value: "1000"
  - name: DRY_RUN
    value: "true"
```

然后在 Go 代码中读取：

```go
batchSize := os.Getenv("BATCH_SIZE")
dryRun := os.Getenv("DRY_RUN") == "true"
```

### 方法3：通过 ConfigMap 传递

创建 Job 专用 ConfigMap：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: data-migration-config
  namespace: evm-scan
data:
  batch_size: "1000"
  dry_run: "false"
```

---

## Job 模板参考

### 基础 Job 模板

参考 `deploys/jobs/job-template.yaml`

### Go 代码模板

参考 `cmd/job/jobs/job_template.go.example`

---

## 最佳实践

### 1. 命名规范

- Job 名称使用小写字母和连字符: `data-migration`, `cache-cleanup`
- Go 文件使用下划线: `data_migration.go`
- Go 函数使用驼峰: `RunDataMigration`

### 2. 日志输出

```go
import "github.com/go-kratos/kratos/v2/log"

log.Info("Starting job...")
log.Infof("Processing batch %d", batchNum)
log.Errorf("Failed to process: %v", err)
```

### 3. 错误处理

```go
func RunMyJob(cmd *cobra.Command, app *App) error {
    // 返回 error 会导致 Job 失败并触发重试
    if err := doSomething(); err != nil {
        return fmt.Errorf("failed to do something: %w", err)
    }
    return nil
}
```

### 4. 幂等性

确保 Job 可以安全重试：

```go
// 使用事务
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 检查是否已处理
if alreadyProcessed(id) {
    log.Infof("Already processed: %s, skipping", id)
    continue
}
```

### 5. 进度记录

对于长时间运行的 Job：

```go
for i, item := range items {
    if err := process(item); err != nil {
        return err
    }
    if i % 100 == 0 {
        log.Infof("Progress: %d/%d (%.1f%%)", i, len(items), float64(i)/float64(len(items))*100)
    }
}
```

---

## 检查清单

添加新 Job 时，确认以下项目：

- [ ] 创建 Go Job 文件 (`cmd/job/jobs/<name>.go`)
- [ ] 实现 `Register<Name>Flags` 函数
- [ ] 实现 `Run<Name>` 函数
- [ ] 在 `main.go` 中注册命令
- [ ] 本地编译测试通过
- [ ] 创建 K8s Job YAML (`deploys/jobs/<name>.yaml`)
- [ ] 更新 GitHub Actions 工作流
- [ ] 提交并推送代码
- [ ] 构建 Job 镜像
- [ ] 在 Dev 环境测试运行
- [ ] 更新文档（如需要）
