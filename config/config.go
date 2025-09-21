package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// Config 主配置结构
type Config struct {
	LogFiles       []LogFile      `yaml:"log_files"`
	LogDirectories []LogDirectory `yaml:"log_directories"`
	Notifiers      []Notifier     `yaml:"notifiers"`
}

// LogFile 日志文件配置
type LogFile struct {
	Path     string   `yaml:"path"`
	Keywords []string `yaml:"keywords"`
	Enabled  bool     `yaml:"enabled"`
}

// LogDirectory 日志目录配置
type LogDirectory struct {
	Path        string   `yaml:"path"`
	Keywords    []string `yaml:"keywords"`
	Extensions  []string `yaml:"extensions"`  // 支持的文件扩展名，如 [".log", ".txt"]
	Recursive   bool     `yaml:"recursive"`   // 是否递归监控子目录
	ExcludeDirs []string `yaml:"exclude_dirs,omitempty"` // 排除的子目录
	Enabled     bool     `yaml:"enabled"`
}

// Notifier 通知器配置
type Notifier struct {
	Type    string `yaml:"type"`    // "feishu" 或 "dingtalk"
	Webhook string `yaml:"webhook"`
	Secret  string `yaml:"secret,omitempty"`
	Enabled bool   `yaml:"enabled"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.LogFiles) == 0 && len(c.LogDirectories) == 0 {
		return fmt.Errorf("至少需要配置一个日志文件或日志目录")
	}

	if len(c.Notifiers) == 0 {
		return fmt.Errorf("至少需要配置一个通知器")
	}

	for i, logFile := range c.LogFiles {
		if logFile.Path == "" {
			return fmt.Errorf("日志文件[%d]路径不能为空", i)
		}
		if len(logFile.Keywords) == 0 {
			return fmt.Errorf("日志文件[%d]关键词不能为空", i)
		}
	}

	for i, logDir := range c.LogDirectories {
		if logDir.Path == "" {
			return fmt.Errorf("日志目录[%d]路径不能为空", i)
		}
		if len(logDir.Keywords) == 0 {
			return fmt.Errorf("日志目录[%d]关键词不能为空", i)
		}
		if len(logDir.Extensions) == 0 {
			return fmt.Errorf("日志目录[%d]必须指定至少一个文件扩展名", i)
		}
	}

	for i, notifier := range c.Notifiers {
		if notifier.Type != "feishu" && notifier.Type != "dingtalk" {
			return fmt.Errorf("通知器[%d]类型必须是 feishu 或 dingtalk", i)
		}
		if notifier.Webhook == "" {
			return fmt.Errorf("通知器[%d]webhook不能为空", i)
		}
	}

	return nil
}