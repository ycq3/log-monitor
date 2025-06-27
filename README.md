# 日志哨兵 (Log Monitor)

一个用Go语言实现的日志监控工具，能够实时监控指定的日志文件，当检测到错误关键词时自动发送告警消息到飞书或钉钉机器人。

## 功能特性

- 🔍 **实时监控**: 使用文件系统事件监控，实时检测日志文件变化
- 🎯 **关键词过滤**: 支持自定义错误关键词，精确匹配告警内容
- 📱 **多平台通知**: 支持飞书和钉钉机器人消息推送
- ⚙️ **YAML配置**: 简单易用的YAML配置文件
- 🚀 **轻量高效**: 低资源占用，高性能监控
- 🔒 **安全支持**: 支持钉钉机器人签名验证

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

### 3. 运行程序

```bash
# 使用默认配置文件
./log-monitor

# 指定配置文件
./log-monitor -config /path/to/your/config.yaml
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

1. **文件权限**: 确保程序有读取日志文件的权限
2. **文件路径**: 使用绝对路径指定日志文件
3. **关键词匹配**: 关键词匹配不区分大小写
4. **网络连接**: 确保服务器能访问飞书/钉钉的API
5. **资源占用**: 监控大量文件时注意系统资源使用情况

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