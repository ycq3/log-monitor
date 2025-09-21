# 日志哨兵 (Log Monitor)

一个用Go语言实现的日志监控工具，能够实时监控指定的日志文件，当检测到错误关键词时自动发送告警消息到飞书或钉钉机器人。

## 功能特性

- 🔍 **实时监控**: 使用文件系统事件监控，实时检测日志文件变化
- 📁 **目录监控**: 支持监控整个目录，自动发现新增日志文件
- 🔄 **递归监控**: 支持递归监控子目录，灵活配置监控范围
- 🎯 **关键词过滤**: 支持自定义错误关键词，精确匹配告警内容
- 📝 **格式过滤**: 支持指定文件扩展名，只监控特定格式的日志文件
- 📱 **多平台通知**: 支持飞书和钉钉机器人消息推送
- ⚙️ **YAML配置**: 简单易用的YAML配置文件
- 🚀 **轻量高效**: 低资源占用，高性能监控
- 🔒 **安全支持**: 支持钉钉机器人签名验证
- 💾 **内存优化**: 智能内存管理，支持大规模文件监控
- 🧹 **自动清理**: 定期清理无效文件记录，保持系统稳定

## 快速开始

### 1. 编译程序

```bash
go mod tidy
go build -o log-monitor
```

### 2. 配置文件

复制并修改 `config.yaml` 文件：

```yaml
# 监控的日志文件列表
log_files:
  - path: "/var/log/app/application.log"
    keywords:
      - "ERROR"
      - "FATAL"
      - "Exception"
    enabled: true

# 通知器配置
notifiers:
  # 飞书机器人
  - type: "feishu"
    webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
    enabled: true
```

## 使用示例

### 基本文件监控

```bash
# 使用默认配置文件监控单个文件
./log-monitor

# 指定配置文件
./log-monitor -config /path/to/your/config.yaml
```

### 目录监控示例

创建测试目录和配置：

```bash
# 创建测试目录结构
mkdir -p test-logs/app1 test-logs/app2 test-logs/backup

# 创建测试配置文件
cat > test-directory-config.yaml << EOF
log_directories:
  - path: "./test-logs"
    keywords: ["ERROR", "FATAL"]
    extensions: [".log", ".txt"]
    recursive: true
    exclude_dirs: ["backup"]
    enabled: true

notifiers:
  - type: "feishu"
    webhook: "your-webhook-url"
    enabled: true
EOF

# 运行监控
./log-monitor -config test-directory-config.yaml
```

### 测试告警

在另一个终端中创建测试日志：

```bash
# 创建会触发告警的日志
echo "$(date) ERROR: This is a test error message" >> test-logs/app1/test.log

# 创建不会触发告警的日志
echo "$(date) INFO: This is an info message" >> test-logs/app1/test.log

# 在排除目录中创建日志（不会被监控）
echo "$(date) ERROR: This error will be ignored" >> test-logs/backup/backup.log
```

## 配置说明

### 日志文件配置

```yaml
log_files:
  - path: "/path/to/logfile.log"    # 日志文件路径
    keywords:                        # 错误关键词列表
      - "ERROR"
      - "FATAL"
      - "Exception"
    enabled: true                    # 是否启用监控
```

### 日志目录配置（新功能）

```yaml
log_directories:
  - path: "/var/log/apps"           # 监控目录路径
    keywords:                        # 错误关键词列表
      - "ERROR"
      - "FATAL"
      - "Exception"
    extensions: [".log", ".txt"]     # 监控的文件扩展名
    recursive: true                  # 是否递归监控子目录
    exclude_dirs: ["backup", "temp"] # 排除的子目录
    enabled: true                    # 是否启用监控
```

#### 目录监控参数说明

- `path`: 要监控的目录路径（绝对路径）
- `keywords`: 触发告警的关键词列表
- `extensions`: 要监控的文件扩展名，如 `[".log", ".txt", ".out"]`
- `recursive`: 是否递归监控子目录
  - `true`: 监控所有子目录
  - `false`: 只监控指定目录，不包含子目录
- `exclude_dirs`: 要排除的子目录名称列表
- `enabled`: 是否启用此目录监控

### 通知器配置

#### 飞书机器人

```yaml
notifiers:
  - type: "feishu"
    webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url"
    enabled: true
```

#### 钉钉机器人（带签名）

```yaml
notifiers:
  - type: "dingtalk"
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=your-access-token"
    secret: "your-secret-key"        # 可选，用于签名验证
    enabled: true
```

#### 钉钉机器人（不带签名）

```yaml
notifiers:
  - type: "dingtalk"
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=your-access-token"
    enabled: true
```

## 机器人配置指南

### 飞书机器人

1. 在飞书群聊中添加机器人
2. 选择"自定义机器人"
3. 复制Webhook地址到配置文件

### 钉钉机器人

1. 在钉钉群聊中添加机器人
2. 选择"自定义机器人"
3. 设置安全设置（推荐使用加签方式）
4. 复制Webhook地址和密钥到配置文件

## 告警消息格式

当检测到错误时，会发送如下格式的消息：

```
🚨 日志告警

文件: /var/log/app/application.log
时间: 2024-01-15 14:30:25
内容: ERROR: Database connection failed
```

## 注意事项

1. **文件权限**: 确保程序有读取日志文件和目录的权限
2. **文件路径**: 使用绝对路径指定日志文件和目录
3. **关键词匹配**: 关键词匹配不区分大小写
4. **网络连接**: 确保服务器能访问飞书/钉钉的API
5. **资源占用**: 监控大量文件时注意系统资源使用情况
6. **文件大小限制**: 默认限制单个文件最大100MB，超过限制的文件会被跳过
7. **内存管理**: 程序会定期清理无效文件记录，每30分钟执行一次
8. **目录监控**: 
   - 递归监控会监控所有子目录，请合理设置排除目录
   - 新创建的文件会自动被监控
   - 删除的文件会自动从监控列表中移除

## 系统要求

- Go 1.21 或更高版本
- Linux/macOS/Windows 操作系统
- 网络连接（用于发送通知）

## 依赖包

- `github.com/fsnotify/fsnotify`: 文件系统事件监控
- `gopkg.in/yaml.v3`: YAML配置文件解析

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request来改进这个项目！