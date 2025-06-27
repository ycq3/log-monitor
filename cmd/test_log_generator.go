package main

import (
	"fmt"
	"os"
	"time"
)

// 测试日志生成器
// 用于生成包含错误关键词的测试日志，验证监控功能
func main() {
	logFile := "test.log"
	
	// 创建或打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("创建日志文件失败: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Printf("开始生成测试日志到文件: %s\n", logFile)
	fmt.Println("按 Ctrl+C 停止生成")

	// 测试消息列表
	testMessages := []string{
		"INFO: Application started successfully",
		"DEBUG: Processing user request",
		"WARN: High memory usage detected",
		"ERROR: Database connection failed",
		"INFO: User login successful",
		"FATAL: System crash detected",
		"DEBUG: Cache hit ratio: 85%",
		"ERROR: Failed to process payment",
		"INFO: Backup completed successfully",
		"Exception: NullPointerException in UserService",
		"INFO: Server health check passed",
		"panic: runtime error: index out of range",
	}

	// 循环生成日志
	for i := 0; ; i++ {
		message := testMessages[i%len(testMessages)]
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
		
		// 写入日志
		_, err := file.WriteString(logLine)
		if err != nil {
			fmt.Printf("写入日志失败: %v\n", err)
			break
		}
		
		// 强制刷新到磁盘
		file.Sync()
		
		fmt.Printf("生成日志: %s", logLine)
		
		// 等待2秒
		time.Sleep(2 * time.Second)
	}
}