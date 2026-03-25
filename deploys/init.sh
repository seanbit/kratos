#!/bin/bash
#
# 部署配置初始化脚本
# 用于将 K8s 清单中的占位符替换为实际项目配置
#

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "============================================"
echo "   K8s 部署配置初始化"
echo "============================================"
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 收集用户输入
read -p "应用名称 (如: my-app): " APP_NAME
if [[ -z "${APP_NAME}" ]]; then
    echo -e "${RED}错误: 应用名称不能为空${NC}"
    exit 1
fi

read -p "K8s Namespace (如: my-namespace): " NAMESPACE
if [[ -z "${NAMESPACE}" ]]; then
    echo -e "${RED}错误: Namespace 不能为空${NC}"
    exit 1
fi

read -p "Harbor Registry (如: harbor.example.com): " HARBOR_REGISTRY
if [[ -z "${HARBOR_REGISTRY}" ]]; then
    echo -e "${RED}错误: Harbor Registry 不能为空${NC}"
    exit 1
fi

read -p "Harbor Project (如: team/project): " HARBOR_PROJECT
if [[ -z "${HARBOR_PROJECT}" ]]; then
    echo -e "${RED}错误: Harbor Project 不能为空${NC}"
    exit 1
fi

read -p "Go Private Module (如: github.com/myorg, 可留空): " GOPRIVATE
# GOPRIVATE 可以为空

echo ""
echo "============================================"
echo "   配置摘要"
echo "============================================"
echo "应用名称:         ${APP_NAME}"
echo "Namespace:        ${NAMESPACE}"
echo "Harbor Registry:  ${HARBOR_REGISTRY}"
echo "Harbor Project:   ${HARBOR_PROJECT}"
echo "Go Private:       ${GOPRIVATE:-<未设置>}"
echo ""

read -p "确认配置? (y/N): " CONFIRM
if [[ "${CONFIRM}" != "y" && "${CONFIRM}" != "Y" ]]; then
    echo -e "${YELLOW}已取消${NC}"
    exit 0
fi

echo ""
echo -e "${GREEN}开始替换占位符...${NC}"

# 替换 K8s 清单中的占位符 (使用 {{PLACEHOLDER}} 格式)
find "${SCRIPT_DIR}" -name "*.yaml" -type f | while read -r file; do
    # 替换 {{PLACEHOLDER}} 格式
    sed -i.bak \
        -e "s|{{APP_NAME}}|${APP_NAME}|g" \
        -e "s|{{NAMESPACE}}|${NAMESPACE}|g" \
        -e "s|{{HARBOR_REGISTRY}}|${HARBOR_REGISTRY}|g" \
        -e "s|{{HARBOR_PROJECT}}|${HARBOR_PROJECT}|g" \
        -e "s|{{GOPRIVATE}}|${GOPRIVATE}|g" \
        "${file}"

    # 删除备份文件
    rm -f "${file}.bak"
done

echo -e "${GREEN}配置初始化完成!${NC}"
echo ""
echo "下一步:"
echo "1. 检查 deploys/ 目录下的配置文件"
echo "2. 修改 deploys/overlays/*/postgres-endpoints.yaml 中的数据库地址"
echo "3. 修改 deploys/overlays/*/config.yaml 中的环境配置"
echo ""
echo "注意: GitHub Actions 配置需要在服务仓库的 Settings > Variables 中设置:"
echo "  - APP_NAME: ${APP_NAME}"
echo "  - NAMESPACE: ${NAMESPACE}"
echo "  - HARBOR_REGISTRY: ${HARBOR_REGISTRY}"
echo "  - HARBOR_PROJECT: ${HARBOR_PROJECT}"
echo "  - GOPRIVATE: ${GOPRIVATE}"
echo "  - RUNNER_LABELS: [\"self-hosted\", \"linux\"]"
echo ""
