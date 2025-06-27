package notifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"log-monitor/config"
)

// Notifier 通知器接口
type Notifier interface {
	Send(message string) error
}

// CreateNotifiers 根据配置创建通知器
func CreateNotifiers(configs []config.Notifier) []Notifier {
	var notifiers []Notifier

	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		switch cfg.Type {
		case "feishu":
			notifiers = append(notifiers, NewFeishuNotifier(cfg.Webhook))
		case "dingtalk":
			notifiers = append(notifiers, NewDingtalkNotifier(cfg.Webhook, cfg.Secret))
		}
	}

	return notifiers
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	webhook string
}

// NewFeishuNotifier 创建飞书通知器
func NewFeishuNotifier(webhook string) *FeishuNotifier {
	return &FeishuNotifier{webhook: webhook}
}

// Send 发送飞书消息
func (f *FeishuNotifier) Send(message string) error {
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}

	return f.sendHTTPRequest(payload)
}

// sendHTTPRequest 发送HTTP请求
func (f *FeishuNotifier) sendHTTPRequest(payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	resp, err := http.Post(f.webhook, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// DingtalkNotifier 钉钉通知器
type DingtalkNotifier struct {
	webhook string
	secret  string
}

// NewDingtalkNotifier 创建钉钉通知器
func NewDingtalkNotifier(webhook, secret string) *DingtalkNotifier {
	return &DingtalkNotifier{
		webhook: webhook,
		secret:  secret,
	}
}

// Send 发送钉钉消息
func (d *DingtalkNotifier) Send(message string) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	return d.sendHTTPRequest(payload)
}

// sendHTTPRequest 发送HTTP请求（带签名）
func (d *DingtalkNotifier) sendHTTPRequest(payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	// 构建请求URL（如果有密钥则添加签名）
	requestURL := d.webhook
	if d.secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		sign := d.generateSign(timestamp)
		requestURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", d.webhook, timestamp, url.QueryEscape(sign))
	}

	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// generateSign 生成钉钉签名
func (d *DingtalkNotifier) generateSign(timestamp int64) string {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + d.secret
	h := hmac.New(sha256.New, []byte(d.secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}