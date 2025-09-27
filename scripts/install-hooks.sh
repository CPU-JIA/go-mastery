#!/bin/bash

# ===================================================================
# 🔧 Go Mastery 项目 - 预提交钩子安装脚本
# 目标: 自动安装和配置预提交质量检查钩子
# 支持: Windows (Git Bash), Linux, macOS
# 版本: v1.0
# ===================================================================

set -e

# 配置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# 路径配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
PRE_COMMIT_HOOK="$GIT_HOOKS_DIR/pre-commit"
PRE_COMMIT_SCRIPT="$SCRIPT_DIR/pre-commit-hook.sh"

print_header() {
    echo -e "${CYAN}================================================${NC}"
    echo -e "${WHITE}🔧 Go Mastery 预提交钩子安装器${NC}"
    echo -e "${CYAN}================================================${NC}"
}

print_step() {
    echo -e "${BLUE}🔍 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}📋 $1${NC}"
}

# 检查环境
check_environment() {
    print_step "检查环境..."

    # 检查是否在Git仓库中
    if [ ! -d "$PROJECT_ROOT/.git" ]; then
        print_error "错误: 当前目录不是Git仓库"
        exit 1
    fi

    # 检查钩子脚本是否存在
    if [ ! -f "$PRE_COMMIT_SCRIPT" ]; then
        print_error "错误: 预提交钩子脚本未找到: $PRE_COMMIT_SCRIPT"
        exit 1
    fi

    # 创建hooks目录（如果不存在）
    mkdir -p "$GIT_HOOKS_DIR"

    print_success "环境检查通过"
}

# 备份现有钩子
backup_existing_hook() {
    if [ -f "$PRE_COMMIT_HOOK" ]; then
        local backup_file="$PRE_COMMIT_HOOK.backup.$(date +%Y%m%d_%H%M%S)"
        print_step "备份现有预提交钩子..."
        cp "$PRE_COMMIT_HOOK" "$backup_file"
        print_success "已备份到: $backup_file"
    fi
}

# 安装钩子
install_hook() {
    print_step "安装预提交钩子..."

    # 复制钩子脚本
    cp "$PRE_COMMIT_SCRIPT" "$PRE_COMMIT_HOOK"

    # 设置执行权限
    chmod +x "$PRE_COMMIT_HOOK"

    print_success "预提交钩子安装完成"
}

# 安装依赖工具
install_dependencies() {
    print_step "检查和安装依赖工具..."

    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go环境未安装，请先安装Go"
        exit 1
    fi
    print_success "Go环境: $(go version)"

    # 检查并安装gosec
    if ! command -v gosec &> /dev/null && [ ! -f "$(go env GOPATH)/bin/gosec" ]; then
        print_step "安装gosec安全扫描器..."
        if curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v2.21.4; then
            print_success "gosec安装成功"
        else
            print_warning "gosec安装失败，可以手动安装: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
        fi
    else
        print_success "gosec已安装"
    fi

    # 检查golangci-lint
    if ! command -v golangci-lint &> /dev/null && [ ! -f "$(go env GOPATH)/bin/golangci-lint" ]; then
        print_step "检查golangci-lint..."
        print_info "golangci-lint未安装，但可选"
        print_info "安装命令: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.64.8"
    else
        print_success "golangci-lint已安装"
    fi
}

# 测试钩子
test_hook() {
    print_step "测试预提交钩子..."

    # 创建临时测试文件
    local test_file="test_hook_file.go"
    cat > "$test_file" << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Test hook file")
}
EOF

    # 暂存测试文件
    git add "$test_file" 2>/dev/null || true

    # 执行钩子测试
    if "$PRE_COMMIT_HOOK" 2>/dev/null; then
        print_success "预提交钩子测试通过"
    else
        print_warning "预提交钩子测试失败，但钩子已安装"
    fi

    # 清理测试文件
    git reset "$test_file" 2>/dev/null || true
    rm -f "$test_file"
}

# 配置选项
configure_options() {
    print_step "配置预提交钩子选项..."

    echo ""
    print_info "预提交钩子配置选项:"
    print_info "================================"
    print_info "环境变量配置 (在 ~/.bashrc 或 ~/.zshrc 中设置):"
    echo ""
    print_info "# 跳过所有预提交检查 (紧急情况使用)"
    print_info "export SKIP_PRECOMMIT_CHECKS=true"
    echo ""
    print_info "# Git命令跳过选项"
    print_info "git commit --no-verify  # 跳过预提交钩子"
    echo ""
    print_info "# 提交消息跳过标记"
    print_info "git commit -m \"fix: emergency fix [skip-checks]\""
    echo ""
    print_info "================================"
}

# 显示使用说明
show_usage_instructions() {
    echo ""
    print_info "🎉 安装完成！预提交钩子已配置"
    echo ""
    print_info "📋 使用说明:"
    print_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    print_info "• 正常提交: git commit -m \"your message\""
    print_info "• 钩子将自动检查:"
    print_info "  ✓ 安全漏洞扫描 (gosec)"
    print_info "  ✓ Go vet验证"
    print_info "  ✓ 编译检查"
    print_info "  ⚠ 代码质量提示 (golangci-lint)"
    echo ""
    print_info "🚨 紧急提交选项:"
    print_info "• 跳过钩子: git commit --no-verify"
    print_info "• 环境变量: SKIP_PRECOMMIT_CHECKS=true git commit"
    print_info "• 消息标记: git commit -m \"fix [skip-checks]\""
    echo ""
    print_info "🔧 钩子管理:"
    print_info "• 卸载钩子: rm $PRE_COMMIT_HOOK"
    print_info "• 重新安装: 重新运行此脚本"
    print_info "• 钩子位置: $PRE_COMMIT_HOOK"
    echo ""
    print_success "🛡️ 零安全漏洞状态保护已激活"
}

# 主函数
main() {
    print_header
    echo ""

    check_environment
    backup_existing_hook
    install_hook
    install_dependencies
    test_hook
    configure_options
    show_usage_instructions

    echo ""
    print_success "🎯 预提交质量保障系统安装完成！"
}

# 执行主函数
main "$@"