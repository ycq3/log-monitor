package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-monitor/config"
	"log-monitor/monitor"
	"log-monitor/notifier"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	log.Printf("配置加载成功，监控 %d 个日志文件，配置 %d 个通知器", len(cfg.LogFiles), len(cfg.Notifiers))

	// 创建通知器
	notifiers := notifier.CreateNotifiers(cfg.Notifiers)
	if len(notifiers) == 0 {
		log.Fatalf("没有可用的通知器")
	}

	log.Printf("创建了 %d 个通知器", len(notifiers))

	// 创建日志监控器
	logMonitor, err := monitor.NewLogMonitor(cfg, notifiers)
	if err != nil {
		log.Fatalf("创建日志监控器失败: %v", err)
	}

	// 启动监控
	if err := logMonitor.Start(); err != nil {
		log.Fatalf("启动监控失败: %v", err)
	}

	log.Println("日志哨兵启动成功，开始监控...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-sigChan

	log.Println("收到退出信号，正在关闭...")

	// 停止监控
	if err := logMonitor.Stop(); err != nil {
		log.Printf("停止监控失败: %v", err)
	}

	log.Println("日志哨兵已关闭")
}