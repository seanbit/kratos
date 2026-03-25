# 部署前准备指南

本文档介绍首次部署前需要完成的所有配置工作。

## 目录

1. [CD 仓库配置 (deploys)](#1-cd-仓库配置-deploys)
2. [服务仓库配置 (GitHub Variables)](#2-服务仓库配置-github-variables)
3. [配置 GitHub Secrets](#3-配置-github-secrets)
4. [配置 Harbor](#4-配置-harbor)
5. [配置 Kubernetes](#5-配置-kubernetes)
6. [配置 ArgoCD](#6-配置-argocd)
7. [验证配置](#7-验证配置)

---

## 1. CD 仓库配置 (deploys)

### 1.1 运行初始化脚本

首次使用 deploys 仓库时，需要运行初始化脚本来配置 K8s 清单：

```bash
cd deploys
./init.sh
```

脚本会引导你输入以下参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| 应用名称 | 项目的唯一标识符 | `my-app` |
| Namespace | Kubernetes 命名空间前缀 | `my-namespace` |
| Harbor Registry | 镜像仓库地址 | `harbor.example.com` |
| Harbor Project | 镜像项目路径 | `team/project` |
| Go Private Module | Go 私有模块前缀（可选） | `github.com/myorg` |

### 1.2 初始化后的文件

初始化脚本会替换 K8s 清单中的占位符：

- `{{APP_NAME}}` → 你的应用名称
- `{{NAMESPACE}}` → 你的命名空间前缀
- `{{HARBOR_REGISTRY}}` → 你的镜像仓库地址
- `{{HARBOR_PROJECT}}` → 你的镜像项目路径
- `{{GOPRIVATE}}` → Go 私有模块配置

### 1.3 手动调整项

初始化后，还需要手动修改：

1. **数据库配置**: `deploys/overlays/*/postgres-endpoints.yaml`
2. **环境配置**: `deploys/overlays/*/config.yaml`
3. **资源配置**: 根据实际需求调整各 deployment 的资源配额

---

## 2. 服务仓库配置 (GitHub Variables)

服务仓库（基于 template 创建）的 GitHub Actions 使用 GitHub Variables 进行配置。

### 2.1 配置路径

GitHub 仓库 → Settings → Secrets and variables → Actions → **Variables** 标签

### 2.2 必需的 Variables

| Variable 名称 | 说明 | 示例值 |
|--------------|------|--------|
| `APP_NAME` | 应用名称（与 deploys 一致） | `my-app` |
| `NAMESPACE` | K8s 命名空间前缀 | `my-namespace` |
| `HARBOR_REGISTRY` | Harbor 镜像仓库地址 | `harbor.example.com` |
| `HARBOR_PROJECT` | Harbor 项目路径 | `team/project` |
| `RUNNER_LABELS` | Runner 标签 JSON 数组 | `["self-hosted", "linux"]` |

### 2.3 可选的 Variables

| Variable 名称 | 说明 | 何时需要 |
|--------------|------|---------|
| `GOPRIVATE` | Go 私有模块前缀 | 使用私有 Go 模块时 |

### 2.4 配置步骤

1. 打开 GitHub 仓库页面
2. 点击 **Settings** 标签
3. 左侧菜单选择 **Secrets and variables** → **Actions**
4. 点击 **Variables** 标签
5. 点击 **New repository variable**
6. 输入 Name 和 Value
7. 点击 **Add variable**

### 2.5 RUNNER_LABELS 格式说明

`RUNNER_LABELS` 必须是 JSON 数组格式：

```json
["self-hosted", "linux", "x64"]
```

如果只使用 GitHub 托管的 runner，可以设置为：

```json
["ubuntu-latest"]
```

---

## 3. 配置 GitHub Secrets

在 GitHub 仓库中配置以下 Secrets：

**路径**: GitHub 仓库 → Settings → Secrets and variables → Actions → **Secrets** 标签

### 必需的 Secrets

| Secret 名称 | 说明 | 示例值 |
|------------|------|-------|
| `CI_GITHUB_TOKEN` | GitHub Token (需 repo 权限) | `ghp_xxxxxxxxxxxxx` |
| `HARBOR_USERNAME` | Harbor 用户名或 Robot Account 名称 | `robot$my-app` |
| `HARBOR_PASSWORD` | Harbor 密码或 Robot Account Token | `xxxxxxxxxxxxx` |

### 可选的 Secrets

| Secret 名称 | 说明 | 何时需要 |
|------------|------|---------|
| `KUBECONFIG` | K8s 配置文件 (Base64 编码) | 如果需要 runner 直接访问 K8s |

### 获取 KUBECONFIG

```bash
# 方法1: 编码整个 kubeconfig 文件
cat ~/.kube/config | base64 | tr -d '\n'

# 方法2: 如果有多个集群，先导出特定集群的配置
kubectl config view --minify --flatten | base64 | tr -d '\n'
```

**注意**: 确保 kubeconfig 中的 server 地址是 GitHub Actions 可以访问的地址（公网或通过 VPN）。

### 配置步骤

1. 打开 GitHub 仓库页面
2. 点击 **Settings** 标签
3. 左侧菜单选择 **Secrets and variables** → **Actions**
4. 点击 **New repository secret**
5. 输入 Name 和 Secret 值
6. 点击 **Add secret**

---

## 4. 配置 Harbor

### 4.1 创建项目

如果项目不存在，需要先创建：

1. 登录 Harbor
2. 点击 **+ 新建项目**
3. 项目名称: 按照 Variables 中配置的 Harbor Project
4. 访问级别: 私有
5. 点击 **确定**

### 4.2 创建 Robot Account (推荐)

使用 Robot Account 比用户账号更安全：

1. 进入项目
2. 点击 **机器人账户** 标签
3. 点击 **+ 添加机器人账户**
4. 配置:
   - 名称: 应用名称
   - 过期时间: 根据需要设置（建议 1 年）
   - 权限: 勾选 `push` 和 `pull`
5. 点击 **添加**
6. **重要**: 复制生成的 Token（只显示一次）

### 4.3 记录账号信息

将以下信息添加到 GitHub Secrets：

| 项目 | 值 |
|-----|---|
| HARBOR_USERNAME | `robot$<project>+<name>` (注意格式) |
| HARBOR_PASSWORD | 生成的 Token |

---

## 5. 配置 Kubernetes

### 5.1 创建 Namespace

```bash
# 创建 dev 命名空间
kubectl create namespace <NAMESPACE>-dev

# 创建 prod 命名空间
kubectl create namespace <NAMESPACE>-prod
```

### 5.2 创建 Harbor Image Pull Secret

在每个命名空间中创建拉取镜像的 Secret：

```bash
# Dev 环境
kubectl create secret docker-registry harbor-secret \
  --docker-server=<HARBOR_REGISTRY> \
  --docker-username='robot$<project>+<name>' \
  --docker-password='<your-token>' \
  -n <NAMESPACE>-dev

# Prod 环境
kubectl create secret docker-registry harbor-secret \
  --docker-server=<HARBOR_REGISTRY> \
  --docker-username='robot$<project>+<name>' \
  --docker-password='<your-token>' \
  -n <NAMESPACE>-prod
```

### 5.3 创建 Prod Secret (手动)

Prod 环境的 Secret 需要手动创建和管理：

```bash
# 创建 secret.yaml 文件 (本地，不要提交到 Git)
cat > /tmp/prod-secret.yaml << 'EOF'
# Prod 环境敏感配置
database:
  password: "your-prod-db-password"
redis:
  password: "your-prod-redis-password"
jwt:
  key_25519: "your-prod-jwt-key"
# ... 其他敏感配置
EOF

# 创建 Secret
kubectl create secret generic <APP_NAME>-secret \
  --from-file=secret.yaml=/tmp/prod-secret.yaml \
  -n <NAMESPACE>-prod

# 删除本地文件
rm /tmp/prod-secret.yaml
```

### 5.4 配置 Self-hosted Runner RBAC

Self-hosted Runner 需要操作 namespace 中的 Job、Pod 等资源（用于 `Run K8s Job` 工作流）。

```bash
kubectl apply -f - <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: job-runner
  namespace: <NAMESPACE>
rules:
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: job-runner-binding
  namespace: <NAMESPACE>
subjects:
  - kind: ServiceAccount
    name: default
    namespace: runners
roleRef:
  kind: Role
  name: job-runner
  apiGroup: rbac.authorization.k8s.io
EOF
```

**权限说明**:

| 资源 | 操作 | 用途 |
|------|------|------|
| `jobs.batch` | get, list, watch, create, delete | 创建 Job、等待完成、查询状态 |
| `pods` | get, list | 查找 Job 对应的 Pod |
| `pods/log` | get | 获取 Job 执行日志 |

### 5.5 验证 Secret 创建

```bash
# 检查 dev 环境
kubectl get secrets -n <NAMESPACE>-dev

# 检查 prod 环境
kubectl get secrets -n <NAMESPACE>-prod

# 预期输出应包含:
# - harbor-secret
# - <APP_NAME>-secret (prod)
```

---

## 6. 配置 ArgoCD

### 6.1 添加 Git 仓库

如果是私有仓库，需要在 ArgoCD 中添加仓库凭证：

```bash
# 使用 argocd CLI
argocd repo add https://github.com/<org>/<repo>.git \
  --username <github-username> \
  --password <github-token>

# 或通过 ArgoCD Web UI:
# Settings → Repositories → + Connect Repo
```

### 6.2 创建 Dev 环境 Application

```yaml
# argocd-app-dev.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: <APP_NAME>-dev
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/<org>/deploys.git
    targetRevision: main
    path: overlays/dev
  destination:
    server: https://kubernetes.default.svc
    namespace: <NAMESPACE>-dev
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

应用配置：

```bash
kubectl apply -f argocd-app-dev.yaml
```

### 6.3 创建 Prod 环境 Application

```yaml
# argocd-app-prod.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: <APP_NAME>-prod
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/<org>/deploys.git
    targetRevision: main
    path: overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: <NAMESPACE>-prod
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

应用配置：

```bash
kubectl apply -f argocd-app-prod.yaml
```

### 6.4 验证 ArgoCD Application

```bash
# 使用 CLI 查看
argocd app list

# 预期输出:
# NAME              CLUSTER                         NAMESPACE        PROJECT  STATUS  HEALTH
# <APP_NAME>-dev    https://kubernetes.default.svc  <NAMESPACE>-dev  default  Synced  Healthy
# <APP_NAME>-prod   https://kubernetes.default.svc  <NAMESPACE>-prod default  Synced  Healthy
```

或通过 ArgoCD Web UI 查看应用状态。

---

## 7. 验证配置

### 7.1 验证 GitHub Actions

1. 进入 GitHub 服务仓库 → Actions
2. 选择 **Build and Deploy** 工作流
3. 点击 **Run workflow**
4. 选择:
   - app_type: `web`
   - environment: `dev`
   - enable_lint: `false`
5. 点击 **Run workflow** 执行
6. 查看执行日志，确保所有步骤成功

### 7.2 验证镜像推送

```bash
# 登录 Harbor 检查镜像是否存在
# 或使用 docker 命令
docker pull <HARBOR_REGISTRY>/<HARBOR_PROJECT>:web-<commit-sha>
```

### 7.3 验证 K8s 部署

```bash
# 检查 Pod 状态
kubectl get pods -n <NAMESPACE>-dev

# 检查 Deployment 状态
kubectl get deployments -n <NAMESPACE>-dev

# 检查服务
kubectl get svc -n <NAMESPACE>-dev

# 查看 Pod 日志
kubectl logs -f deployment/dev-<APP_NAME>-web -n <NAMESPACE>-dev
```

### 7.4 验证 ArgoCD 同步

```bash
# 检查同步状态
argocd app get <APP_NAME>-dev

# 手动触发同步（如果需要）
argocd app sync <APP_NAME>-dev
```

---

## 配置清单检查表

在首次部署前，请确认以下所有项目已完成：

### CD 仓库 (deploys)
- [ ] 运行 `deploys/init.sh` 完成配置初始化
- [ ] 手动调整数据库地址等配置

### 服务仓库 (GitHub Variables)
- [ ] `APP_NAME` 已配置
- [ ] `NAMESPACE` 已配置
- [ ] `HARBOR_REGISTRY` 已配置
- [ ] `HARBOR_PROJECT` 已配置
- [ ] `RUNNER_LABELS` 已配置

### GitHub Secrets
- [ ] `CI_GITHUB_TOKEN` 已配置
- [ ] `HARBOR_USERNAME` 已配置
- [ ] `HARBOR_PASSWORD` 已配置
- [ ] `KUBECONFIG` 已配置（如果需要）

### Harbor
- [ ] 项目已创建
- [ ] Robot Account 已创建
- [ ] 权限已配置 (push + pull)

### Kubernetes
- [ ] Namespace `<NAMESPACE>-dev` 已创建
- [ ] Namespace `<NAMESPACE>-prod` 已创建
- [ ] `harbor-secret` 已在两个命名空间创建
- [ ] `<APP_NAME>-secret` 已在 prod 命名空间创建
- [ ] Self-hosted Runner RBAC 已配置（如果需要）

### ArgoCD
- [ ] Git 仓库已添加（如果是私有仓库）
- [ ] Application `<APP_NAME>-dev` 已创建
- [ ] Application `<APP_NAME>-prod` 已创建
- [ ] 自动同步已启用

### 验证
- [ ] GitHub Actions 测试运行成功
- [ ] 镜像已推送到 Harbor
- [ ] Pod 在 K8s 中正常运行
- [ ] ArgoCD 显示 Synced 和 Healthy 状态

---

## 下一步

配置完成后，请参考 [使用指南](./usage-guide.md) 了解日常操作流程。
