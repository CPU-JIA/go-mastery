#!/bin/bash

# ===================================================================
# 🚀 Go Mastery 智能质量保障体系一键部署脚本
# 目标: 快速部署完整的CI/CD质量保障系统
# 版本: v1.0
# 最后更新: 2025年1月27日
# ===================================================================

set -e

# 配置颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

# 项目信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

print_header() {
    echo -e "${CYAN}================================================${NC}"
    echo -e "${WHITE}🚀 Go Mastery 智能质量保障体系${NC}"
    echo -e "${WHITE}   一键部署脚本${NC}"
    echo -e "${CYAN}================================================${NC}"
    echo ""
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
    print_step "检查运行环境..."

    # 检查操作系统
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        print_info "检测到Windows环境 (Git Bash)"
        OS_TYPE="windows"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        print_info "检测到macOS环境"
        OS_TYPE="macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        print_info "检测到Linux环境"
        OS_TYPE="linux"
    else
        print_warning "未知操作系统: $OSTYPE"
        OS_TYPE="unknown"
    fi

    # 检查Git仓库
    if [ ! -d "$PROJECT_ROOT/.git" ]; then
        print_error "当前目录不是Git仓库"
        print_info "请在Git仓库根目录运行此脚本"
        exit 1
    fi

    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go环境未安装"
        print_info "请先安装Go 1.24或更高版本"
        exit 1
    fi

    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go环境检查通过: $go_version"

    echo ""
}

# 安装依赖工具
install_dependencies() {
    print_step "安装质量保障工具..."

    # 安装gosec
    if ! command -v gosec &> /dev/null && [ ! -f "$(go env GOPATH)/bin/gosec" ]; then
        print_info "安装gosec安全扫描器..."
        if curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v2.21.4; then
            print_success "gosec安装成功"
        else
            print_warning "gosec安装失败，尝试Go安装方式..."
            go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        fi
    else
        print_success "gosec已安装"
    fi

    # 安装golangci-lint
    if ! command -v golangci-lint &> /dev/null && [ ! -f "$(go env GOPATH)/bin/golangci-lint" ]; then
        print_info "安装golangci-lint代码质量检查器..."
        if curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v1.64.8; then
            print_success "golangci-lint安装成功"
        else
            print_warning "golangci-lint安装失败，尝试Go安装方式..."
            go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
        fi
    else
        print_success "golangci-lint已安装"
    fi

    # 检查工具是否在PATH中
    local gopath_bin="$(go env GOPATH)/bin"
    if [[ ":$PATH:" != *":$gopath_bin:"* ]]; then
        print_warning "$(go env GOPATH)/bin 不在PATH中"
        print_info "请将以下行添加到你的shell配置文件 (~/.bashrc, ~/.zshrc 等):"
        print_info "export PATH=\$PATH:\$(go env GOPATH)/bin"
        echo ""
    fi

    echo ""
}

# 安装预提交钩子
install_precommit_hooks() {
    print_step "安装预提交钩子..."

    if [ -f "$PROJECT_ROOT/scripts/install-hooks.sh" ]; then
        if bash "$PROJECT_ROOT/scripts/install-hooks.sh"; then
            print_success "预提交钩子安装成功"
        else
            print_error "预提交钩子安装失败"
            return 1
        fi
    else
        print_error "预提交钩子安装脚本不存在"
        return 1
    fi

    echo ""
}

# 验证GitHub Actions配置
verify_github_actions() {
    print_step "验证GitHub Actions配置..."

    local workflow_file="$PROJECT_ROOT/.github/workflows/quality-assurance.yml"

    if [ -f "$workflow_file" ]; then
        print_success "GitHub Actions工作流配置存在"
        print_info "工作流将在下次推送时自动激活"
    else
        print_error "GitHub Actions工作流配置缺失"
        return 1
    fi

    echo ""
}

# 执行初始质量检查
run_initial_quality_check() {
    print_step "执行初始质量检查..."

    # 检查脚本是否存在
    if [ ! -f "$PROJECT_ROOT/scripts/quality-monitor.sh" ]; then
        print_error "质量监控脚本不存在"
        return 1
    fi

    print_info "运行质量监控脚本..."
    if bash "$PROJECT_ROOT/scripts/quality-monitor.sh"; then
        print_success "初始质量检查完成"

        # 显示报告位置
        if [ -f "$PROJECT_ROOT/quality-reports/latest_quality_report.md" ]; then
            print_info "质量报告已生成: quality-reports/latest_quality_report.md"
        fi
    else
        print_warning "质量检查过程中出现问题，但系统已部署"
        print_info "请手动运行: bash scripts/quality-monitor.sh"
    fi

    echo ""
}

# 创建使用说明
create_usage_guide() {
    print_step "创建快速使用说明..."

    local usage_file="$PROJECT_ROOT/QUICK_START.md"

    cat > "$usage_file" << 'EOF'
# 🚀 Go Mastery 质量保障系统快速开始

## 📋 系统已部署完成

恭喜！Go Mastery 智能质量保障体系已成功部署到你的项目中。

## 🔧 核心功能

### 1. 预提交检查
每次 `git commit` 时自动执行：
- 🛡️ 安全漏洞扫描
- 🔧 Go vet验证
- 🏗️ 编译检查

### 2. CI/CD流水线
代码推送后自动执行：
- 📊 完整质量分析
- 🚨 自动警报
- 📋 详细报告

### 3. 质量监控
定期执行质量检查：
```bash
# 手动执行质量监控
bash scripts/quality-monitor.sh

# 查看质量报告
cat quality-reports/latest_quality_report.md
```

### 4. 趋势分析
深度分析质量趋势：
```bash
# 执行趋势分析
bash scripts/quality-trends.sh

# 查看分析结果
cat quality-reports/analysis/latest_analysis.md
```

## 🚨 紧急情况

如需跳过质量检查（不推荐）：
```bash
# 跳过预提交检查
git commit --no-verify

# 临时禁用检查
export SKIP_PRECOMMIT_CHECKS=true
git commit -m "emergency fix"
```

## 📚 详细文档

- [完整质量保障文档](QUALITY_ASSURANCE.md)
- [质量优化策略](QUALITY_OPTIMIZATION_STRATEGY.md)
- [综合质量报告](COMPREHENSIVE_QUALITY_REPORT.md)

## 🎯 目标

维护**零安全漏洞**状态，确保代码质量达到企业级标准。

---
*系统已就绪，开始你的高质量Go开发之journey！*
EOF

    print_success "快速使用说明已创建: QUICK_START.md"
    echo ""
}

# 显示部署总结
show_deployment_summary() {
    print_step "部署总结"

    echo -e "${GREEN}🎉 Go Mastery 智能质量保障体系部署完成！${NC}"
    echo ""

    echo -e "${CYAN}✅ 已部署组件:${NC}"
    echo "   🔧 预提交钩子 - 提交前质量检查"
    echo "   🔄 GitHub Actions - CI/CD自动化"
    echo "   📊 质量监控系统 - 持续质量跟踪"
    echo "   📈 趋势分析系统 - 深度数据分析"
    echo ""

    echo -e "${CYAN}📋 下一步操作:${NC}"
    echo "   1. 阅读快速开始指南: cat QUICK_START.md"
    echo "   2. 查看质量报告: cat quality-reports/latest_quality_report.md"
    echo "   3. 测试提交: git commit -m \"test: verify quality system\""
    echo "   4. 推送代码触发CI: git push"
    echo ""

    echo -e "${CYAN}🔍 验证系统:${NC}"
    echo "   # 检查预提交钩子"
    echo "   ls -la .git/hooks/pre-commit"
    echo ""
    echo "   # 手动运行质量检查"
    echo "   bash scripts/quality-monitor.sh"
    echo ""
    echo "   # 查看工具版本"
    echo "   gosec --version"
    echo "   golangci-lint --version"
    echo ""

    echo -e "${CYAN}📚 重要文档:${NC}"
    echo "   - QUALITY_ASSURANCE.md     # 完整系统文档"
    echo "   - QUICK_START.md           # 快速开始指南"
    echo "   - quality-reports/         # 质量报告目录"
    echo ""

    echo -e "${GREEN}🛡️ 零安全漏洞目标已激活！${NC}"
    echo -e "${YELLOW}💡 提示: 第一次提交时，预提交钩子会执行质量检查${NC}"
}

# 主执行函数
main() {
    print_header

    print_info "开始部署 Go Mastery 智能质量保障体系..."
    echo ""

    # 执行部署步骤
    check_environment
    install_dependencies
    install_precommit_hooks
    verify_github_actions
    run_initial_quality_check
    create_usage_guide
    show_deployment_summary

    echo ""
    print_success "🎯 部署完成！你的Go项目现在具备企业级质量保障能力。"
}

# 执行主函数
main "$@"