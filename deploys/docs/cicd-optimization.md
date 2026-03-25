# CI/CD 优化方案

## 一、架构概述

### 约束条件

| 条件 | 说明 |
|---|---|
| Harbor 镜像仓库 | 内网部署，不对外网开放 |
| K8s 集群 | 内网部署，不对外网开放 |
| CI 平台 | GitHub Actions（云端 runner + 内网 self-hosted runner） |
| CD 平台 | ArgoCD（监听 Git 仓库变更，自动同步到 K8s） |

### 优化后部署流程

```
代码推送
  │
  ├── PR → ci.yaml (lint + build 验证)
  │
  └── 手动触发 build-deploy.yaml
        │
        ├── lint (可选，云端 runner)
        │
        ├── build (云端 runner)
        │     ├── 参数校验
        │     ├── Go 定向编译目标二进制 (CGO_ENABLED=0)
        │     ├── docker build (Dockerfile.runtime, 仅 alpine + 二进制)
        │     ├── docker save + gzip 导出镜像 tarball
        │     └── 上传镜像 artifact (image.tar.gz)
        │
        ├── push-to-harbor (内网 self-hosted runner)
        │     ├── 下载镜像 artifact
        │     ├── docker load 加载镜像
        │     ├── docker push → Harbor
        │     └── 清理本地镜像 + 过期缓存
        │
        └── update-manifests (云端 runner)
              ├── yq 更新 kustomization.yaml 镜像标签
              ├── diff 检查 config 是否变化
              ├── git commit + push (带 rebase 重试)
              └── ArgoCD 自动检测并同步到 K8s
```

## 二、已实施的优化

### 1. 删除不可用的 build-deploy.yaml（直推 Harbor 方案）

**问题**：原 `build-deploy.yaml` 在云端 runner 上直接 `docker push` 到内网 Harbor，因网络不通永远无法成功。

**修改**：删除该文件，将 `build-deploy-hybrid.yaml` 重命名为 `build-deploy.yaml` 作为唯一的部署工作流。

### 2. 云端构建镜像 + 内网加载推送

**问题**：内网 self-hosted runner 处于受限网络环境，无法访问 Docker Hub 拉取基础镜像（如 `alpine:3.18.3`），在内网构建 Docker 镜像会因网络超时失败。

**修改**：
- 云端 runner 完成 Go 编译 + Docker 镜像构建（云端可正常访问 Docker Hub）
- 云端通过 `docker save | gzip` 导出镜像为 `image.tar.gz`，上传为 GitHub Artifact
- 内网 runner 下载 artifact，通过 `docker load` 加载镜像，再 `docker push` 到 Harbor
- 内网 runner 无需访问任何外部镜像仓库

**相关文件**：
- `deploys/Dockerfile.runtime` — 仅包含运行时层的轻量 Dockerfile
- `.github/workflows/build-deploy.yaml` — build job 负责编译 + 构建镜像 + 导出 tarball

### 3. Dockerfile 优化

**问题**：
- `COPY .. .` 无 `.dockerignore`，构建上下文包含 `.git/`（36MB）等无用文件
- 依赖下载（`go mod download`）与源码变更耦合，无法利用 Docker 层缓存
- Builder 阶段安装了 vim/gdb/protobuf 等无用调试工具
- Runtime 阶段包含无意义的 Go 环境变量和 `wget` 下载

**修改**：
- 添加 `.dockerignore`（排除 .git, .github, .idea, bin/, docs 等）
- 前置 `COPY go.mod go.sum` + `RUN go mod download`，利用层缓存
- Builder 阶段只保留 `git gcc g++ make`
- Runtime 阶段移除 GOPRIVATE/GO111MODULE 等无用 ENV，移除 wget

### 4. 定向编译

**问题**：`make build` 执行 `go build ./...`，编译 web/scanner/job 全部三个二进制，但每次 CI 只部署一个 app_type，浪费约 2/3 编译时间。

**修改**：CI 中直接执行定向编译：
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-X main.Version=${VERSION}" \
  -o ./bin/${APP_TYPE} \
  ./cmd/${APP_TYPE}
```

> 注意：使用 `CGO_ENABLED=0` 产出静态链接二进制，确保在 Alpine 上运行兼容。如果项目后续引入 CGO 依赖（如 go-sqlite3），编译会报错，届时需改为在 Alpine Docker 容器中编译或回退到 Dockerfile 全量构建方案。

### 5. 合并 validate/summary 独立 Job

**问题**：validate（仅一个 if 判断）和 summary（仅 echo 输出）各自独占一个 ubuntu runner，每个 runner 启动需要 20-40 秒。

**修改**：
- validate 合并为 build job 的第一个 step
- summary 合并为 update-manifests job 的最后一个 step
- 减少两次 runner 启动开销

### 6. Lint 阻断 Build

**问题**：原方案 lint 和 build 并行运行，即使 lint 失败，build 仍然继续，lint 形同虚设。

**修改**：
```yaml
build:
  needs: [lint]
  if: always() && (needs.lint.result == 'success' || needs.lint.result == 'skipped')
```
- lint 启用时：lint 失败 → build 不执行
- lint 未启用时：lint 被跳过 → build 正常执行

### 7. PR 自动 CI 验证

**问题**：所有工作流仅支持手动触发（workflow_dispatch），PR 没有自动 CI 门禁检查。

**修改**：新增 `.github/workflows/ci.yaml`，在 PR 推送到 main/dev 时自动触发：
- lint：golangci-lint 代码检查
- build：matrix 策略并行编译 web/scanner/job，验证编译通过
- test：单元测试（当前已预留，注释状态，待项目编写测试用例后启用）

### 8. run-job.yaml 改为内网 Runner

**问题**：`run-job.yaml` 使用 `ubuntu-latest` 云端 runner 执行 kubectl 操作，但 K8s 集群仅内网可达，kubectl 命令永远无法连接。

**修改**：`runs-on` 改为 `[self-hosted, github-runner, coding-seanbit]`。

同时修复命名空间硬编码：原 `NAMESPACE="evm-scan-${{ inputs.environment }}"` 改为 `NAMESPACE="web3-analyse"`（与 Kustomize 配置一致）。

### 9. Config 更新 Diff 检查

**问题**：每次部署都执行 `cp configs/config.yaml → deploys/overlays/dev/config.yaml`，即使配置未变也会产生 git commit 和 ArgoCD sync。

**修改**：先执行 `diff -q` 比较，仅在配置实际变化时才复制。

### 10. Git Push 竞态处理

**问题**：两个人同时部署 web 和 scanner 时，`git push` 可能冲突失败。

**修改**：添加 rebase 重试机制：
```bash
for i in 1 2 3; do
  git push && break
  echo "Push failed (attempt $i), retrying with rebase..."
  git pull --rebase
done
```

### 11. Self-Hosted Runner 磁盘清理

**问题**：长期运行后内网 runner 的磁盘被历史镜像和构建缓存占满。

**修改**：在 push-to-harbor 的 cleanup step 中添加：
```bash
docker system prune -f --filter "until=72h"
```

### 12. 修复 Lint 跳过导致下游 Job 被跳过

**问题**：GitHub Actions 已知行为（runner#2205）—— 当 `lint` 被跳过时，`build` 通过 `if: always()` 正常运行并成功，但 "skipped" 状态沿依赖链传播，导致 `push-to-harbor` 和 `update-manifests` 也被跳过。

**修改**：为下游 Job 添加显式条件，绕过 skip 状态传播：
```yaml
push-to-harbor:
  if: ${{ !cancelled() && needs.build.result == 'success' }}

update-manifests:
  if: ${{ !cancelled() && needs.build.result == 'success' && needs.push-to-harbor.result == 'success' }}
```

使用 `!cancelled()` 而非 `always()`，确保手动取消 workflow 时不会继续执行。

### 13. Self-Hosted Runner 工作目录持久化

**问题**：K8s 部署的 self-hosted runner 将工作目录挂载在宿主机 `/tmp` 下，Linux 系统重启后 `/tmp` 被清理，导致 runner 无法创建 `_temp` 目录而报错 `FileNotFoundException`。

**修改**：
- `hostPath.path` 从 `/tmp/github-runner-coding-seanbit` 改为 `/data/github-runner-coding-seanbit`
- 添加 `type: DirectoryOrCreate`，目录不存在时自动创建
- `RUNNER_WORKDIR` 和 `mountPath` 同步修改

## 三、文件变更清单

| 操作 | 文件 | 说明 |
|---|---|---|
| 删除 | `.github/workflows/build-deploy.yaml`（原直推方案） | 不可用的死代码 |
| 删除 | `.github/workflows/build-deploy-hybrid.yaml` | 被优化版替代 |
| 新建 | `.github/workflows/build-deploy.yaml` | 优化后的统一部署工作流 |
| 新建 | `.github/workflows/ci.yaml` | PR 自动 CI 验证 |
| 新建 | `.dockerignore` | 减少构建上下文 |
| 新建 | `deploys/Dockerfile.runtime` | CI 专用轻量运行时 Dockerfile |
| 修改 | `deploys/Dockerfile` | 层缓存优化、清理无用依赖 |
| 修改 | `.github/workflows/run-job.yaml` | 内网 runner + namespace 修复 |
| 修改 | `deploys/base/scanner-deployment.yaml` | liveness probe 进程名修复 |
| 修改 | `deploys/jobs/event-re-dispatch.yaml` | 二进制名 + Cobra 命令格式修复 |

## 四、后续可选优化路径

以下优化投入较高，建议在当前方案稳定运行后按需引入。

### A. 应用仓库与 GitOps 仓库分离

**现状**：CI 将 kustomization.yaml 和 config 变更 commit 到应用代码仓库，导致：
- Git 历史被 `ci: update xxx` 自动提交污染
- ArgoCD 无法区分代码提交和部署提交
- 并发部署存在 push 冲突风险

**方案**：
1. 创建独立的 GitOps 仓库（如 `evm-scan-deploy`），仅存放 `deploys/` 目录
2. CI 完成镜像构建后，向 GitOps 仓库提交清单变更
3. ArgoCD 指向 GitOps 仓库
4. 应用代码仓库保持纯净

### B. ArgoCD Image Updater

**现状**：CI 负责构建镜像 + 更新清单 + commit/push，流程较长且有竞态风险。

**方案**：
1. 部署 [ArgoCD Image Updater](https://argocd-image-updater.readthedocs.io/)
2. 配置监听 Harbor 仓库中的 `evm-scan` 镜像
3. CI 只负责构建和推送镜像，不再修改 Git 清单
4. Image Updater 自动检测新镜像标签，更新 Kustomize 配置并触发同步
5. 彻底消除 CI commit 到 Git 的需求

### C. 自动部署触发

**现状**：部署仅通过 `workflow_dispatch` 手动触发。

**方案**：在 `build-deploy.yaml` 中添加 push 触发：
```yaml
on:
  push:
    branches:
      - dev      # 推送到 dev 分支自动部署到 dev 环境
      - main     # 推送到 main 分支自动部署到 prod 环境
```

需要解决的问题：
- 自动触发时无法通过 `inputs` 指定 app_type，需要按文件路径检测变更或构建全部三个组件
- prod 部署建议添加 GitHub Environment 的 approval 审批
