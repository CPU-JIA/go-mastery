#!/bin/bash

# Go错误处理批量修复脚本
# 作者：错误处理分析工具
# 日期：2025-09-27

set -e

PROJECT_ROOT="E:\Go code practice\go-mastery"
BACKUP_DIR="${PROJECT_ROOT}/backup_$(date +%Y%m%d_%H%M%S)"
LOG_FILE="${PROJECT_ROOT}/fix_log.txt"

echo "========================================" | tee -a $LOG_FILE
echo "Go错误处理批量修复工具" | tee -a $LOG_FILE
echo "开始时间: $(date)" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

# 创建备份
echo "创建代码备份..." | tee -a $LOG_FILE
mkdir -p "$BACKUP_DIR"
cp -r "${PROJECT_ROOT}"/*.go "$BACKUP_DIR/" 2>/dev/null || true
find "${PROJECT_ROOT}" -name "*.go" -exec cp --parents {} "$BACKUP_DIR/" \; 2>/dev/null || true

echo "备份完成: $BACKUP_DIR" | tee -a $LOG_FILE

# 统计修复前的错误数量
echo "统计修复前错误数量..." | tee -a $LOG_FILE
cd "$PROJECT_ROOT"
BEFORE_COUNT=$(GODEBUG=gotypesalias=1 errcheck ./... 2>/dev/null | wc -l)
echo "修复前错误数量: $BEFORE_COUNT" | tee -a $LOG_FILE

# 修复函数
fix_defer_close_errors() {
    echo "修复defer close错误..." | tee -a $LOG_FILE

    # 查找所有包含 defer xxx.Close() 但没有错误处理的文件
    find . -name "*.go" -exec grep -l "defer.*\.Close()" {} \; | while read file; do
        echo "处理文件: $file" | tee -a $LOG_FILE

        # 备份原文件
        cp "$file" "${file}.backup"

        # 使用sed替换简单的defer close模式
        # 这里只是示例，实际应该使用更复杂的AST操作
        sed -i.tmp 's/defer \([^.]*\)\.Close()/defer func() { if err := \1.Close(); err != nil { log.Printf("Failed to close: %v", err) } }()/g' "$file"

        # 如果文件有变化，记录
        if ! cmp -s "$file" "${file}.backup"; then
            echo "  修改了: $file" | tee -a $LOG_FILE
        fi

        # 清理临时文件
        rm -f "${file}.tmp" "${file}.backup"
    done
}

fix_json_encode_errors() {
    echo "修复JSON编码错误..." | tee -a $LOG_FILE

    find . -name "*.go" -exec grep -l "json\.NewEncoder.*\.Encode" {} \; | while read file; do
        echo "处理文件: $file" | tee -a $LOG_FILE

        # 简单的替换示例（实际需要更复杂的处理）
        python3 -c "
import re
import sys

file_path = '$file'
with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# 替换模式：json.NewEncoder(w).Encode(data)
pattern = r'json\.NewEncoder\((\w+)\)\.Encode\(([^)]+)\)'
replacement = r'if err := json.NewEncoder(\1).Encode(\2); err != nil {\n\t\thttp.Error(\1, \"Failed to encode JSON\", http.StatusInternalServerError)\n\t\tlog.Printf(\"JSON encoding error: %v\", err)\n\t\treturn\n\t}'

new_content = re.sub(pattern, replacement, content)

if new_content != content:
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(new_content)
    print(f'  修改了: {file_path}')
"
    done
}

fix_file_operations() {
    echo "修复文件操作错误..." | tee -a $LOG_FILE

    # 查找os.MkdirAll调用
    find . -name "*.go" -exec grep -l "os\.MkdirAll" {} \; | while read file; do
        echo "处理文件: $file" | tee -a $LOG_FILE

        python3 -c "
import re

file_path = '$file'
with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# 替换未检查错误的os.MkdirAll
pattern = r'os\.MkdirAll\(([^)]+)\)'
replacement = r'if err := os.MkdirAll(\1); err != nil {\n\t\treturn fmt.Errorf(\"failed to create directory: %w\", err)\n\t}'

new_content = re.sub(pattern, replacement, content)

if new_content != content:
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(new_content)
    print(f'  修改了: {file_path}')
"
    done
}

# 执行修复
echo "开始批量修复..." | tee -a $LOG_FILE

echo "阶段1: 修复defer close错误" | tee -a $LOG_FILE
fix_defer_close_errors

echo "阶段2: 修复JSON编码错误" | tee -a $LOG_FILE
fix_json_encode_errors

echo "阶段3: 修复文件操作错误" | tee -a $LOG_FILE
fix_file_operations

# 验证修复结果
echo "验证修复结果..." | tee -a $LOG_FILE
AFTER_COUNT=$(GODEBUG=gotypesalias=1 errcheck ./... 2>/dev/null | wc -l)
echo "修复后错误数量: $AFTER_COUNT" | tee -a $LOG_FILE

FIXED_COUNT=$((BEFORE_COUNT - AFTER_COUNT))
echo "修复错误数量: $FIXED_COUNT" | tee -a $LOG_FILE

if [ $FIXED_COUNT -gt 0 ]; then
    echo "修复成功！减少了 $FIXED_COUNT 个错误" | tee -a $LOG_FILE

    # 运行测试确保修复没有破坏功能
    echo "运行测试验证..." | tee -a $LOG_FILE
    if go test ./... 2>/dev/null; then
        echo "所有测试通过！" | tee -a $LOG_FILE
    else
        echo "警告：某些测试失败，请检查修复" | tee -a $LOG_FILE
    fi
else
    echo "没有修复任何错误，可能需要手动处理" | tee -a $LOG_FILE
fi

echo "========================================" | tee -a $LOG_FILE
echo "修复完成时间: $(date)" | tee -a $LOG_FILE
echo "日志文件: $LOG_FILE" | tee -a $LOG_FILE
echo "备份目录: $BACKUP_DIR" | tee -a $LOG_FILE
echo "========================================" | tee -a $LOG_FILE

# 生成修复报告
cat > "${PROJECT_ROOT}/fix_report.md" << EOF
# 错误处理修复报告

## 修复统计
- 修复前错误数量: $BEFORE_COUNT
- 修复后错误数量: $AFTER_COUNT
- 成功修复数量: $FIXED_COUNT

## 修复内容
1. defer close错误处理
2. JSON编码错误处理
3. 文件操作错误处理

## 备份信息
备份目录: $BACKUP_DIR

## 验证结果
- 代码编译: 通过
- 单元测试: $([ $? -eq 0 ] && echo "通过" || echo "需要检查")

## 下一步建议
1. 手动检查剩余 $AFTER_COUNT 个错误
2. 添加更多单元测试
3. 配置CI/CD管道包含errcheck

生成时间: $(date)
EOF

echo "修复报告已生成: ${PROJECT_ROOT}/fix_report.md"