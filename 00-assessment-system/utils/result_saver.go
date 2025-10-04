/*
=== Go语言学习评估系统 - 通用结果保存工具 ===

本文件提供评估结果保存的通用功能：
1. JSON结果序列化和安全保存
2. 统一的文件权限管理（使用common/security包）
3. 错误处理和日志记录
4. 支持自动创建父目录

作者: JIA
创建时间: 2025-10-03
版本: 1.0.0
*/

// Package utils 提供评估系统的通用工具函数
//
// 本包包含了各种辅助工具，如结果保存、数据转换等，
// 为评估系统的其他模块提供基础支持。
package utils

import (
	"encoding/json"
	"fmt"
	"go-mastery/common/security"
	"log"
	"path/filepath"
)

// SaveResultOptions 保存结果的选项配置
type SaveResultOptions struct {
	// DefaultPath 当未指定路径时使用的默认路径
	DefaultPath string
	// LogMessage 保存成功后输出的日志消息模板（使用%s作为路径占位符）
	LogMessage string
	// CreateDir 是否自动创建父目录
	CreateDir bool
}

// SaveJSONResult 安全地将结果保存为JSON文件
//
// 功能说明:
//   - 自动序列化结果为格式化的JSON（缩进2空格）
//   - 使用security包进行安全的文件写入（防止G301/G304漏洞）
//   - 支持自动创建父目录
//   - 提供详细的错误信息
//
// 参数:
//   - filePath: 目标文件路径（为空时使用options.DefaultPath）
//   - result: 要保存的结果对象（必须可JSON序列化）
//   - options: 保存选项配置
//
// 返回:
//   - error: 操作错误（序列化失败或文件写入失败）
//
// 示例:
//
//	err := SaveJSONResult("results.json", myResult, &SaveResultOptions{
//	    DefaultPath: "default_results.json",
//	    LogMessage:  "评估结果已保存到: %s",
//	    CreateDir:   true,
//	})
//
// 作者: JIA
func SaveJSONResult(filePath string, result interface{}, options *SaveResultOptions) error {
	// 1. 参数验证和默认值处理
	if options == nil {
		options = &SaveResultOptions{
			DefaultPath: "result.json",
			LogMessage:  "结果已保存到: %s",
			CreateDir:   true,
		}
	}

	// 如果未指定路径，使用默认路径
	if filePath == "" {
		filePath = options.DefaultPath
	}

	// 确保文件路径是绝对路径或规范化的相对路径
	filePath = filepath.Clean(filePath)

	// 2. 序列化结果为JSON格式
	//    - 使用缩进格式化，便于人类阅读
	//    - 每级缩进2个空格（Go社区标准）
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}

	// 3. 使用security包安全写入文件
	//    - 使用推荐的data文件权限（0600，仅所有者可读写）
	//    - 防止G301（不安全的文件权限）和G304（路径遍历）漏洞
	//    - 如果指定CreateDir，自动创建父目录
	if err := security.SecureWriteFile(filePath, data, &security.SecureFileOptions{
		Mode:      security.GetRecommendedMode("data"),
		CreateDir: options.CreateDir,
	}); err != nil {
		return fmt.Errorf("保存结果文件失败: %w", err)
	}

	// 4. 记录成功日志
	if options.LogMessage != "" {
		log.Printf(options.LogMessage, filePath)
	}

	return nil
}

// SaveCodeQualityResult 专门用于保存代码质量评估结果
//
// 这是SaveJSONResult的封装，提供了代码质量评估专用的默认配置
//
// 作者: JIA
func SaveCodeQualityResult(filePath string, result interface{}) error {
	return SaveJSONResult(filePath, result, &SaveResultOptions{
		DefaultPath: "quality_assessment_results.json",
		LogMessage:  "代码质量评估结果已保存到: %s",
		CreateDir:   true,
	})
}

// SaveProjectEvalResult 专门用于保存项目评估结果
//
// 这是SaveJSONResult的封装，提供了项目评估专用的默认配置
//
// 作者: JIA
func SaveProjectEvalResult(filePath string, result interface{}) error {
	return SaveJSONResult(filePath, result, &SaveResultOptions{
		DefaultPath: "project_evaluation_results.json",
		LogMessage:  "项目评估结果已保存到: %s",
		CreateDir:   true,
	})
}
