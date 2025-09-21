package monitor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"log-monitor/config"
	"log-monitor/notifier"
)

// LogMonitor 日志监控器
type LogMonitor struct {
	watcher      *fsnotify.Watcher
	config       *config.Config
	notifiers    []notifier.Notifier
	filePos      map[string]int64 // 记录文件读取位置
	watchedFiles map[string]*config.LogFile // 监控的文件映射
	watchedDirs  map[string]*config.LogDirectory // 监控的目录映射
	mu           sync.RWMutex // 保护并发访问
	maxFileSize  int64        // 最大文件大小限制 (默认100MB)
	bufferSize   int          // 读取缓冲区大小 (默认64KB)
}

// NewLogMonitor 创建新的日志监控器
func NewLogMonitor(cfg *config.Config, notifiers []notifier.Notifier) (*LogMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %v", err)
	}

	return &LogMonitor{
		watcher:      watcher,
		config:       cfg,
		notifiers:    notifiers,
		filePos:      make(map[string]int64),
		watchedFiles: make(map[string]*config.LogFile),
		watchedDirs:  make(map[string]*config.LogDirectory),
		maxFileSize:  100 * 1024 * 1024, // 100MB
		bufferSize:   64 * 1024,          // 64KB
	}, nil
}

// Start 开始监控
func (m *LogMonitor) Start() error {
	// 添加监控文件
	for _, logFile := range m.config.LogFiles {
		if !logFile.Enabled {
			continue
		}

		err := m.addFileWatch(logFile.Path, &logFile)
		if err != nil {
			log.Printf("添加监控文件失败 %s: %v", logFile.Path, err)
			continue
		}

		log.Printf("开始监控文件: %s", logFile.Path)
	}

	// 添加监控目录
	for _, logDir := range m.config.LogDirectories {
		if !logDir.Enabled {
			continue
		}

		err := m.addDirectoryWatch(&logDir)
		if err != nil {
			log.Printf("添加监控目录失败 %s: %v", logDir.Path, err)
			continue
		}

		log.Printf("开始监控目录: %s (递归: %v)", logDir.Path, logDir.Recursive)
	}

	// 启动监控循环
	go m.watchLoop()

	// 启动定期清理任务
	go m.cleanupLoop()

	return nil
}

// Stop 停止监控
func (m *LogMonitor) Stop() error {
	return m.watcher.Close()
}

// addFileWatch 添加文件监控
func (m *LogMonitor) addFileWatch(filePath string, logFile *config.LogFile) error {
	err := m.watcher.Add(filePath)
	if err != nil {
		return err
	}

	// 初始化文件位置
	if stat, err := os.Stat(filePath); err == nil {
		m.mu.Lock()
		m.filePos[filePath] = stat.Size()
		m.mu.Unlock()
	}

	// 记录监控的文件
	m.mu.Lock()
	m.watchedFiles[filePath] = logFile
	m.mu.Unlock()
	return nil
}

// addDirectoryWatch 添加目录监控
func (m *LogMonitor) addDirectoryWatch(logDir *config.LogDirectory) error {
	// 记录监控的目录
	m.mu.Lock()
	m.watchedDirs[logDir.Path] = logDir
	m.mu.Unlock()

	if logDir.Recursive {
		return m.addRecursiveWatch(logDir)
	} else {
		return m.addSingleDirWatch(logDir)
	}
}

// addSingleDirWatch 添加单个目录监控
func (m *LogMonitor) addSingleDirWatch(logDir *config.LogDirectory) error {
	err := m.watcher.Add(logDir.Path)
	if err != nil {
		return err
	}

	// 扫描现有文件
	return m.scanExistingFiles(logDir.Path, logDir, false)
}

// addRecursiveWatch 添加递归目录监控
func (m *LogMonitor) addRecursiveWatch(logDir *config.LogDirectory) error {
	return filepath.Walk(logDir.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 检查是否在排除列表中
			if m.isExcludedDir(path, logDir.ExcludeDirs) {
				return filepath.SkipDir
			}

			// 添加目录监控
			if err := m.watcher.Add(path); err != nil {
				log.Printf("添加目录监控失败 %s: %v", path, err)
				return nil // 继续处理其他目录
			}

			// 扫描目录中的现有文件
			return m.scanExistingFiles(path, logDir, false)
		}

		return nil
	})
}

// scanExistingFiles 扫描现有文件
func (m *LogMonitor) scanExistingFiles(dirPath string, logDir *config.LogDirectory, isNewDir bool) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		
		// 检查文件扩展名
		if !m.matchesExtensions(filePath, logDir.Extensions) {
			continue
		}

		// 检查文件大小限制
		if stat, err := os.Stat(filePath); err == nil {
			if stat.Size() > m.maxFileSize {
				log.Printf("跳过大文件 %s (大小: %d bytes, 限制: %d bytes)", 
					filePath, stat.Size(), m.maxFileSize)
				continue
			}

			// 初始化文件位置
			m.mu.Lock()
			if isNewDir {
				// 新目录，从文件末尾开始监控
				m.filePos[filePath] = stat.Size()
			} else {
				// 现有目录，从文件末尾开始监控（避免重复处理历史日志）
				m.filePos[filePath] = stat.Size()
			}
			m.mu.Unlock()
		}
	}

	return nil
}

// isExcludedDir 检查目录是否被排除
func (m *LogMonitor) isExcludedDir(dirPath string, excludeDirs []string) bool {
	for _, excludeDir := range excludeDirs {
		if strings.Contains(dirPath, excludeDir) {
			return true
		}
	}
	return false
}

// matchesExtensions 检查文件是否匹配扩展名
func (m *LogMonitor) matchesExtensions(filePath string, extensions []string) bool {
	fileExt := strings.ToLower(filepath.Ext(filePath))
	for _, ext := range extensions {
		if strings.ToLower(ext) == fileExt {
			return true
		}
	}
	return false
}
// watchLoop 监控循环
func (m *LogMonitor) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			m.handleEvent(event)

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("监控错误: %v", err)
		}
	}
}

// handleEvent 处理文件系统事件
func (m *LogMonitor) handleEvent(event fsnotify.Event) {
	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		m.handleFileWrite(event.Name)
	case event.Op&fsnotify.Create == fsnotify.Create:
		m.handleFileCreate(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		m.handleFileRemove(event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		m.handleFileRename(event.Name)
	}
}

// handleFileCreate 处理文件创建事件
func (m *LogMonitor) handleFileCreate(filePath string) {
	// 检查是否是目录中的新文件
	m.mu.RLock()
	for watchedDir, logDir := range m.watchedDirs {
		if m.isFileInDirectory(filePath, watchedDir, logDir.Recursive) {
			// 检查文件扩展名
			if m.matchesExtensions(filePath, logDir.Extensions) {
				// 检查文件大小限制
				if stat, err := os.Stat(filePath); err == nil && stat.Size() <= m.maxFileSize {
					// 初始化文件位置（新文件从头开始）
					m.mu.RUnlock()
					m.mu.Lock()
					m.filePos[filePath] = 0
					m.mu.Unlock()
					log.Printf("检测到新日志文件: %s", filePath)
				} else if err == nil {
					log.Printf("跳过大文件 %s (大小: %d bytes)", filePath, stat.Size())
				}
			}
			break
		}
	}
	if len(m.watchedDirs) > 0 {
		m.mu.RUnlock()
	}
}

// handleFileRemove 处理文件删除事件
func (m *LogMonitor) handleFileRemove(filePath string) {
	// 清理文件位置记录
	m.mu.Lock()
	delete(m.filePos, filePath)
	m.mu.Unlock()
	log.Printf("日志文件已删除: %s", filePath)
}

// handleFileRename 处理文件重命名事件
func (m *LogMonitor) handleFileRename(filePath string) {
	// 清理旧文件位置记录
	m.mu.Lock()
	delete(m.filePos, filePath)
	m.mu.Unlock()
	log.Printf("日志文件已重命名: %s", filePath)
}

// isFileInDirectory 检查文件是否在监控目录中
func (m *LogMonitor) isFileInDirectory(filePath, dirPath string, recursive bool) bool {
	if recursive {
		return strings.HasPrefix(filePath, dirPath)
	} else {
		return filepath.Dir(filePath) == dirPath
	}
}

// handleFileWrite 处理文件写入事件
func (m *LogMonitor) handleFileWrite(filePath string) {
	// 查找对应的日志文件配置
	var keywords []string
	
	// 检查是否是直接监控的文件
	m.mu.RLock()
	if logFile, exists := m.watchedFiles[filePath]; exists {
		keywords = logFile.Keywords
	} else {
		// 检查是否是目录监控中的文件
		for watchedDir, logDir := range m.watchedDirs {
			if m.isFileInDirectory(filePath, watchedDir, logDir.Recursive) {
				// 检查文件扩展名
				if m.matchesExtensions(filePath, logDir.Extensions) {
					keywords = logDir.Keywords
					break
				}
			}
		}
	}
	m.mu.RUnlock()

	if len(keywords) == 0 {
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
		if m.containsKeywords(line, keywords) {
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
	
	// 检查文件大小限制
	if currentSize > m.maxFileSize {
		log.Printf("文件 %s 超过大小限制，跳过读取", filePath)
		return nil, nil
	}

	m.mu.RLock()
	lastPos := m.filePos[filePath]
	m.mu.RUnlock()

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
	
	// 设置缓冲区大小
	buf := make([]byte, 0, m.bufferSize)
	scanner.Buffer(buf, m.bufferSize)
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 更新文件位置
	m.mu.Lock()
	m.filePos[filePath] = currentSize
	m.mu.Unlock()

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

// cleanupLoop 定期清理任务
func (m *LogMonitor) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Minute) // 每30分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		}
	}
}

// performCleanup 执行清理任务
func (m *LogMonitor) performCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理不存在的文件记录
	for filePath := range m.filePos {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			delete(m.filePos, filePath)
			log.Printf("清理不存在的文件记录: %s", filePath)
		}
	}

	// 检查文件大小变化，重置过大文件的位置
	for filePath, pos := range m.filePos {
		if stat, err := os.Stat(filePath); err == nil {
			if stat.Size() > m.maxFileSize && pos < stat.Size() {
				// 文件变得过大，从末尾开始监控
				m.filePos[filePath] = stat.Size()
				log.Printf("重置大文件位置: %s (大小: %d bytes)", filePath, stat.Size())
			}
		}
	}

	log.Printf("内存清理完成，当前监控文件数: %d", len(m.filePos))
}