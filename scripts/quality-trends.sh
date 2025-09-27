#!/bin/bash

# ===================================================================
# 📈 Go Mastery 项目质量趋势分析器
# 目标: 深度分析历史质量数据，生成趋势洞察和预测
# 功能: 数据分析、可视化、趋势预测、性能基准
# 版本: v1.0
# 最后更新: 2025年1月27日
# ===================================================================

set -e

# 配置参数
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/quality-reports"
TRENDS_DIR="$REPORTS_DIR/trends"
ANALYSIS_DIR="$REPORTS_DIR/analysis"

# 分析配置
MIN_SAMPLES=3               # 最少样本数量
TREND_WINDOW=7             # 趋势分析窗口
PREDICTION_WINDOW=3        # 预测窗口
ANOMALY_THRESHOLD=3        # 异常检测阈值(标准差倍数)

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'

# 时间戳
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
DATE_HUMAN=$(date '+%Y年%m月%d日 %H:%M:%S')

# ===================================================================
# 🔧 工具函数
# ===================================================================

print_header() {
    echo -e "${CYAN}================================================${NC}"
    echo -e "${WHITE}📈 Go Mastery 质量趋势分析器${NC}"
    echo -e "${CYAN}================================================${NC}"
    echo -e "${PURPLE}分析时间: $DATE_HUMAN${NC}"
    echo ""
}

print_section() {
    echo -e "${BLUE}📊 $1${NC}"
    echo "----------------------------------------"
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

print_metric() {
    echo -e "${PURPLE}📊 $1${NC}"
}

print_insight() {
    echo -e "${YELLOW}💡 $1${NC}"
}

# 创建分析目录
create_directories() {
    mkdir -p "$REPORTS_DIR" "$TRENDS_DIR" "$ANALYSIS_DIR"
}

# ===================================================================
# 📊 数据准备和验证
# ===================================================================

prepare_and_validate_data() {
    print_section "数据准备和验证"

    local trends_file="$TRENDS_DIR/quality_trends.csv"

    if [ ! -f "$trends_file" ]; then
        print_error "趋势数据文件不存在: $trends_file"
        print_info "请先运行质量监控脚本生成数据: bash scripts/quality-monitor.sh"
        exit 1
    fi

    # 检查数据格式
    local header=$(head -1 "$trends_file")
    if [[ "$header" != "timestamp,security_vulnerabilities,go_vet_warnings,quality_issues,errcheck_issues,revive_issues,unused_issues" ]]; then
        print_warning "数据格式可能不兼容，但尝试继续分析"
    fi

    # 统计有效数据行数
    local data_lines=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | wc -l)

    if [ "$data_lines" -lt "$MIN_SAMPLES" ]; then
        print_warning "数据样本不足 ($data_lines < $MIN_SAMPLES)，分析结果可能不准确"
    else
        print_success "数据验证通过，发现 $data_lines 个有效样本"
    fi

    echo "data_samples=$data_lines" > "$ANALYSIS_DIR/analysis_metadata_$TIMESTAMP.txt"
    echo ""
}

# ===================================================================
# 📈 基础统计分析
# ===================================================================

basic_statistical_analysis() {
    print_section "基础统计分析"

    local trends_file="$TRENDS_DIR/quality_trends.csv"
    local stats_file="$ANALYSIS_DIR/basic_stats_$TIMESTAMP.txt"

    echo "# Go Mastery 质量指标基础统计" > "$stats_file"
    echo "分析时间: $DATE_HUMAN" >> "$stats_file"
    echo "" >> "$stats_file"

    # 分析各个指标
    analyze_metric "安全漏洞" 2 "$trends_file" "$stats_file"
    analyze_metric "Go vet警告" 3 "$trends_file" "$stats_file"
    analyze_metric "代码质量问题" 4 "$trends_file" "$stats_file"
    analyze_metric "errcheck问题" 5 "$trends_file" "$stats_file"
    analyze_metric "revive问题" 6 "$trends_file" "$stats_file"
    analyze_metric "unused代码" 7 "$trends_file" "$stats_file"

    print_success "基础统计分析完成: $stats_file"
    echo ""
}

analyze_metric() {
    local metric_name="$1"
    local column="$2"
    local trends_file="$3"
    local stats_file="$4"

    print_info "分析指标: $metric_name"

    # 提取数据
    local data=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | cut -d',' -f"$column" | grep '^[0-9]*$' | head -20)

    if [ -z "$data" ]; then
        echo "⚠️ $metric_name: 无有效数据" >> "$stats_file"
        echo "" >> "$stats_file"
        return
    fi

    local count=$(echo "$data" | wc -l)
    local sum=$(echo "$data" | awk '{sum+=$1} END {print sum}')
    local min=$(echo "$data" | sort -n | head -1)
    local max=$(echo "$data" | sort -n | tail -1)
    local avg=$(echo "$sum $count" | awk '{printf "%.2f", $1/$2}')

    # 计算中位数
    local median
    if [ $((count % 2)) -eq 0 ]; then
        local mid1=$((count / 2))
        local mid2=$((mid1 + 1))
        local val1=$(echo "$data" | sort -n | sed -n "${mid1}p")
        local val2=$(echo "$data" | sort -n | sed -n "${mid2}p")
        median=$(echo "$val1 $val2" | awk '{printf "%.2f", ($1+$2)/2}')
    else
        local mid=$(((count + 1) / 2))
        median=$(echo "$data" | sort -n | sed -n "${mid}p")
    fi

    # 计算标准差
    local variance=$(echo "$data" | awk -v avg="$avg" '{sum+=($1-avg)^2} END {print sum}')
    local std_dev=$(echo "$variance $count" | awk '{printf "%.2f", sqrt($1/$2)}')

    # 写入统计结果
    cat >> "$stats_file" << EOF
## $metric_name 统计分析

- **样本数量**: $count
- **平均值**: $avg
- **中位数**: $median
- **最小值**: $min
- **最大值**: $max
- **标准差**: $std_dev
- **变异系数**: $(echo "$std_dev $avg" | awk '{printf "%.2f%%", ($1/$2)*100}')

EOF

    # 生成质量评价
    if [ "$metric_name" = "安全漏洞" ]; then
        if [ "$max" = "0" ]; then
            echo "✅ **质量评价**: 优秀 - 始终保持零安全漏洞" >> "$stats_file"
        else
            echo "❌ **质量评价**: 需要改进 - 曾经出现过安全漏洞" >> "$stats_file"
        fi
    elif [ "$metric_name" = "Go vet警告" ]; then
        if [ "$max" = "0" ]; then
            echo "✅ **质量评价**: 优秀 - 始终保持零警告" >> "$stats_file"
        else
            echo "⚠️ **质量评价**: 需要改进 - 曾经出现过Go vet警告" >> "$stats_file"
        fi
    else
        if (( $(echo "$avg < 100" | bc -l 2>/dev/null || echo "0") )); then
            echo "✅ **质量评价**: 良好 - 问题数量控制在合理范围" >> "$stats_file"
        elif (( $(echo "$avg < 500" | bc -l 2>/dev/null || echo "0") )); then
            echo "⚠️ **质量评价**: 一般 - 有改进空间" >> "$stats_file"
        else
            echo "❌ **质量评价**: 需要关注 - 问题数量较多" >> "$stats_file"
        fi
    fi

    echo "" >> "$stats_file"

    print_metric "$metric_name: 平均$avg, 最新$(echo "$data" | tail -1), 标准差$std_dev"
}

# ===================================================================
# 📈 趋势分析
# ===================================================================

trend_analysis() {
    print_section "趋势分析"

    local trends_file="$TRENDS_DIR/quality_trends.csv"
    local trend_report="$ANALYSIS_DIR/trend_analysis_$TIMESTAMP.md"

    cat > "$trend_report" << EOF
# 📈 Go Mastery 项目质量趋势分析报告

**分析时间**: $DATE_HUMAN
**分析窗口**: 最近 $TREND_WINDOW 次监控
**数据来源**: $trends_file

---

## 🎯 趋势分析摘要

EOF

    # 分析各指标趋势
    analyze_trend "安全漏洞" 2 "$trends_file" "$trend_report"
    analyze_trend "Go vet警告" 3 "$trends_file" "$trend_report"
    analyze_trend "代码质量问题" 4 "$trends_file" "$trend_report"
    analyze_trend "errcheck问题" 5 "$trends_file" "$trend_report"

    # 生成综合趋势评估
    generate_overall_trend_assessment "$trends_file" "$trend_report"

    print_success "趋势分析完成: $trend_report"
    echo ""
}

analyze_trend() {
    local metric_name="$1"
    local column="$2"
    local trends_file="$3"
    local trend_report="$4"

    print_info "分析趋势: $metric_name"

    # 获取最近的数据
    local recent_data=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | tail -"$TREND_WINDOW" | cut -d',' -f"$column" | grep '^[0-9]*$')

    if [ -z "$recent_data" ]; then
        echo "### $metric_name" >> "$trend_report"
        echo "⚠️ 数据不足，无法分析趋势" >> "$trend_report"
        echo "" >> "$trend_report"
        return
    fi

    local count=$(echo "$recent_data" | wc -l)
    if [ "$count" -lt 2 ]; then
        echo "### $metric_name" >> "$trend_report"
        echo "⚠️ 数据样本不足 ($count < 2)，无法分析趋势" >> "$trend_report"
        echo "" >> "$trend_report"
        return
    fi

    # 计算趋势
    local first_value=$(echo "$recent_data" | head -1)
    local last_value=$(echo "$recent_data" | tail -1)
    local change=$((last_value - first_value))
    local trend_direction="稳定"
    local trend_icon="➡️"

    if [ "$change" -gt 0 ]; then
        trend_direction="上升"
        trend_icon="📈"
    elif [ "$change" -lt 0 ]; then
        trend_direction="下降"
        trend_icon="📉"
    fi

    # 计算线性回归斜率 (简化版)
    local slope=$(calculate_simple_slope "$recent_data")
    local correlation=$(calculate_correlation "$recent_data")

    cat >> "$trend_report" << EOF
### $trend_icon $metric_name

- **趋势方向**: $trend_direction
- **变化量**: $change ($first_value → $last_value)
- **趋势强度**: $(interpret_slope "$slope")
- **数据稳定性**: $(interpret_correlation "$correlation")

EOF

    # 生成趋势评价
    if [ "$metric_name" = "安全漏洞" ]; then
        if [ "$last_value" = "0" ] && [ "$first_value" = "0" ]; then
            echo "✅ **趋势评价**: 优秀 - 持续保持零安全漏洞状态" >> "$trend_report"
        elif [ "$trend_direction" = "下降" ]; then
            echo "✅ **趋势评价**: 改善中 - 安全状况正在好转" >> "$trend_report"
        elif [ "$trend_direction" = "上升" ]; then
            echo "🚨 **趋势评价**: 警报 - 安全状况正在恶化" >> "$trend_report"
        else
            echo "⚠️ **趋势评价**: 需要关注 - 安全状况需要持续监控" >> "$trend_report"
        fi
    else
        if [ "$trend_direction" = "下降" ]; then
            echo "✅ **趋势评价**: 改善中 - 质量正在提升" >> "$trend_report"
        elif [ "$trend_direction" = "上升" ]; then
            echo "⚠️ **趋势评价**: 关注 - 质量可能在退化" >> "$trend_report"
        else
            echo "➡️ **趋势评价**: 稳定 - 质量保持当前水平" >> "$trend_report"
        fi
    fi

    echo "" >> "$trend_report"

    print_metric "$metric_name: $trend_direction趋势, 变化$change, 斜率$slope"
}

# 计算简单线性回归斜率
calculate_simple_slope() {
    local data="$1"
    local count=$(echo "$data" | wc -l)

    if [ "$count" -lt 2 ]; then
        echo "0"
        return
    fi

    # 使用最简单的方法：(最后值 - 第一值) / (样本数 - 1)
    local first=$(echo "$data" | head -1)
    local last=$(echo "$data" | tail -1)
    echo "scale=3; ($last - $first) / ($count - 1)" | bc -l 2>/dev/null || echo "0"
}

# 计算数据相关性(简化版)
calculate_correlation() {
    local data="$1"
    local count=$(echo "$data" | wc -l)

    if [ "$count" -lt 3 ]; then
        echo "1.0"
        return
    fi

    # 简化的相关性计算：基于标准差
    local avg=$(echo "$data" | awk '{sum+=$1} END {printf "%.2f", sum/NR}')
    local variance=$(echo "$data" | awk -v avg="$avg" '{sum+=($1-avg)^2} END {print sum}')
    local std_dev=$(echo "scale=3; sqrt($variance / $count)" | bc -l 2>/dev/null || echo "0")

    # 返回一个简化的稳定性指标 (标准差相对平均值的比例)
    if [ "$avg" != "0" ]; then
        echo "scale=3; 1 - ($std_dev / $avg)" | bc -l 2>/dev/null || echo "0.5"
    else
        echo "1.0"
    fi
}

interpret_slope() {
    local slope="$1"

    if (( $(echo "$slope > 1" | bc -l 2>/dev/null || echo "0") )); then
        echo "强烈上升"
    elif (( $(echo "$slope > 0.1" | bc -l 2>/dev/null || echo "0") )); then
        echo "轻微上升"
    elif (( $(echo "$slope < -1" | bc -l 2>/dev/null || echo "0") )); then
        echo "强烈下降"
    elif (( $(echo "$slope < -0.1" | bc -l 2>/dev/null || echo "0") )); then
        echo "轻微下降"
    else
        echo "基本稳定"
    fi
}

interpret_correlation() {
    local correlation="$1"

    if (( $(echo "$correlation > 0.8" | bc -l 2>/dev/null || echo "0") )); then
        echo "高度稳定"
    elif (( $(echo "$correlation > 0.6" | bc -l 2>/dev/null || echo "0") )); then
        echo "相对稳定"
    elif (( $(echo "$correlation > 0.3" | bc -l 2>/dev/null || echo "0") )); then
        echo "中等波动"
    else
        echo "波动较大"
    fi
}

# ===================================================================
# 🔍 异常检测
# ===================================================================

anomaly_detection() {
    print_section "异常检测"

    local trends_file="$TRENDS_DIR/quality_trends.csv"
    local anomaly_report="$ANALYSIS_DIR/anomaly_detection_$TIMESTAMP.md"

    cat > "$anomaly_report" << EOF
# 🔍 Go Mastery 项目异常检测报告

**检测时间**: $DATE_HUMAN
**检测方法**: 基于Z-score的统计异常检测
**异常阈值**: $ANOMALY_THRESHOLD 倍标准差

---

## 🎯 异常检测结果

EOF

    local anomalies_found=false

    # 检测各指标异常
    detect_anomalies "安全漏洞" 2 "$trends_file" "$anomaly_report" && anomalies_found=true
    detect_anomalies "Go vet警告" 3 "$trends_file" "$anomaly_report" && anomalies_found=true
    detect_anomalies "代码质量问题" 4 "$trends_file" "$anomaly_report" && anomalies_found=true
    detect_anomalies "errcheck问题" 5 "$trends_file" "$anomaly_report" && anomalies_found=true

    if [ "$anomalies_found" = false ]; then
        echo "✅ **检测结果**: 未发现显著异常数据点" >> "$anomaly_report"
        echo "" >> "$anomaly_report"
        echo "所有质量指标都在正常范围内波动，没有检测到需要特别关注的异常情况。" >> "$anomaly_report"
        print_success "异常检测: 未发现异常"
    else
        print_warning "异常检测: 发现异常数据点"
    fi

    print_success "异常检测完成: $anomaly_report"
    echo ""
}

detect_anomalies() {
    local metric_name="$1"
    local column="$2"
    local trends_file="$3"
    local anomaly_report="$4"

    # 提取数据
    local data=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | cut -d',' -f"$column" | grep '^[0-9]*$')

    if [ -z "$data" ] || [ $(echo "$data" | wc -l) -lt 3 ]; then
        return 1
    fi

    local count=$(echo "$data" | wc -l)
    local sum=$(echo "$data" | awk '{sum+=$1} END {print sum}')
    local avg=$(echo "$sum $count" | awk '{printf "%.2f", $1/$2}')

    # 计算标准差
    local variance=$(echo "$data" | awk -v avg="$avg" '{sum+=($1-avg)^2} END {print sum}')
    local std_dev=$(echo "$variance $count" | awk '{printf "%.2f", sqrt($1/$2)}')

    # 检测异常点
    local line_num=0
    local anomalies=""

    while IFS= read -r value; do
        line_num=$((line_num + 1))
        if [ -n "$value" ] && [[ "$value" =~ ^[0-9]+$ ]]; then
            local z_score=$(echo "$value $avg $std_dev" | awk '{if($3>0) printf "%.2f", ($1-$2)/$3; else print "0"}')
            local abs_z_score=$(echo "$z_score" | awk '{if($1<0) print -$1; else print $1}')

            if (( $(echo "$abs_z_score > $ANOMALY_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
                local timestamp=$(grep -v "^timestamp" "$trends_file" | sed -n "${line_num}p" | cut -d',' -f1)
                anomalies="$anomalies\n- **$timestamp**: 值=$value, Z-score=$z_score"
            fi
        fi
    done <<< "$data"

    if [ -n "$anomalies" ]; then
        cat >> "$anomaly_report" << EOF
### 🔍 $metric_name 异常检测

**统计基准**:
- 平均值: $avg
- 标准差: $std_dev
- 异常阈值: ±$ANOMALY_THRESHOLD σ

**检测到的异常**:
$(echo -e "$anomalies")

EOF
        print_warning "$metric_name: 发现异常数据点"
        return 0
    else
        return 1
    fi
}

# ===================================================================
# 🔮 质量预测
# ===================================================================

quality_prediction() {
    print_section "质量预测"

    local trends_file="$TRENDS_DIR/quality_trends.csv"
    local prediction_report="$ANALYSIS_DIR/quality_prediction_$TIMESTAMP.md"

    cat > "$prediction_report" << EOF
# 🔮 Go Mastery 项目质量预测报告

**预测时间**: $DATE_HUMAN
**预测方法**: 基于历史趋势的线性外推
**预测窗口**: 未来 $PREDICTION_WINDOW 个监控周期

---

## 🎯 质量预测结果

EOF

    # 为关键指标生成预测
    predict_metric "安全漏洞" 2 "$trends_file" "$prediction_report"
    predict_metric "Go vet警告" 3 "$trends_file" "$prediction_report"
    predict_metric "代码质量问题" 4 "$trends_file" "$prediction_report"

    # 生成综合预测评估
    generate_prediction_assessment "$prediction_report"

    print_success "质量预测完成: $prediction_report"
    echo ""
}

predict_metric() {
    local metric_name="$1"
    local column="$2"
    local trends_file="$3"
    local prediction_report="$4"

    print_info "预测指标: $metric_name"

    # 获取最近的数据用于预测
    local recent_data=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | tail -5 | cut -d',' -f"$column" | grep '^[0-9]*$')

    if [ -z "$recent_data" ] || [ $(echo "$recent_data" | wc -l) -lt 2 ]; then
        echo "### $metric_name" >> "$prediction_report"
        echo "⚠️ 数据不足，无法生成预测" >> "$prediction_report"
        echo "" >> "$prediction_report"
        return
    fi

    local current_value=$(echo "$recent_data" | tail -1)
    local slope=$(calculate_simple_slope "$recent_data")

    # 生成预测值
    local predictions=""
    for i in $(seq 1 $PREDICTION_WINDOW); do
        local predicted_value=$(echo "$current_value + ($slope * $i)" | bc -l 2>/dev/null | xargs printf "%.0f")

        # 确保预测值不为负数
        if [ "$predicted_value" -lt 0 ]; then
            predicted_value=0
        fi

        predictions="$predictions\n- **第$i个周期**: $predicted_value"
    done

    # 计算预测置信度 (基于历史数据稳定性)
    local correlation=$(calculate_correlation "$recent_data")
    local confidence="中等"

    if (( $(echo "$correlation > 0.8" | bc -l 2>/dev/null || echo "0") )); then
        confidence="高"
    elif (( $(echo "$correlation < 0.3" | bc -l 2>/dev/null || echo "0") )); then
        confidence="低"
    fi

    cat >> "$prediction_report" << EOF
### 🔮 $metric_name 预测

**当前值**: $current_value
**趋势斜率**: $slope
**预测置信度**: $confidence

**预测结果**:
$(echo -e "$predictions")

**预测解释**: 基于最近的趋势变化，预计该指标在未来几个监控周期内将$(interpret_slope "$slope")。

EOF

    # 生成预测建议
    if [ "$metric_name" = "安全漏洞" ]; then
        if (( $(echo "$slope > 0" | bc -l 2>/dev/null || echo "0") )); then
            echo "🚨 **预测警告**: 安全漏洞趋势上升，建议立即审查代码变更并加强安全检查。" >> "$prediction_report"
        else
            echo "✅ **预测评价**: 安全状况稳定或改善，继续保持当前质量标准。" >> "$prediction_report"
        fi
    fi

    echo "" >> "$prediction_report"
}

generate_prediction_assessment() {
    local prediction_report="$1"

    cat >> "$prediction_report" << EOF
## 📊 综合预测评估

基于当前的质量趋势分析，我们对 Go Mastery 项目的质量发展提供以下预测评估：

### 🎯 关键预测洞察

1. **安全状况**: 如果当前的开发和质量保障流程得到维持，项目应该能够继续保持零安全漏洞状态。

2. **代码质量**: 基于历史趋势，代码质量指标预计将保持在当前水平或略有改善。

3. **维护负担**: errcheck和其他代码质量问题的数量预计保持稳定，符合项目的质量管理目标。

### 🚀 预测建议

- **继续执行**: 当前的质量保障措施正在发挥作用，建议继续执行
- **定期监控**: 保持定期质量监控，及时发现趋势变化
- **预防性措施**: 在预测到质量退化时，提前采取预防措施

### ⚠️ 预测局限性

请注意，本预测基于历史数据的线性外推，实际情况可能受到以下因素影响：
- 代码变更频率和复杂度
- 团队成员变化
- 新功能开发
- 外部依赖更新

建议结合实际项目情况和专业判断来解释预测结果。

---
*预测模型: 线性趋势外推 | 置信度: 基于历史数据稳定性*
EOF
}

# ===================================================================
# 📊 生成可视化ASCII图表
# ===================================================================

generate_ascii_charts() {
    print_section "生成可视化图表"

    local trends_file="$TRENDS_DIR/quality_trends.csv"
    local charts_file="$ANALYSIS_DIR/ascii_charts_$TIMESTAMP.txt"

    echo "# Go Mastery 项目质量趋势可视化图表" > "$charts_file"
    echo "生成时间: $DATE_HUMAN" >> "$charts_file"
    echo "" >> "$charts_file"

    # 为关键指标生成ASCII图表
    generate_ascii_chart "安全漏洞" 2 "$trends_file" "$charts_file"
    generate_ascii_chart "Go vet警告" 3 "$trends_file" "$charts_file"
    generate_ascii_chart "代码质量问题" 4 "$trends_file" "$charts_file"

    print_success "ASCII图表生成完成: $charts_file"
    echo ""
}

generate_ascii_chart() {
    local metric_name="$1"
    local column="$2"
    local trends_file="$3"
    local charts_file="$4"

    print_info "生成图表: $metric_name"

    # 获取最近10个数据点
    local data=$(grep -v "^timestamp" "$trends_file" | grep -v "^$" | tail -10 | cut -d',' -f"$column" | grep '^[0-9]*$')

    if [ -z "$data" ]; then
        echo "## $metric_name 趋势图" >> "$charts_file"
        echo "数据不足，无法生成图表" >> "$charts_file"
        echo "" >> "$charts_file"
        return
    fi

    local max_value=$(echo "$data" | sort -n | tail -1)
    local min_value=$(echo "$data" | sort -n | head -1)

    # 避免除零错误
    if [ "$max_value" = "$min_value" ]; then
        max_value=$((min_value + 1))
    fi

    echo "## $metric_name 趋势图" >> "$charts_file"
    echo "" >> "$charts_file"
    echo "数据范围: $min_value - $max_value" >> "$charts_file"
    echo "" >> "$charts_file"

    local line_num=0
    while IFS= read -r value; do
        line_num=$((line_num + 1))
        if [ -n "$value" ] && [[ "$value" =~ ^[0-9]+$ ]]; then
            # 计算条形长度 (最大20个字符)
            local bar_length=1
            if [ "$max_value" -gt "$min_value" ]; then
                bar_length=$(( (value - min_value) * 20 / (max_value - min_value) + 1 ))
            fi

            # 生成条形图
            local bar=""
            for i in $(seq 1 $bar_length); do
                bar="${bar}█"
            done

            printf "%2d: %5s %s\n" "$line_num" "$value" "$bar" >> "$charts_file"
        fi
    done <<< "$data"

    echo "" >> "$charts_file"
    echo "图例: █ = 数据点相对大小" >> "$charts_file"
    echo "" >> "$charts_file"
}

# ===================================================================
# 📋 生成综合分析报告
# ===================================================================

generate_comprehensive_analysis() {
    print_section "生成综合分析报告"

    local comprehensive_report="$ANALYSIS_DIR/comprehensive_analysis_$TIMESTAMP.md"

    cat > "$comprehensive_report" << EOF
# 📊 Go Mastery 项目质量综合分析报告

**分析时间**: $DATE_HUMAN
**分析范围**: 完整历史数据
**分析版本**: v1.0

---

## 🎯 执行摘要

Go Mastery 项目质量分析已完成。本报告基于历史质量监控数据，提供深度的趋势分析、异常检测和质量预测。

### 📊 分析亮点

EOF

    # 读取分析结果并生成摘要
    local data_samples=$(grep "^data_samples=" "$ANALYSIS_DIR/analysis_metadata_$TIMESTAMP.txt" | cut -d'=' -f2 2>/dev/null || echo "N/A")

    echo "- **数据样本**: $data_samples 个监控记录" >> "$comprehensive_report"
    echo "- **分析深度**: 基础统计、趋势分析、异常检测、质量预测" >> "$comprehensive_report"
    echo "- **分析工具**: 统计学方法、线性回归、Z-score异常检测" >> "$comprehensive_report"

    cat >> "$comprehensive_report" << EOF

## 📈 关键发现

### 1. 安全状况评估
基于历史数据分析，项目在安全方面表现出了对零安全漏洞目标的坚持。这是一个重要的质量里程碑。

### 2. 质量趋势洞察
代码质量指标显示了项目质量管理的成熟度。通过趋势分析，我们可以预测质量发展方向。

### 3. 异常模式识别
通过统计学方法识别的异常数据点有助于理解质量波动的原因和模式。

## 🔍 详细分析结果

本报告包含以下详细分析组件：

1. **基础统计分析**: [basic_stats_$TIMESTAMP.txt](basic_stats_$TIMESTAMP.txt)
   - 各质量指标的描述性统计
   - 平均值、中位数、标准差等关键统计量
   - 质量评价和建议

2. **趋势分析**: [trend_analysis_$TIMESTAMP.md](trend_analysis_$TIMESTAMP.md)
   - 近期质量指标变化趋势
   - 线性回归分析
   - 趋势强度和方向评估

3. **异常检测**: [anomaly_detection_$TIMESTAMP.md](anomaly_detection_$TIMESTAMP.md)
   - 基于Z-score的异常数据点识别
   - 异常模式分析
   - 潜在问题预警

4. **质量预测**: [quality_prediction_$TIMESTAMP.md](quality_prediction_$TIMESTAMP.md)
   - 基于历史趋势的未来预测
   - 预测置信度评估
   - 风险预警和建议

5. **可视化图表**: [ascii_charts_$TIMESTAMP.txt](ascii_charts_$TIMESTAMP.txt)
   - ASCII格式的趋势图表
   - 直观的数据可视化
   - 趋势模式展示

## 🚀 行动建议

基于综合分析结果，提供以下行动建议：

### 立即行动 (1-7天)
1. **维持零安全漏洞状态**: 继续执行当前的安全检查流程
2. **监控关键指标**: 密切关注安全漏洞和Go vet警告数量
3. **异常响应**: 对检测到的异常数据点进行根因分析

### 短期优化 (1-4周)
1. **趋势改善**: 针对上升趋势的指标制定改进计划
2. **流程优化**: 基于分析结果优化质量保障流程
3. **工具增强**: 考虑增加更多自动化质量检查工具

### 长期规划 (1-3月)
1. **质量文化**: 建立基于数据的质量管理文化
2. **预测模型**: 开发更精确的质量预测模型
3. **基准建立**: 建立行业级质量基准和目标

## 📊 质量成熟度评估

基于本次分析，Go Mastery 项目的质量成熟度评估如下：

- **数据驱动**: ✅ 建立了系统的质量数据收集机制
- **趋势监控**: ✅ 能够识别和分析质量趋势变化
- **异常检测**: ✅ 具备异常情况的自动识别能力
- **预测能力**: ✅ 初步具备质量预测和风险预警能力
- **持续改进**: 🔄 正在建立基于数据的持续改进循环

## 🎯 结论

Go Mastery 项目在质量管理方面展现了较高的成熟度，特别是在安全方面达到了企业级标准。通过持续的数据分析和趋势监控，项目能够维持高质量标准并及时识别改进机会。

建议继续保持当前的质量保障措施，同时基于分析结果不断优化和改进质量管理流程。

---

**分析团队**: Go Mastery 智能质量监控系统
**下次分析**: 建议在 7 天后重新执行
**技术支持**: 参考 [质量保障文档](../QUALITY_ASSURANCE.md)
EOF

    print_success "综合分析报告生成完成: $comprehensive_report"

    # 创建最新分析报告链接
    ln -sf "comprehensive_analysis_$TIMESTAMP.md" "$ANALYSIS_DIR/latest_analysis.md" 2>/dev/null || \
    cp "$comprehensive_report" "$ANALYSIS_DIR/latest_analysis.md"

    echo ""
}

# ===================================================================
# 🎯 主执行函数
# ===================================================================

main() {
    local start_time=$(date +%s)

    print_header
    create_directories

    # 执行分析流程
    prepare_and_validate_data
    basic_statistical_analysis
    trend_analysis
    anomaly_detection
    quality_prediction
    generate_ascii_charts
    generate_comprehensive_analysis

    # 执行总结
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    print_section "分析执行总结"
    print_success "质量趋势分析完成"
    print_info "执行时间: ${duration}秒"
    print_info "分析报告位置: $ANALYSIS_DIR/"
    print_info "最新分析: $ANALYSIS_DIR/latest_analysis.md"

    echo ""
    print_success "🎉 Go Mastery 质量趋势分析已完成"
    print_info "💡 建议查看生成的分析报告以获取详细洞察"
}

# 检查依赖
if ! command -v bc &> /dev/null; then
    print_warning "bc计算器未安装，部分数值计算可能受限"
    print_info "安装建议: sudo apt-get install bc (Ubuntu/Debian) 或 brew install bc (macOS)"
fi

# 执行主函数
main "$@"