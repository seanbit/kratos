# CI/CD 架构说明

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              开发者工作流                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   1. 代码提交          2. 手动触发           3. 自动部署                      │
│   ┌─────────┐         ┌─────────────┐       ┌─────────────┐                │
│   │  开发者  │ ──────▶ │   GitHub    │ ────▶ │   ArgoCD    │                │
│   │         │  push   │   Actions   │  sync │             │                │
│   └─────────┘         └─────────────┘       └─────────────┘                │
│                              │                     │                        │
│                              ▼                     ▼                        │
│                       ┌─────────────┐       ┌─────────────┐                │
│                       │   Harbor    │       │ Kubernetes  │                │
│                       │  (镜像仓库)  │ ────▶ │   Cluster   │                │
│                       └─────────────┘  pull └─────────────┘                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 详细流程

### CI 流程 (GitHub Actions)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Build and Deploy Workflow                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ Validate │───▶│   Test   │───▶│  Build   │───▶│  Update  │  │
│  │  Inputs  │    │          │    │  Image   │    │ Manifests│  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│       │               │               │               │         │
│       ▼               ▼               ▼               ▼         │
│  检查参数有效性    运行单元测试    构建Docker镜像   更新Kustomize  │
│                  (可选)Lint检查   推送到Harbor     提交到Git      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### CD 流程 (ArgoCD)

```
┌─────────────────────────────────────────────────────────────────┐
│                       ArgoCD Sync Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  Detect  │───▶│   Sync   │───▶│  Apply   │───▶│  Health  │  │
│  │  Change  │    │          │    │ Resources│    │  Check   │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│       │               │               │               │         │
│       ▼               ▼               ▼               ▼         │
│  监控Git仓库变化   比较期望状态    应用K8s资源     检查应用健康    │
│  (kustomization)   与实际状态                     状态           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 组件说明

### 1. GitHub Actions

负责 CI 阶段，包含两个工作流：

#### Build and Deploy (`build-deploy.yaml`)
- **触发方式**: 手动触发 (workflow_dispatch)
- **功能**: 构建镜像、推送到 Harbor、更新 Kustomize 配置
- **参数**:
  - `app_type`: 应用类型 (web/scanner/job)
  - `environment`: 部署环境 (dev/prod)
  - `enable_lint`: 是否启用代码检查
  - `job_name`: Job 名称 (仅 job 类型)

#### Run K8s Job (`run-job.yaml`)
- **触发方式**: 手动触发
- **功能**: 在 K8s 中运行一次性 Job
- **参数**:
  - `job_name`: Job 名称
  - `environment`: 运行环境

### 2. Harbor

私有 Docker 镜像仓库：
- **地址**: harbor.sean.vip
- **项目**: coding/evm-scan
- **镜像标签格式**: `<app_type>-<commit_sha_7位>`
  - 例: `web-a1b2c3d`, `scanner-e4f5g6h`

### 3. Kustomize

Kubernetes 配置管理：

```
deploys/
├── base/                    # 基础配置 (所有环境共享)
│   ├── kustomization.yaml   # 资源清单
│   ├── web-deployment.yaml  # Web 应用部署
│   ├── scanner-deployment.yaml
│   ├── configmap.yaml       # 配置文件
│   └── secret.yaml          # 敏感配置
│
└── overlays/
    ├── dev/                 # Dev 环境覆盖
    │   ├── kustomization.yaml
    │   ├── config.yaml      # Dev 配置 (CI自动更新)
    │   ├── secret.yaml      # Dev Secret (CI自动更新)
    │   └── patches/         # 资源补丁
    │
    └── prod/                # Prod 环境覆盖
        ├── kustomization.yaml
        ├── config.yaml      # Prod 配置 (CI自动更新)
        └── patches/         # 资源补丁
```

### 4. ArgoCD

GitOps 持续部署：
- 监控 Git 仓库的 `deploys/overlays/<env>` 目录
- 自动检测 Kustomize 配置变化
- 自动同步到 Kubernetes 集群
- 提供 Web UI 查看部署状态

### 5. Kubernetes

运行环境：

| 命名空间 | 说明 |
|---------|------|
| evm-scan-dev | 开发环境 |
| evm-scan-prod | 生产环境 |

## 镜像标签策略

```
镜像地址格式: harbor.sean.vip/coding/evm-scan:<tag>

标签格式: <app_type>-<git_commit_sha_7位>

示例:
- web-a1b2c3d       # Web 应用
- scanner-e4f5g6h   # Scanner 应用
- job-i7j8k9l       # Job 应用
```

## 配置更新策略

### ConfigMap (config.yaml)

| 环境 | 更新方式 | 说明 |
|-----|---------|------|
| dev | 自动更新 | CI 自动从 `configs/config.yaml` 复制 |
| prod | 自动更新 | CI 自动从 `configs/config.yaml` 复制 |

### Secret (secret.yaml)

| 环境 | 更新方式 | 说明 |
|-----|---------|------|
| dev | 自动更新 | CI 自动从 `configs/secret.yaml` 复制 |
| prod | 手动配置 | 需在 K8s 集群中手动创建/更新 |

## 资源配额

### Dev 环境

| 应用 | CPU 请求 | CPU 限制 | 内存请求 | 内存限制 | 副本数 |
|-----|---------|---------|---------|---------|-------|
| web | 50m | 200m | 64Mi | 256Mi | 1 |
| scanner | 100m | 500m | 128Mi | 512Mi | 1 |
| job | 100m | 500m | 128Mi | 512Mi | 1 |

### Prod 环境

| 应用 | CPU 请求 | CPU 限制 | 内存请求 | 内存限制 | 副本数 |
|-----|---------|---------|---------|---------|-------|
| web | 100m | 500m | 128Mi | 512Mi | 2 |
| scanner | 200m | 1000m | 256Mi | 1Gi | 1 |
| job | 100m | 500m | 128Mi | 512Mi | 1 |

## 网络架构

```
                    ┌─────────────────┐
                    │    Ingress      │
                    │  (可选,外部访问) │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Web Service   │
                    │  :8000 (HTTP)   │
                    │  :9000 (gRPC)   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────┐  ┌──────▼──────┐  ┌───▼────┐
     │  Web Pod    │  │ Scanner Pod │  │Job Pod │
     │             │  │             │  │(临时)   │
     └─────────────┘  └─────────────┘  └────────┘
              │              │              │
              └──────────────┼──────────────┘
                             │
                    ┌────────▼────────┐
                    │   ConfigMap     │
                    │   Secret        │
                    └─────────────────┘
```

## 安全考虑

1. **Harbor 认证**: 使用 Robot Account，最小权限原则
2. **K8s Secret**: Prod 环境 Secret 不自动更新，需手动管理
3. **GitHub Secrets**: 敏感信息存储在 GitHub Secrets
4. **镜像拉取**: 使用 imagePullSecrets 认证
5. **配置文件**: 敏感配置不提交到 Git (通过 .gitignore)
