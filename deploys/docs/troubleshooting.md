# 故障排除指南

本文档收集常见问题及其解决方案。

## 目录

1. [GitHub Actions 问题](#1-github-actions-问题)
2. [Harbor 镜像问题](#2-harbor-镜像问题)
3. [Kubernetes 部署问题](#3-kubernetes-部署问题)
4. [ArgoCD 同步问题](#4-argocd-同步问题)
5. [应用运行问题](#5-应用运行问题)

---

## 1. GitHub Actions 问题

### 1.1 工作流未触发

**症状**: 点击 Run workflow 后没有反应

**可能原因**:
- 工作流文件语法错误
- 分支保护规则阻止

**解决方案**:
```bash
# 验证工作流语法
# 在本地安装 actionlint
brew install actionlint
actionlint .github/workflows/build-deploy.yaml
```

### 1.2 单元测试失败

**症状**: Test 阶段失败

**解决方案**:
```bash
# 本地运行测试查看详细错误
go test -v ./...

# 跳过特定测试（临时方案）
go test -v ./... -skip "TestXxx"
```

### 1.3 镜像推送失败: unauthorized

**症状**: `unauthorized: unauthorized to access repository`

**可能原因**:
- Harbor 凭证错误
- Robot Account 过期
- 权限不足

**解决方案**:
1. 检查 GitHub Secrets:
   - `HARBOR_USERNAME` 格式应为 `robot$project+name`
   - `HARBOR_PASSWORD` 是否正确

2. 在 Harbor 中验证:
   ```bash
   docker login harbor.sean.vip -u 'robot$coding+evm-scan'
   ```

3. 重新创建 Robot Account（如果过期）

### 1.4 Kustomize 更新失败

**症状**: `kustomize edit set image` 失败

**可能原因**:
- kustomization.yaml 格式错误
- 镜像名称不匹配

**解决方案**:
```bash
# 验证 kustomize 配置
cd deploys/overlays/dev
kustomize build .

# 检查镜像名称是否正确
cat kustomization.yaml | grep -A5 "images:"
```

### 1.5 Git push 失败: Permission denied

**症状**: `Permission denied to github-actions[bot]`

**可能原因**:
- GitHub Actions 没有写入权限

**解决方案**:
1. 检查仓库设置:
   - Settings → Actions → General
   - Workflow permissions → 选择 "Read and write permissions"

2. 或使用 Personal Access Token:
   - 创建 PAT 并添加到 Secrets
   - 修改 checkout 步骤使用 `token: ${{ secrets.PAT }}`

---

## 2. Harbor 镜像问题

### 2.1 镜像拉取失败: ImagePullBackOff

**症状**: Pod 状态显示 `ImagePullBackOff`

**诊断**:
```bash
kubectl describe pod <pod-name> -n evm-scan-dev
# 查看 Events 部分的错误信息
```

**可能原因与解决方案**:

1. **harbor-secret 不存在或错误**:
   ```bash
   # 检查 secret 是否存在
   kubectl get secret harbor-secret -n evm-scan-dev

   # 重新创建
   kubectl create secret docker-registry harbor-secret \
     --docker-server=harbor.sean.vip \
     --docker-username='robot$coding+evm-scan' \
     --docker-password='<token>' \
     -n evm-scan-dev
   ```

2. **镜像不存在**:
   ```bash
   # 检查镜像是否存在
   docker pull harbor.sean.vip/coding/evm-scan:<tag>
   ```

3. **网络问题**:
   ```bash
   # 在 K8s 节点上测试
   curl -v https://harbor.sean.vip/v2/
   ```

### 2.2 镜像推送速度慢

**解决方案**:
- 启用 Docker 缓存（已在 workflow 中配置）
- 检查网络带宽
- 考虑使用更近的镜像仓库

---

## 3. Kubernetes 部署问题

### 3.1 Pod CrashLoopBackOff

**症状**: Pod 反复重启

**诊断**:
```bash
# 查看 Pod 日志
kubectl logs <pod-name> -n evm-scan-dev

# 查看之前容器的日志
kubectl logs <pod-name> -n evm-scan-dev --previous

# 查看 Pod 详情
kubectl describe pod <pod-name> -n evm-scan-dev
```

**常见原因与解决方案**:

1. **配置文件错误**:
   ```bash
   # 检查 ConfigMap 内容
   kubectl get configmap evm-scan-config -n evm-scan-dev -o yaml
   ```

2. **Secret 缺失**:
   ```bash
   # 检查 Secret
   kubectl get secret evm-scan-secret -n evm-scan-dev
   ```

3. **启动命令错误**:
   ```bash
   # 检查 Deployment 的 command/args
   kubectl get deployment <name> -n evm-scan-dev -o yaml | grep -A10 "containers:"
   ```

### 3.2 Pod Pending

**症状**: Pod 一直处于 Pending 状态

**诊断**:
```bash
kubectl describe pod <pod-name> -n evm-scan-dev
```

**常见原因与解决方案**:

1. **资源不足**:
   ```bash
   # 检查节点资源
   kubectl describe nodes

   # 降低资源请求
   # 修改 deploys/overlays/<env>/patches/*-patch.yaml
   ```

2. **节点选择器不匹配**:
   ```bash
   # 检查节点标签
   kubectl get nodes --show-labels
   ```

### 3.3 Service 无法访问

**症状**: 无法通过 Service 访问应用

**诊断**:
```bash
# 检查 Service
kubectl get svc -n evm-scan-dev

# 检查 Endpoints
kubectl get endpoints -n evm-scan-dev

# 测试 Pod 连通性
kubectl exec -it <pod-name> -n evm-scan-dev -- wget -qO- localhost:8000/health
```

**解决方案**:
1. 确保 Pod 标签与 Service selector 匹配
2. 确保 Pod 健康检查通过
3. 检查网络策略（NetworkPolicy）

### 3.4 健康检查失败

**症状**: Pod 被反复重启，日志显示 `Liveness probe failed`

**诊断**:
```bash
kubectl describe pod <pod-name> -n evm-scan-dev | grep -A10 "Liveness:"
```

**解决方案**:
1. 增加 `initialDelaySeconds`:
   ```yaml
   livenessProbe:
     initialDelaySeconds: 60  # 增加启动等待时间
   ```

2. 检查健康检查端点:
   ```bash
   kubectl exec -it <pod-name> -n evm-scan-dev -- wget -qO- localhost:8000/health
   ```

---

## 4. ArgoCD 同步问题

### 4.1 同步状态: OutOfSync

**症状**: ArgoCD 显示 OutOfSync

**诊断**:
```bash
# 查看差异
argocd app diff evm-scan-dev

# 查看详情
argocd app get evm-scan-dev
```

**解决方案**:
```bash
# 手动同步
argocd app sync evm-scan-dev

# 强制同步（覆盖手动修改）
argocd app sync evm-scan-dev --force
```

### 4.2 同步状态: Unknown

**症状**: ArgoCD 无法获取应用状态

**可能原因**:
- Git 仓库无法访问
- Kustomize 配置错误

**解决方案**:
```bash
# 刷新应用
argocd app get evm-scan-dev --refresh

# 检查仓库连接
argocd repo list

# 验证 Kustomize 配置
cd deploys/overlays/dev && kustomize build .
```

### 4.3 健康状态: Degraded

**症状**: ArgoCD 显示 Degraded

**诊断**:
```bash
argocd app get evm-scan-dev
# 查看哪些资源不健康
```

**解决方案**:
1. 检查对应的 K8s 资源状态
2. 修复 Pod/Deployment 问题
3. 等待 ArgoCD 重新检测

### 4.4 Application 无法创建

**症状**: `Application creation failed`

**可能原因**:
- 路径不存在
- kustomization.yaml 语法错误

**解决方案**:
```bash
# 验证路径存在
ls deploys/overlays/dev/

# 验证 Kustomize 配置
cd deploys/overlays/dev && kustomize build .
```

---

## 5. 应用运行问题

### 5.1 数据库连接失败

**症状**: 日志显示数据库连接错误

**诊断**:
```bash
kubectl logs <pod-name> -n evm-scan-dev | grep -i "database\|postgres\|connection"
```

**解决方案**:
1. 检查配置文件中的数据库地址
2. 检查 K8s NetworkPolicy
3. 测试网络连通性:
   ```bash
   kubectl exec -it <pod-name> -n evm-scan-dev -- nc -zv <db-host> 5432
   ```

### 5.2 Redis 连接失败

**症状**: 日志显示 Redis 连接错误

**解决方案**:
```bash
# 测试 Redis 连通性
kubectl exec -it <pod-name> -n evm-scan-dev -- nc -zv <redis-host> 6379
```

### 5.3 内存溢出 (OOMKilled)

**症状**: Pod 被 OOMKilled

**诊断**:
```bash
kubectl describe pod <pod-name> -n evm-scan-dev | grep -i "oom\|memory"
```

**解决方案**:
1. 增加内存限制:
   ```yaml
   # 修改 patch 文件
   resources:
     limits:
       memory: "1Gi"  # 增加内存
   ```

2. 优化应用内存使用

### 5.4 CPU 限制导致性能问题

**症状**: 应用响应慢，CPU 使用率接近 100%

**诊断**:
```bash
kubectl top pod -n evm-scan-dev
```

**解决方案**:
1. 增加 CPU 限制:
   ```yaml
   resources:
     limits:
       cpu: "1000m"  # 增加 CPU
   ```

2. 增加副本数实现水平扩展

### 5.5 Job 执行超时

**症状**: Job 执行时间过长被终止

**解决方案**:
1. 增加 GitHub Actions 超时:
   ```yaml
   # 修改 run-job.yaml
   kubectl wait --timeout=3600s  # 增加到 1 小时
   ```

2. 增加 K8s Job 超时:
   ```yaml
   spec:
     activeDeadlineSeconds: 3600  # 1 小时
   ```

---

## 快速诊断命令

```bash
# 查看所有资源状态
kubectl get all -n evm-scan-dev

# 查看 Pod 日志
kubectl logs -f deployment/dev-evm-scan-web -n evm-scan-dev

# 查看事件（按时间排序）
kubectl get events -n evm-scan-dev --sort-by='.lastTimestamp'

# 查看资源使用
kubectl top pods -n evm-scan-dev

# 进入 Pod 调试
kubectl exec -it <pod-name> -n evm-scan-dev -- /bin/sh

# 检查 ArgoCD 状态
argocd app get evm-scan-dev --refresh
```

---

## 获取帮助

如果以上方案无法解决问题：

1. 收集诊断信息:
   ```bash
   kubectl describe pod <pod-name> -n evm-scan-dev > pod-describe.txt
   kubectl logs <pod-name> -n evm-scan-dev > pod-logs.txt
   kubectl get events -n evm-scan-dev > events.txt
   ```

2. 联系项目维护者并提供:
   - 问题描述
   - 复现步骤
   - 诊断信息文件
   - 相关截图
