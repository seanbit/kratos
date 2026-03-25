# EVM-Scan CI/CD 部署文档

## 概述

本项目采用 **GitHub Actions + Harbor + ArgoCD + Kubernetes** 的 CI/CD 架构，实现自动化构建和部署。

## 文档索引

| 文档 | 说明 |
|-----|------|
| [架构说明](./architecture.md) | CI/CD 整体架构和流程图 |
| [部署前准备](./setup-guide.md) | 首次部署前的环境配置 |
| [使用指南](./usage-guide.md) | 日常构建部署操作说明 |
| [添加新 Job](./add-new-job.md) | 如何添加新的一次性任务 |
| [日志收集](./logging.md) | PLG 日志收集配置指南 |
| [故障排除](./troubleshooting.md) | 常见问题和解决方案 |

## 快速开始

### 1. 完成部署前准备

参考 [部署前准备](./setup-guide.md) 完成以下配置：
- 配置 GitHub Secrets
- 配置 Harbor 认证
- 配置 K8s 集群访问
- 配置 ArgoCD Application

### 2. 触发构建部署

1. 进入 GitHub 仓库 → Actions
2. 选择 **Build and Deploy** 工作流
3. 点击 **Run workflow**
4. 填写参数并执行

### 3. 查看部署状态

- GitHub Actions: 查看构建日志
- ArgoCD: 查看同步状态
- Kubernetes: 查看 Pod 状态

## 应用类型

| 应用 | 说明 | K8s 资源类型 |
|-----|------|-------------|
| web | HTTP/gRPC API 服务 | Deployment + Service |
| scanner | 区块链扫描服务 | Deployment |
| job | 一次性任务 | Job |

## 环境说明

| 环境 | 分支 | 命名空间 | 说明 |
|-----|------|---------|------|
| dev | dev | evm-scan-dev | 开发测试环境 |
| prod | main | evm-scan-prod | 生产环境 |

## 目录结构

```
deploys/
├── Dockerfile              # Docker 构建文件
├── docs/                   # 文档目录
├── base/                   # Kustomize 基础配置
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── harbor-secret.yaml
│   ├── web-deployment.yaml
│   ├── web-service.yaml
│   └── scanner-deployment.yaml
├── overlays/               # 环境特定配置
│   ├── dev/
│   │   ├── kustomization.yaml
│   │   ├── config.yaml
│   │   ├── secret.yaml
│   │   └── patches/
│   └── prod/
│       ├── kustomization.yaml
│       ├── config.yaml
│       └── patches/
└── jobs/                   # Job 模板
    ├── job-template.yaml
    └── event-re-dispatch.yaml
```

## 技术栈

- **CI**: GitHub Actions
- **镜像仓库**: Harbor (harbor.sean.vip)
- **CD**: ArgoCD
- **编排**: Kubernetes + Kustomize
- **日志**: PLG (Promtail + Loki + Grafana)
- **语言**: Go 1.24

## 联系方式

如有问题，请联系项目维护者。
