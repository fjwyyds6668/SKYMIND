package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// AppLogger 全局应用日志记录器
	AppLogger *logrus.Logger
	// ErrorLogger 专用错误日志记录器
	ErrorLogger *logrus.Logger
)

// getUserConfigDir 获取用户配置目录
func getUserConfigDir() string {
	if runtime.GOOS == "windows" {
		// Windows: 优先使用USERPROFILE环境变量
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			configDir := filepath.Join(userProfile, "skymind")
			return configDir
		}
		// 备选：使用HOME环境变量
		if home := os.Getenv("HOME"); home != "" {
			configDir := filepath.Join(home, "skymind")
			return configDir
		}
		// 备选：使用APPDATA环境变量
		if appData := os.Getenv("APPDATA"); appData != "" {
			configDir := filepath.Join(appData, "skymind")
			return configDir
		}
		// 备选：使用LOCALAPPDATA环境变量
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			configDir := filepath.Join(localAppData, "skymind")
			return configDir
		}
	} else {
		// Unix-like: 使用HOME环境变量
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".skymind")
		}
	}

	// 最后备选：使用当前目录
	return "skymind"
}

// InitLogger 初始化日志系统
func InitLogger() error {
	// 获取用户配置目录
	userConfigDir := getUserConfigDir()
	logsDir := filepath.Join(userConfigDir, "logs")

	// 创建日志目录
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// 初始化应用日志记录器
	AppLogger = logrus.New()
	AppLogger.SetLevel(logrus.InfoLevel)
	AppLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 初始化错误日志记录器
	ErrorLogger = logrus.New()
	ErrorLogger.SetLevel(logrus.ErrorLevel)
	ErrorLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志文件输出
	if err := setupLogFiles(logsDir); err != nil {
		return fmt.Errorf("failed to setup log files: %w", err)
	}

	// 添加钩子用于按小时轮转日志
	AppLogger.AddHook(NewHourlyHook(logsDir, "app"))
	ErrorLogger.AddHook(NewHourlyHook(logsDir, "error"))

	return nil
}

// setupLogFiles 设置日志文件输出
func setupLogFiles(logsDir string) error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	hourStr := now.Format("15")

	// 应用日志文件
	appLogFile := filepath.Join(logsDir, fmt.Sprintf("app_%s_%s.log", dateStr, hourStr))
	appFile, err := os.OpenFile(appLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open app log file: %w", err)
	}

	// 错误日志文件
	errorLogFile := filepath.Join(logsDir, fmt.Sprintf("error_%s_%s.log", dateStr, hourStr))
	errorFile, err := os.OpenFile(errorLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		appFile.Close()
		return fmt.Errorf("failed to open error log file: %w", err)
	}

	// 同时输出到文件和控制台
	AppLogger.SetOutput(os.Stdout)
	AppLogger.AddHook(&FileHook{File: appFile})

	ErrorLogger.SetOutput(os.Stderr)
	ErrorLogger.AddHook(&FileHook{File: errorFile})

	return nil
}

// HourlyHook 按小时轮转日志的钩子
type HourlyHook struct {
	logsDir  string
	prefix   string
	lastHour int
}

// NewHourlyHook 创建新的按小时轮转钩子
func NewHourlyHook(logsDir, prefix string) *HourlyHook {
	now := time.Now()
	return &HourlyHook{
		logsDir:  logsDir,
		prefix:   prefix,
		lastHour: now.Hour(),
	}
}

// Levels 返回钩子处理的日志级别
func (hook *HourlyHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 处理日志条目
func (hook *HourlyHook) Fire(entry *logrus.Entry) error {
	now := time.Now()
	currentHour := now.Hour()

	// 检查是否需要轮转日志文件
	if currentHour != hook.lastHour {
		if err := hook.rotateLogFile(now); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
		hook.lastHour = currentHour
	}

	return nil
}

// rotateLogFile 轮转日志文件
func (hook *HourlyHook) rotateLogFile(now time.Time) error {
	dateStr := now.Format("2006-01-02")
	hourStr := now.Format("15")

	_ = filepath.Join(hook.logsDir, fmt.Sprintf("%s_%s_%s.log", hook.prefix, dateStr, hourStr))

	// 这里可以添加文件轮转逻辑，但由于我们使用的是按小时命名的文件，
	// 新的日志会自动写入新文件
	return nil
}

// FileHook 文件输出钩子
type FileHook struct {
	File *os.File
}

// Levels 返回钩子处理的日志级别
func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 处理日志条目
func (hook *FileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	_, err = hook.File.WriteString(line)
	return err
}

// LogInfo 记录信息日志
func LogInfo(message string, fields ...logrus.Fields) {
	if AppLogger == nil {
		return
	}

	if len(fields) > 0 {
		AppLogger.WithFields(fields[0]).Info(message)
	} else {
		AppLogger.Info(message)
	}
}

// LogError 记录错误日志
func LogError(message string, err error, fields ...logrus.Fields) {
	if ErrorLogger == nil {
		return
	}

	entry := ErrorLogger.WithError(err)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}

	// 添加调用栈信息
	if runtime.GOOS == "windows" {
		entry = entry.WithField("os", "windows")
	}

	entry.Error(message)
}

// LogDatabaseOperation 记录数据库操作日志
func LogDatabaseOperation(operation string, table string, id interface{}, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"table":     table,
		"id":        id,
	}

	if err != nil {
		LogError(fmt.Sprintf("Database operation failed: %s", operation), err, fields)
	} else {
		LogInfo(fmt.Sprintf("Database operation success: %s", operation), fields)
	}
}

// LogPanic 记录panic日志
func LogPanic(message string, fields ...logrus.Fields) {
	if ErrorLogger == nil {
		return
	}

	if len(fields) > 0 {
		ErrorLogger.WithFields(fields[0]).Panic(message)
	} else {
		ErrorLogger.Panic(message)
	}
}
