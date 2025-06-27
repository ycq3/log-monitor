package monitor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"log-monitor/config"
	"log-monitor/notifier"
)

// LogMonitor 日志监控器
type LogMonitor struct {
	watcher   *fsnotify.Watcher
	config    *config.Config
	notifiers []notifier.Notifier
	filePos   map[string]int64 // 记录文件读取位置
}

// NewLogMonitor 创建新的日志监控器
func NewLogMonitor(cfg *config.Config, notifiers []notifier.Notifier) (*LogMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %v", err)
	}

	return &LogMonitor{
		watcher:   watcher,
		config:    cfg,
		notifiers: notifiers,
		filePos:   make(map[string]int64),
	}, nil
}

// Start 开始监控
func (m *LogMonitor) Start() error {
	// 添加监控文件
	for _, logFile := range m.config.LogFiles {
		if !logFile.Enabled {
			continue
		}

		err := m.watcher.Add(logFile.Path)
		if err != nil {
			log.Printf("添加监控文件失败 %s: %v", logFile.Path, err)
			continue
		}

		// 初始化文件位置
		if stat, err := os.Stat(logFile.Path); err == nil {
			m.filePos[logFile.Path] = stat.Size()
		}

		log.Printf("开始监控文件: %s", logFile.Path)
	}

	// 启动监控循环
	go m.watchLoop()

	return nil
}

// Stop 停止监控
func (m *LogMonitor) Stop() error {
	return m.watcher.Close()
}

// watchLoop 监控循环
func (m *LogMonitor) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				m.handleFileWrite(event.Name)
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("监控错误: %v", err)
		}
	}
}

// handleFileWrite 处理文件写入事件
func (m *LogMonitor) handleFileWrite(filePath string) {
	// 查找对应的日志文件配置
	var logFileConfig *config.LogFile
	for _, lf := range m.config.LogFiles {
		if lf.Path == filePath && lf.Enabled {
			logFileConfig = &lf
			break
		}
	}

	if logFileConfig == nil {
		return
	}

	// 读取新增内容
	newLines, err := m.readNewLines(filePath)
	if err != nil {
		log.Printf("读取文件新内容失败 %s: %v", filePath, err)
		return
	}

	// 检查关键词
	for _, line := range newLines {
		if m.containsKeywords(line, logFileConfig.Keywords) {
			m.sendAlert(filePath, line)
		}
	}
}

// readNewLines 读取文件新增行
func (m *LogMonitor) readNewLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取当前文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	currentSize := stat.Size()
	lastPos := m.filePos[filePath]

	// 如果文件被截断或重新创建
	if currentSize < lastPos {
		lastPos = 0
	}

	// 定位到上次读取位置
	_, err = file.Seek(lastPos, 0)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 更新文件位置
	m.filePos[filePath] = currentSize

	return lines, scanner.Err()
}

// containsKeywords 检查行是否包含关键词
func (m *LogMonitor) containsKeywords(line string, keywords []string) bool {
	lineLower := strings.ToLower(line)
	for _, keyword := range keywords {
		if strings.Contains(lineLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// sendAlert 发送告警
func (m *LogMonitor) sendAlert(filePath, line string) {
	message := fmt.Sprintf("🚨 日志告警\n\n文件: %s\n时间: %s\n内容: %s",
		filePath,
		time.Now().Format("2006-01-02 15:04:05"),
		line)

	for _, n := range m.notifiers {
		go func(notifier notifier.Notifier) {
			if err := notifier.Send(message); err != nil {
				log.Printf("发送通知失败: %v", err)
			}
		}(n)
	}
}