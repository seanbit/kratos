# 部署前准备指南

本文档介绍首次部署前需要完成的所有配置工作。

## 目录

1. [配置 GitHub Secrets](#1-配置-github-secrets)
2. [配置 Harbor](#2-配置-harbor)
3. [配置 Kubernetes](#3-配置-kubernetes)
4. [配置 ArgoCD](#4-配置-argocd)
5. [验证配置](#5-验证配置)

---

## 1. 配置 GitHub Secrets

在 GitHub 仓库中配置以下 Secrets：

**路径**: GitHub 仓库 → Settings → Secrets and variables → Actions → New repository secret

### 必需的 Secrets

| Secret 名称 | 说明 | 示例值 |
|------------|------|-------|
| `HARBOR_USERNAME` | Harbor 用户名或 Robot Account 名称 | `robot$evm-scan` |
| `HARBOR_PASSWORD` | Harbor 密码或 Robot Account Token | `xxxxxxxxxxxxx` |
| `KUBECONFIG` | K8s 配置文件 (Base64 编码) | 见下方说明 |

### 获取 KUBECONFIG

```bash
# 方法1: 编码整个 kubeconfig 文件
cat ~/.kube/config | base64 | tr -d '\n'

# 方法2: 如果有多个集群，先导出特定集群的配置
kubectl config view --minify --flatten | base64 | tr -d '\n'
```

**注意**: 确保 kubeconfig 中的 server 地址是 GitHub Actions 可以访问的地址（公网或通过 VPN）。

### 配置步骤截图说明

1. 打开 GitHub 仓库页面
2. 点击 **Settings** 标签
3. 左侧菜单选择 **Secrets and variables** → **Actions**
4. 点击 **New repository secret**
5. 输入 Name 和 Secret 值
6. 点击 **Add secret**

---

## 2. 配置 Harbor

### 2.1 创建项目

如果项目不存在，需要先创建：

1. 登录 Harbor: https://harbor.sean.vip
2. 点击 **+ 新建项目**
3. 项目名称: `coding`
4. 访问级别: 私有
5. 点击 **确定**

### 2.2 创建 Robot Account (推荐)

使用 Robot Account 比用户账号更安全：

1. 进入项目 `coding`
2. 点击 **机器人账户** 标签
3. 点击 **+ 添加机器人账户**
4. 配置:
   - 名称: `evm-scan`
   - 过期时间: 根据需要设置（建议 1 年）
   - 权限: 勾选 `push` 和 `pull`
5. 点击 **添加**
6. **重要**: 复制生成的 Token（只显示一次）

### 2.3 记录账号信息

将以下信息添加到 GitHub Secrets：

| 项目 | 值 |
|-----|---|
| HARBOR_USERNAME | `robot$coding+evm-scan` (注意格式) |
| HARBOR_PASSWORD | 生成的 Token |

---

## 3. 配置 Kubernetes

### 3.1 创建 Namespace

```bash
# 创建 dev 命名空间
kubectl create namespace evm-scan-dev

# 创建 prod 命名空间
kubectl create namespace evm-scan-prod
```

### 3.2 创建 Harbor Image Pull Secret

在每个命名空间中创建拉取镜像的 Secret：

```bash
# Dev 环境
kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.sean.vip \
  --docker-username='robot$coding+evm-scan' \
  --docker-password='<your-token>' \
  -n evm-scan-dev

# Prod 环境
kubectl create secret docker-registry harbor-secret \
  --docker-server=harbor.sean.vip \
  --docker-username='robot$coding+evm-scan' \
  --docker-password='<your-token>' \
  -n evm-scan-prod
```

### 3.3 创建 Prod Secret (手动)

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
kubectl create secret generic evm-scan-secret \
  --from-file=secret.yaml=/tmp/prod-secret.yaml \
  -n evm-scan-prod

# 删除本地文件
rm /tmp/prod-secret.yaml
```

### 3.4 配置 Self-hosted Runner RBAC

Self-hosted Runner 需要操作 `web3-analyse` namespace 中的 Job、Pod 等资源（用于 `Run K8s Job` 工作流）。

```bash
kubectl apply -f - <<'EOF'
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: job-runner
  namespace: web3-analyse
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
  namespace: web3-analyse
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

### 3.5 验证 Secret 创建

```bash
# 检查 dev 环境
kubectl get secrets -n evm-scan-dev

# 检查 prod 环境
kubectl get secrets -n evm-scan-prod

# 预期输出应包含:
# - harbor-secret
# - evm-scan-secret (prod)
```

---

## 4. 配置 ArgoCD

### 4.1 添加 Git 仓库

如果是私有仓库，需要在 ArgoCD 中添加仓库凭证：

```bash
# 使用 argocd CLI
argocd repo add https://github.com/<org>/evm-scan.git \
  --username <github-username> \
  --password <github-token>

# 或通过 ArgoCD Web UI:
# Settings → Repositories → + Connect Repo
```

### 4.2 创建 Dev 环境 Application

```yaml
# argocd-app-dev.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: evm-scan-dev
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/<org>/evm-scan.git
    targetRevision: dev
    path: deploys/overlays/dev
  destination:
    server: https://kubernetes.default.svc
    namespace: evm-scan-dev
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

### 4.3 创建 Prod 环境 Application

```yaml
# argocd-app-prod.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: evm-scan-prod
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://github.com/<org>/evm-scan.git
    targetRevision: main
    path: deploys/overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: evm-scan-prod
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

### 4.4 验证 ArgoCD Application

```bash
# 使用 CLI 查看
argocd app list

# 预期输出:
# NAME           CLUSTER                         NAMESPACE       PROJECT  STATUS  HEALTH
# evm-scan-dev   https://kubernetes.default.svc  evm-scan-dev    default  Synced  Healthy
# evm-scan-prod  https://kubernetes.default.svc  evm-scan-prod   default  Synced  Healthy
```

或通过 ArgoCD Web UI 查看应用状态。

---

## 5. 验证配置

### 5.1 验证 GitHub Actions

1. 进入 GitHub 仓库 → Actions
2. 选择 **Build and Deploy** 工作流
3. 点击 **Run workflow**
4. 选择:
   - app_type: `web`
   - environment: `dev`
   - enable_lint: `false`
5. 点击 **Run workflow** 执行
6. 查看执行日志，确保所有步骤成功

### 5.2 验证镜像推送

```bash
# 登录 Harbor 检查镜像是否存在
# 或使用 docker 命令
docker pull harbor.sean.vip/coding/evm-scan:web-<commit-sha>
```

### 5.3 验证 K8s 部署

```bash
# 检查 Pod 状态
kubectl get pods -n evm-scan-dev

# 检查 Deployment 状态
kubectl get deployments -n evm-scan-dev

# 检查服务
kubectl get svc -n evm-scan-dev

# 查看 Pod 日志
kubectl logs -f deployment/dev-evm-scan-web -n evm-scan-dev
```

### 5.4 验证 ArgoCD 同步

```bash
# 检查同步状态
argocd app get evm-scan-dev

# 手动触发同步（如果需要）
argocd app sync evm-scan-dev
```

---

## 配置清单检查表

在首次部署前，请确认以下所有项目已完成：

### GitHub Secrets
- [ ] `HARBOR_USERNAME` 已配置
- [ ] `HARBOR_PASSWORD` 已配置
- [ ] `KUBECONFIG` 已配置

### Harbor
- [ ] 项目 `coding` 已创建
- [ ] Robot Account 已创建
- [ ] 权限已配置 (push + pull)

### Kubernetes
- [ ] Namespace `evm-scan-dev` 已创建
- [ ] Namespace `evm-scan-prod` 已创建
- [ ] `harbor-secret` 已在两个命名空间创建
- [ ] `evm-scan-secret` 已在 prod 命名空间创建
- [ ] Self-hosted Runner RBAC 已配置（Role + RoleBinding）

### ArgoCD
- [ ] Git 仓库已添加（如果是私有仓库）
- [ ] Application `evm-scan-dev` 已创建
- [ ] Application `evm-scan-prod` 已创建
- [ ] 自动同步已启用

### 验证
- [ ] GitHub Actions 测试运行成功
- [ ] 镜像已推送到 Harbor
- [ ] Pod 在 K8s 中正常运行
- [ ] ArgoCD 显示 Synced 和 Healthy 状态

---

## 下一步

配置完成后，请参考 [使用指南](./usage-guide.md) 了解日常操作流程。
