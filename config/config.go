package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// Config 主配置结构
type Config struct {
	LogFiles []LogFile `yaml:"log_files"`
	Notifiers []Notifier `yaml:"notifiers"`
}

// LogFile 日志文件配置
type LogFile struct {
	Path     string   `yaml:"path"`
	Keywords []string `yaml:"keywords"`
	Enabled  bool     `yaml:"enabled"`
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
	if len(c.LogFiles) == 0 {
		return fmt.Errorf("至少需要配置一个日志文件")
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