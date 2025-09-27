#!/bin/bash

# ===================================================================
# 🛡️ Go Mastery 项目智能预提交质量检查钩子
# 目标: 在提交前确保零安全漏洞标准
# 版本: v1.0
# 最后更新: 2025年1月27日
# ===================================================================

set -e

# 配置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# 配置选项
ENABLE_SECURITY_CHECK=true
ENABLE_GO_VET_CHECK=true
ENABLE_BUILD_CHECK=true
ENABLE_GOLANGCI_LINT=false  # 预提交时默认关闭，避免过长等待
SKIP_ON_MERGE=true  # 合并提交时跳过检查

# 工具路径检测
GOSEC_CMD=""
GOLANGCI_LINT_CMD=""

# ===================================================================
# 🔧 工具检测和初始化
# ===================================================================

print_header() {
    echo -e "${CYAN}=============================================${NC}"
    echo -e "${WHITE}🛡️  Go Mastery 智能预提交质量检查${NC}"
    echo -e "${CYAN}=============================================${NC}"
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
    echo -e "${PURPLE}📋 $1${NC}"
}

# 检测工具是否可用
detect_tools() {
    print_step "检测质量检查工具..."

    # 检测Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go环境未安装或不在PATH中"
        exit 1
    fi

    # 检测gosec
    if command -v gosec &> /dev/null; then
        GOSEC_CMD="gosec"
        print_success "gosec 可用: $(gosec --version 2>/dev/null | head -1)"
    elif [ -f "$(go env GOPATH)/bin/gosec" ]; then
        GOSEC_CMD="$(go env GOPATH)/bin/gosec"
        print_success "gosec 可用: $($GOSEC_CMD --version 2>/dev/null | head -1)"
    else
        print_warning "gosec 未安装，将跳过安全检查"
        ENABLE_SECURITY_CHECK=false
    fi

    # 检测golangci-lint
    if command -v golangci-lint &> /dev/null; then
        GOLANGCI_LINT_CMD="golangci-lint"
        print_info "golangci-lint 可用: $(golangci-lint --version 2>/dev/null | head -1)"
    elif [ -f "$(go env GOPATH)/bin/golangci-lint" ]; then
        GOLANGCI_LINT_CMD="$(go env GOPATH)/bin/golangci-lint"
        print_info "golangci-lint 可用: $($GOLANGCI_LINT_CMD --version 2>/dev/null | head -1)"
    else
        print_warning "golangci-lint 未安装，质量检查功能受限"
    fi

    echo ""
}

# ===================================================================
# 🚨 安全漏洞检查 (P0级别)
# ===================================================================

check_security_vulnerabilities() {
    if [ "$ENABLE_SECURITY_CHECK" != "true" ]; then
        print_warning "安全检查已禁用，跳过..."
        return 0
    fi

    print_step "执行安全漏洞扫描 (P0级别检查)..."

    # 创建临时报告文件
    local temp_report=$(mktemp)
    local temp_json=$(mktemp)

    # 执行gosec扫描
    if $GOSEC_CMD -fmt json -out "$temp_json" ./... 2>/dev/null; then
        # 解析JSON结果
        if command -v jq &> /dev/null; then
            local vulnerability_count=$(jq '.Stats.found // 0' "$temp_json" 2>/dev/null)
        else
            # 如果没有jq，使用grep简单解析
            local vulnerability_count=$(grep -o '"found":[0-9]*' "$temp_json" 2>/dev/null | cut -d':' -f2 || echo "0")
        fi

        if [ "$vulnerability_count" -eq 0 ] 2>/dev/null; then
            print_success "安全扫描通过: 零安全漏洞状态维持"
            rm -f "$temp_report" "$temp_json"
            return 0
        else
            print_error "检测到 $vulnerability_count 个安全漏洞！"
            echo ""

            # 显示详细错误信息
            $GOSEC_CMD -fmt text ./... 2>/dev/null | head -20
            echo ""
            print_error "❌ 提交被阻止: 违反零安全漏洞要求"
            print_info "🔧 修复建议:"
            print_info "   1. 运行: gosec ./... 查看详细漏洞信息"
            print_info "   2. 参考: COMPREHENSIVE_QUALITY_REPORT.md 获取修复策略"
            print_info "   3. 修复后重新提交"
            echo ""

            rm -f "$temp_report" "$temp_json"
            return 1
        fi
    else
        print_error "安全扫描执行失败"
        rm -f "$temp_report" "$temp_json"
        return 1
    fi
}

# ===================================================================
# 🔧 Go官方工具验证 (P1级别)
# ===================================================================

check_go_vet() {
    if [ "$ENABLE_GO_VET_CHECK" != "true" ]; then
        return 0
    fi

    print_step "执行Go vet检查 (P1级别检查)..."

    local vet_output=$(mktemp)

    if go vet ./... 2>"$vet_output"; then
        print_success "Go vet检查通过"
        rm -f "$vet_output"
        return 0
    else
        print_error "Go vet检查失败"
        echo ""
        print_error "=== Go vet 警告详情 ==="
        cat "$vet_output"
        print_error "======================="
        echo ""
        print_info "🔧 修复建议:"
        print_info "   1. 修复上述Go vet警告"
        print_info "   2. 运行: go vet ./... 进行本地验证"
        echo ""

        rm -f "$vet_output"
        return 1
    fi
}

check_go_build() {
    if [ "$ENABLE_BUILD_CHECK" != "true" ]; then
        return 0
    fi

    print_step "执行Go build验证..."

    if go build ./... >/dev/null 2>&1; then
        print_success "Go build验证通过"
        return 0
    else
        print_error "Go build验证失败"
        echo ""
        print_error "=== 编译错误详情 ==="
        go build ./... 2>&1 | head -10
        print_error "=================="
        echo ""
        print_info "🔧 修复建议:"
        print_info "   1. 修复编译错误"
        print_info "   2. 运行: go build ./... 进行本地验证"
        echo ""
        return 1
    fi
}

# ===================================================================
# 📋 快速代码质量检查 (可选)
# ===================================================================

check_golangci_lint_quick() {
    if [ "$ENABLE_GOLANGCI_LINT" != "true" ] || [ -z "$GOLANGCI_LINT_CMD" ]; then
        return 0
    fi

    print_step "执行快速代码质量检查..."

    # 只检查修改的文件，避免耗时过长
    local changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

    if [ -z "$changed_files" ]; then
        print_info "没有Go文件被修改，跳过代码质量检查"
        return 0
    fi

    print_info "检查修改的文件: $(echo "$changed_files" | wc -l) 个"

    # 使用快速配置进行检查
    if $GOLANGCI_LINT_CMD run --fast --new-from-rev=HEAD~ --timeout=30s $changed_files >/dev/null 2>&1; then
        print_success "快速代码质量检查通过"
        return 0
    else
        print_warning "代码质量检查发现改进点"
        print_info "提示: 运行 'golangci-lint run' 查看详细信息"
        # 质量问题不阻止提交，只给出警告
        return 0
    fi
}

# ===================================================================
# 🔍 提交信息分析
# ===================================================================

analyze_commit() {
    # 检查是否为合并提交
    if [ "$SKIP_ON_MERGE" = "true" ] && git rev-parse --verify MERGE_HEAD >/dev/null 2>&1; then
        print_info "检测到合并提交，跳过质量检查"
        return 1
    fi

    # 获取暂存的Go文件
    local staged_go_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

    if [ -z "$staged_go_files" ]; then
        print_info "没有Go文件被修改，跳过质量检查"
        return 1
    fi

    print_info "检测到 $(echo "$staged_go_files" | wc -l) 个Go文件修改"
    return 0
}

# ===================================================================
# 🎯 智能跳过检查
# ===================================================================

should_skip_checks() {
    # 检查环境变量跳过选项
    if [ "$SKIP_PRECOMMIT_CHECKS" = "true" ]; then
        print_warning "环境变量设置跳过预提交检查"
        return 0
    fi

    # 检查提交消息中的跳过标记
    if git rev-parse --verify HEAD >/dev/null 2>&1; then
        local commit_msg=$(git log -1 --pretty=%B 2>/dev/null || echo "")
        if [[ "$commit_msg" =~ \[skip-checks\] ]] || [[ "$commit_msg" =~ \[no-verify\] ]]; then
            print_warning "提交消息中包含跳过标记"
            return 0
        fi
    fi

    return 1
}

# ===================================================================
# 📊 执行报告
# ===================================================================

generate_execution_report() {
    local start_time=$1
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    echo ""
    print_info "=== 预提交检查执行报告 ==="
    print_info "执行时间: ${duration}秒"
    print_info "检查项目: $([ "$ENABLE_SECURITY_CHECK" = "true" ] && echo "✅安全扫描" || echo "⏸️安全扫描") | $([ "$ENABLE_GO_VET_CHECK" = "true" ] && echo "✅Go vet" || echo "⏸️Go vet") | $([ "$ENABLE_BUILD_CHECK" = "true" ] && echo "✅构建验证" || echo "⏸️构建验证")"
    print_info "=========================="
    echo ""
}

# ===================================================================
# 🚀 主执行函数
# ===================================================================

main() {
    local start_time=$(date +%s)

    print_header

    # 智能跳过检查
    if should_skip_checks; then
        print_warning "跳过预提交检查"
        exit 0
    fi

    # 分析提交内容
    if ! analyze_commit; then
        exit 0
    fi

    # 检测工具
    detect_tools

    # 执行检查序列
    local checks_passed=true

    # P0级别: 安全检查 (必须通过)
    if ! check_security_vulnerabilities; then
        checks_passed=false
    fi

    # P1级别: Go官方工具验证 (必须通过)
    if [ "$checks_passed" = "true" ] && ! check_go_vet; then
        checks_passed=false
    fi

    if [ "$checks_passed" = "true" ] && ! check_go_build; then
        checks_passed=false
    fi

    # P2级别: 代码质量检查 (警告级别)
    if [ "$checks_passed" = "true" ]; then
        check_golangci_lint_quick || true  # 不影响提交结果
    fi

    # 生成执行报告
    generate_execution_report "$start_time"

    # 最终结果
    if [ "$checks_passed" = "true" ]; then
        print_success "🎉 所有关键检查通过，提交已允许"
        print_success "🛡️ 零安全漏洞状态已维持"
        exit 0
    else
        print_error "💥 关键检查失败，提交被阻止"
        print_error "🚨 请修复问题后重新提交"
        echo ""
        print_info "💡 快速修复提示:"
        print_info "   • 安全问题: 参考 COMPREHENSIVE_QUALITY_REPORT.md"
        print_info "   • 编译问题: 运行 go build ./... 检查"
        print_info "   • 紧急提交: 设置 SKIP_PRECOMMIT_CHECKS=true (不推荐)"
        echo ""
        exit 1
    fi
}

# 执行主函数
main "$@"