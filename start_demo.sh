#!/bin/bash

# 日志哨兵演示脚本
# 此脚本将启动日志监控和测试日志生成器进行演示

echo "=== 日志哨兵演示 ==="
echo

# 检查是否已编译
if [ ! -f "log-monitor" ]; then
    echo "正在编译日志哨兵..."
    make build
    echo
fi

if [ ! -f "test-generator" ]; then
    echo "正在编译测试工具..."
    make build-test
    echo
fi

# 清理旧的测试日志
if [ -f "test.log" ]; then
    echo "清理旧的测试日志..."
    rm test.log
    echo
fi

echo "=== 使用说明 ==="
echo "1. 此演示将使用 test-config.yaml 配置文件"
echo "2. 监控文件: test.log"
echo "3. 关键词: ERROR, FATAL, Exception, panic, failed"
echo "4. 通知器已禁用，避免发送真实消息"
echo
echo "=== 开始演示 ==="
echo

# 在后台启动日志监控
echo "启动日志哨兵（后台运行）..."
./log-monitor -config test-config.yaml &
MONITOR_PID=$!
echo "日志哨兵 PID: $MONITOR_PID"
echo

# 等待监控器启动
sleep 2

echo "启动测试日志生成器..."
echo "（将每2秒生成一条日志，包含各种级别的消息）"
echo "按 Ctrl+C 停止演示"
echo

# 启动测试日志生成器
./test-generator &
GENERATOR_PID=$!

# 等待用户中断
trap 'echo; echo "正在停止演示..."; kill $MONITOR_PID $GENERATOR_PID 2>/dev/null; echo "演示已停止"; exit 0' INT

# 保持脚本运行
wait