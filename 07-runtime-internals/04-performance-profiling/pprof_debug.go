//go:build debug
// +build debug

package main

import _ "net/http/pprof"

// 本文件只在使用 -tags=debug 构建时才会包含pprof端点
// 使用方式：
// 开发环境: go run -tags=debug main.go pprof_debug.go
// 生产环境: go run main.go (不包含网络pprof端点，确保生产安全)