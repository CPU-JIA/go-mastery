/*
Package sysutil 提供跨平台的系统编程工具集。

本包包含以下子模块：

  - process: 进程管理工具（创建、监控、信号处理）
  - network: 网络诊断和工具（连接池、端口扫描、网络信息）
  - resource: 资源管理工具（文件描述符、内存、CPU）
  - platform: 平台抽象层（Windows/Linux 统一接口）

设计原则：

 1. 跨平台兼容：所有工具在 Windows 和 Linux 上都能正常工作
 2. 生产级质量：包含完整的错误处理和资源清理
 3. 高性能：使用高效的系统调用和内存管理
 4. 易于使用：提供简洁的 API 和详细的文档

使用示例：

	import "go-mastery/09-system-programming/sysutil/process"

	// 获取系统进程列表
	procs, err := process.List()
	if err != nil {
	    log.Fatal(err)
	}

	for _, p := range procs {
	    fmt.Printf("PID: %d, Name: %s\n", p.PID, p.Name)
	}

注意事项：

  - 某些功能可能需要管理员/root 权限
  - 平台特定的功能会在文档中标注
  - 建议在生产环境中进行充分测试
*/
package sysutil
