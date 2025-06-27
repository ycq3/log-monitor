# Makefile for Log Monitor

.PHONY: build clean test run help

# 默认目标
all: build

# 编译主程序
build:
	@echo "编译日志哨兵..."
	go build -o log-monitor
	@echo "编译完成: log-monitor"

# 编译测试工具
build-test:
	@echo "编译测试日志生成器..."
	go build -o test-generator ./cmd/test_log_generator.go
	@echo "编译完成: test-generator"

# 编译所有
build-all: build build-test

# 运行主程序（使用默认配置）
run:
	./log-monitor

# 运行主程序（使用测试配置）
run-test:
	./log-monitor -config test-config.yaml

# 生成测试日志
generate-test-log:
	./test-generator

# 清理编译文件
clean:
	@echo "清理编译文件..."
	rm -f log-monitor test-generator test.log
	@echo "清理完成"

# 下载依赖
deps:
	@echo "下载依赖包..."
	go mod tidy
	@echo "依赖下载完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "格式化完成"

# 运行测试
test:
	@echo "运行测试..."
	go test ./...
	@echo "测试完成"

# 显示帮助
help:
	@echo "可用的命令:"
	@echo "  build          - 编译主程序"
	@echo "  build-test     - 编译测试工具"
	@echo "  build-all      - 编译所有程序"
	@echo "  run            - 运行主程序（默认配置）"
	@echo "  run-test       - 运行主程序（测试配置）"
	@echo "  generate-test-log - 生成测试日志"
	@echo "  clean          - 清理编译文件"
	@echo "  deps           - 下载依赖包"
	@echo "  fmt            - 格式化代码"
	@echo "  test           - 运行测试"
	@echo "  help           - 显示此帮助信息"