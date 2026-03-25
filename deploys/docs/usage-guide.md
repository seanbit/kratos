# 使用指南

本文档介绍日常构建、部署和运维操作。

## 目录

1. [构建和部署应用](#1-构建和部署应用)
2. [运行一次性 Job](#2-运行一次性-job)
3. [查看部署状态](#3-查看部署状态)
4. [回滚操作](#4-回滚操作)
5. [更新配置](#5-更新配置)
6. [日志查看](#6-日志查看)

---

## 1. 构建和部署应用

### 1.1 通过 GitHub Actions 触发

1. 打开 GitHub 仓库页面
2. 点击 **Actions** 标签
3. 左侧选择 **Build and Deploy** 工作流
4. 点击 **Run workflow** 按钮
5. 填写参数:

| 参数 | 说明 | 选项 |
|-----|------|-----|
| 应用类型 | 要部署的应用 | `web` / `scanner` / `job` |
| 部署环境 | 目标环境 | `dev` / `prod` |
| 启用代码检查 | 是否运行 golangci-lint | `true` / `false` |

6. 点击绿色 **Run workflow** 按钮
7. 等待执行完成

### 1.2 执行流程说明

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Validate   │───▶│    Test     │───▶│    Build    │───▶│   Update    │
│   Inputs    │    │             │    │   Image     │    │  Manifests  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
      │                  │                  │                  │
      ▼                  ▼                  ▼                  ▼
 检查参数有效性      运行单元测试       构建推送镜像      更新Kustomize
                   (可选)Lint                           提交到Git
```

### 1.3 查看执行日志

点击正在运行的工作流，可以查看每个步骤的详细日志：

- **Validate inputs**: 参数验证
- **Run unit tests**: 单元测试结果
- **Run golangci-lint**: 代码检查结果（如启用）
- **Build and push image**: 镜像构建和推送日志
- **Update image tag**: Kustomize 更新结果
- **Deployment Summary**: 部署摘要

### 1.4 部署完成确认

部署完成后，ArgoCD 会自动同步。可通过以下方式确认：

```bash
# 查看 Pod 状态
kubectl get pods -n evm-scan-<env>

# 查看 Deployment 状态
kubectl describe deployment <app-name> -n evm-scan-<env>

# 通过 ArgoCD CLI
argocd app get evm-scan-<env>
```

---

## 2. 运行一次性 Job

### 2.1 通过 GitHub Actions 触发

1. 打开 GitHub 仓库 → Actions
2. 选择 **Run K8s Job** 工作流
3. 点击 **Run workflow**
4. 填写参数:
   - **Job 名称**: 选择要运行的 Job
   - **运行环境**: `dev` / `prod`
5. 点击 **Run workflow** 执行

### 2.2 查看 Job 执行日志

工作流会自动等待 Job 完成并输出日志：

- Job 创建日志
- Job 执行日志
- Job 完成状态

### 2.3 手动在 K8s 中运行 Job

如果需要手动运行：

```bash
# 查看可用的 Job 模板
ls deploys/jobs/

# 应用 Job (需要先修改镜像标签)
kubectl apply -f deploys/jobs/<job-name>.yaml

# 查看 Job 状态
kubectl get jobs -n evm-scan-<env>

# 查看 Job 日志
kubectl logs job/<job-name> -n evm-scan-<env>
```

---

## 3. 查看部署状态

### 3.1 通过 kubectl

```bash
# 查看所有资源
kubectl get all -n evm-scan-dev
kubectl get all -n evm-scan-prod

# 查看 Pod 详情
kubectl describe pod <pod-name> -n evm-scan-<env>

# 查看 Deployment 详情
kubectl describe deployment <deployment-name> -n evm-scan-<env>

# 查看事件
kubectl get events -n evm-scan-<env> --sort-by='.lastTimestamp'
```

### 3.2 通过 ArgoCD Web UI

1. 访问 ArgoCD Web UI
2. 点击对应的 Application (evm-scan-dev 或 evm-scan-prod)
3. 查看:
   - **Sync Status**: 同步状态
   - **Health Status**: 健康状态
   - **Resource Tree**: 资源树状图
   - **Events**: 事件日志

### 3.3 通过 ArgoCD CLI

```bash
# 查看应用列表
argocd app list

# 查看应用详情
argocd app get evm-scan-dev
argocd app get evm-scan-prod

# 查看应用资源
argocd app resources evm-scan-dev
```

---

## 4. 回滚操作

### 4.1 通过 ArgoCD 回滚

```bash
# 查看历史版本
argocd app history evm-scan-dev

# 回滚到指定版本
argocd app rollback evm-scan-dev <revision>
```

通过 Web UI：
1. 打开 Application
2. 点击 **History and Rollback**
3. 选择要回滚的版本
4. 点击 **Rollback**

### 4.2 通过 Git 回滚

```bash
# 查看 Kustomize 配置历史
git log --oneline deploys/overlays/<env>/kustomization.yaml

# 回滚到指定提交
git revert <commit-sha>
git push
```

### 4.3 通过 kubectl 回滚

```bash
# 查看 Deployment 历史
kubectl rollout history deployment/<deployment-name> -n evm-scan-<env>

# 回滚到上一版本
kubectl rollout undo deployment/<deployment-name> -n evm-scan-<env>

# 回滚到指定版本
kubectl rollout undo deployment/<deployment-name> --to-revision=<revision> -n evm-scan-<env>
```

---

## 5. 更新配置

### 5.1 更新 config.yaml

配置文件会在每次部署时自动从 `configs/config.yaml` 同步：

1. 修改 `configs/config.yaml`
2. 提交并推送到对应分支
3. 触发 GitHub Actions 部署
4. ConfigMap 会自动更新

### 5.2 更新 Secret (Dev 环境)

Dev 环境的 Secret 会自动同步：

1. 修改 `configs/secret.yaml`
2. 触发部署
3. Secret 会自动更新

### 5.3 更新 Secret (Prod 环境)

Prod 环境需要手动更新：

```bash
# 方法1: 直接编辑
kubectl edit secret evm-scan-secret -n evm-scan-prod

# 方法2: 从文件更新
kubectl create secret generic evm-scan-secret \
  --from-file=secret.yaml=/path/to/secret.yaml \
  --dry-run=client -o yaml | kubectl apply -f - -n evm-scan-prod
```

### 5.4 更新资源配额

修改对应环境的 patch 文件：

- Dev: `deploys/overlays/dev/patches/web-patch.yaml`
- Prod: `deploys/overlays/prod/patches/web-patch.yaml`

提交后触发部署或等待 ArgoCD 自动同步。

---

## 6. 日志查看

### 6.1 实时日志

```bash
# 查看 Web 应用日志
kubectl logs -f deployment/dev-evm-scan-web -n evm-scan-dev

# 查看 Scanner 日志
kubectl logs -f deployment/dev-evm-scan-scanner -n evm-scan-dev

# 查看特定 Pod 日志
kubectl logs -f <pod-name> -n evm-scan-<env>

# 查看之前容器的日志（如果重启过）
kubectl logs <pod-name> -n evm-scan-<env> --previous
```

### 6.2 历史日志

如果配置了日志收集系统（如 ELK、Loki），可以通过相应的 UI 查看历史日志。

### 6.3 Job 日志

```bash
# 查看 Job 列表
kubectl get jobs -n evm-scan-<env>

# 查看 Job 日志
kubectl logs job/<job-name> -n evm-scan-<env>

# 如果 Job 已完成，查看 Pod 日志
kubectl get pods -n evm-scan-<env> -l job-name=<job-name>
kubectl logs <pod-name> -n evm-scan-<env>
```

---

## 常用命令速查

### 部署相关

```bash
# 查看所有资源
kubectl get all -n evm-scan-dev

# 查看 Pod 状态
kubectl get pods -n evm-scan-dev -w

# 进入 Pod 调试
kubectl exec -it <pod-name> -n evm-scan-dev -- /bin/sh

# 端口转发（本地调试）
kubectl port-forward svc/dev-evm-scan-web 8000:8000 -n evm-scan-dev
```

### ArgoCD 相关

```bash
# 手动同步
argocd app sync evm-scan-dev

# 强制同步（忽略差异）
argocd app sync evm-scan-dev --force

# 刷新应用状态
argocd app get evm-scan-dev --refresh
```

### 故障排查

```bash
# 查看 Pod 详情
kubectl describe pod <pod-name> -n evm-scan-dev

# 查看事件
kubectl get events -n evm-scan-dev --sort-by='.lastTimestamp'

# 查看资源使用
kubectl top pods -n evm-scan-dev
```

---

## 下一步

- 如需添加新的 Job，请参考 [添加新 Job](./add-new-job.md)
- 遇到问题请参考 [故障排除](./troubleshooting.md)
