# 具体修复示例：04-web/01-http-basics/main.go

## 问题分析

该文件有38个错误处理遗漏，主要集中在：
1. `fmt.Fprintf` 调用未检查返回的错误
2. `fmt.Fprint` 调用未检查返回的错误
3. `defer resp.Body.Close()` 未检查关闭错误
4. `reader.ReadString()` 未检查读取错误

## 修复前后对比

### 修复前（错误代码）：
```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.Header().Set("X-Custom-Header", "Go-HTTP-Server")

    // 未检查错误
    fmt.Fprintf(w, "Hello, World!\n")
    fmt.Fprintf(w, "Method: %s\n", r.Method)
    fmt.Fprintf(w, "URL Path: %s\n", r.URL.Path)
    fmt.Fprintf(w, "Remote Address: %s\n", r.RemoteAddr)
    fmt.Fprintf(w, "User-Agent: %s\n", r.Header.Get("User-Agent"))
}
```

### 修复后（正确代码）：
```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.Header().Set("X-Custom-Header", "Go-HTTP-Server")

    // 检查写入错误的辅助函数
    writeOrLog := func(format string, args ...interface{}) {
        if _, err := fmt.Fprintf(w, format, args...); err != nil {
            log.Printf("Error writing response: %v", err)
        }
    }

    writeOrLog("Hello, World!\n")
    writeOrLog("Method: %s\n", r.Method)
    writeOrLog("URL Path: %s\n", r.URL.Path)
    writeOrLog("Remote Address: %s\n", r.RemoteAddr)
    writeOrLog("User-Agent: %s\n", r.Header.Get("User-Agent"))
}
```

### 资源关闭错误修复：

**修复前：**
```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()  // 未检查关闭错误
```

**修复后：**
```go
resp, err := http.Get(url)
if err != nil {
    return err
}
defer func() {
    if closeErr := resp.Body.Close(); closeErr != nil {
        log.Printf("Failed to close response body: %v", closeErr)
    }
}()
```

### 文件读取错误修复：

**修复前：**
```go
reader := bufio.NewReader(os.Stdin)
input, _ := reader.ReadString('\n')  // 忽略错误
```

**修复后：**
```go
reader := bufio.NewReader(os.Stdin)
input, err := reader.ReadString('\n')
if err != nil {
    log.Printf("Error reading input: %v", err)
    return
}
```

## 批量修复策略

### 1. 创建辅助函数
```go
// ResponseWriter 错误处理辅助函数
func writeResponse(w http.ResponseWriter, format string, args ...interface{}) {
    if _, err := fmt.Fprintf(w, format, args...); err != nil {
        log.Printf("Error writing HTTP response: %v", err)
    }
}

func writeJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Error encoding JSON response: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
```

### 2. 使用Go template进行批量替换
```bash
# 查找所有fmt.Fprintf调用
grep -rn "fmt\.Fprintf(w," . --include="*.go"

# 使用sed进行简单替换（示例）
find . -name "*.go" -exec sed -i 's/fmt\.Fprintf(w,/writeResponse(w,/g' {} \;
```

### 3. 使用gofmt和goimports清理代码
```bash
gofmt -w .
goimports -w .
```

## 验证修复效果

### 修复前errcheck输出：
```
04-web\01-http-basics\main.go:67:13:	fmt.Fprintf(w, "Hello, World!\n")
04-web\01-http-basics\main.go:68:13:	fmt.Fprintf(w, "Method: %s\n", r.Method)
04-web\01-http-basics\main.go:69:13:	fmt.Fprintf(w, "URL Path: %s\n", r.URL.Path)
04-web\01-http-basics\main.go:70:13:	fmt.Fprintf(w, "Remote Address: %s\n", r.RemoteAddr)
04-web\01-http-basics\main.go:71:13:	fmt.Fprintf(w, "User-Agent: %s\n", r.Header.Get("User-Agent"))
... (38个错误)
```

### 修复后应该没有输出（或大幅减少）

## 注意事项

1. **HTTP响应写入错误通常不需要终止请求**：对于HTTP响应的fmt.Fprintf错误，通常记录日志即可，不需要返回错误
2. **资源关闭错误必须处理**：文件、连接等资源的关闭错误必须妥善处理
3. **保持代码简洁**：避免过度的错误检查影响代码可读性
4. **测试验证**：修复后运行测试确保功能正常

## 自动化脚本示例

```bash
#!/bin/bash
# 针对04-web/01-http-basics/main.go的修复脚本

FILE="04-web/01-http-basics/main.go"
BACKUP="${FILE}.backup"

# 创建备份
cp "$FILE" "$BACKUP"

# 添加辅助函数（需要手动添加到文件开头）
cat >> fix_helpers.go << 'EOF'
package main

import "log"

func writeResponse(w http.ResponseWriter, format string, args ...interface{}) {
    if _, err := fmt.Fprintf(w, format, args...); err != nil {
        log.Printf("Error writing HTTP response: %v", err)
    }
}
EOF

# 替换fmt.Fprintf调用
sed -i 's/fmt\.Fprintf(w,/writeResponse(w,/g' "$FILE"

echo "修复完成，原文件备份为: $BACKUP"
```