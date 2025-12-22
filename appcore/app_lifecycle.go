package appcore

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"syscall"
	"unsafe"

	sysruntime "runtime"
	"skymind/database"
	"skymind/logger"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Startup is called at application startup
func Startup(a *App, ctx context.Context) {
	// 设置 panic 捕获
	defer func() {
		if r := recover(); r != nil {
			// 捕获到 panic，记录详细信息
			buf := make([]byte, 4096)
			n := sysruntime.Stack(buf, false)
			stackTrace := string(buf[:n])
			
			logger.LogError("Application panic recovered", fmt.Errorf("panic: %v", r), logrus.Fields{
				"panicValue": r,
				"stackTrace": stackTrace,
				"goroutines": sysruntime.NumGoroutine(),
				"version": sysruntime.Version(),
			})
		}
	}()

	// Perform your setup here
	a.Ctx = ctx

	// 初始化日志系统
	if err := logger.InitLogger(); err != nil {
		// 如果日志初始化失败，使用标准输出记录错误
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.LogInfo("Application starting up")

	// 初始化用户配置文件
	if err := database.EnsureUserConfigFile(); err != nil {
		logger.LogError("Failed to initialize user config file", err)
	} else {
		logger.LogInfo("User config file initialized successfully")
	}

	// 初始化数据库
	if err := InitDatabase(a); err != nil {
		logger.LogError("Failed to initialize database", err)
	}
}

// DomReady is called after front-end resources have been loaded
func DomReady(a *App, ctx context.Context) {
	// 应用启动时显示窗口
	if a.Ctx != nil {
		runtime.WindowShow(a.Ctx)
	}
}

// BeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func BeforeClose(a *App, ctx context.Context) (prevent bool) {
	if a.Quitting {
		return false // 如果是退出操作，允许关闭
	}

	// 如果不是退出操作，隐藏窗口到系统托盘
	HideWindow(a)
	a.InTray = true
	return true // 阻止应用关闭
}

// Shutdown is called at application termination
func Shutdown(a *App, ctx context.Context) {
	// 释放互斥体
	ReleaseMutex(a)
	// Perform your teardown here
}

// ShowWindow 显示应用窗口
func ShowWindow(a *App) {
	if a.Ctx != nil {
		runtime.WindowShow(a.Ctx)
		a.InTray = false // 窗口显示时，不在托盘中
	}
}

// HideWindow 隐藏应用窗口
func HideWindow(a *App) {
	if a.Ctx != nil {
		runtime.WindowHide(a.Ctx)
		a.InTray = true // 窗口隐藏时，在托盘中
	}
}

// QuitApp 退出应用
func QuitApp(a *App) {
	a.Quitting = false
	runtime.Quit(a.Ctx)
}

// TerminateProcess 终结进程
func TerminateProcess(a *App) {
	a.Quitting = true
	runtime.Quit(a.Ctx)
}

// Greet returns a greeting for the given name
func Greet(a *App, name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// InitDatabase 初始化数据库
func InitDatabase(a *App) error {
	logger.LogInfo("Initializing database")

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.LogError("Failed to get user home directory", err)
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// 创建skymind数据目录
	dataDir := homeDir + "\\skymind"
	logger.LogInfo("Database directory", map[string]interface{}{"dataDir": dataDir})

	// 创建GORM数据库实例
	_, err = database.NewGormDatabase(dataDir)
	if err != nil {
		logger.LogError("Failed to create GORM database", err)
		return fmt.Errorf("failed to create gorm database: %w", err)
	}

	logger.LogInfo("GORM Database initialized successfully")
	return nil
}

// API方法 - 委托给SmartQueryAPI
func (a *App) GetAssistants() ([]interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetAssistants()
}

func (a *App) GetAssistantByID(id string) (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetAssistantByID(id)
}

func (a *App) CreateAssistant(assistant map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.CreateAssistant(assistant)
}

func (a *App) UpdateAssistant(assistant map[string]interface{}) error {
	return a.SmartQueryAPI.AssistantAPI.UpdateAssistant(assistant)
}

func (a *App) DeleteAssistant(id string) error {
	return a.SmartQueryAPI.AssistantAPI.DeleteAssistant(id)
}

func (a *App) GetDefaultAssistant() (interface{}, error) {
	return a.SmartQueryAPI.AssistantAPI.GetDefaultAssistant()
}

func (a *App) GetTopics(assistantID string) ([]interface{}, error) {
	return a.SmartQueryAPI.TopicAPI.GetTopics(assistantID)
}

func (a *App) CreateTopic(topic map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.TopicAPI.CreateTopic(topic)
}

func (a *App) GetMessages(topicID string) ([]interface{}, error) {
	return a.SmartQueryAPI.MessageAPI.GetMessages(topicID)
}

func (a *App) GetConversations(topicID string) ([]interface{}, error) {
	return a.SmartQueryAPI.ConversationAPI.GetConversations(topicID)
}

func (a *App) CreateConversation(conversation map[string]interface{}) (string, error) {
	return a.SmartQueryAPI.ConversationAPI.CreateConversation(conversation)
}

func (a *App) CreateMessage(message map[string]interface{}) (interface{}, error) {
	return a.SmartQueryAPI.MessageAPI.CreateMessage(message)
}

func (a *App) StreamChatCompletion(streamID, streamType, relatedID string, messages []map[string]interface{}, modelType string) error {
	return a.SmartQueryAPI.MessageAPI.StreamChatCompletion(a.Ctx, streamID, streamType, relatedID, messages, modelType)
}

func (a *App) UpdateAssistantsSortOrder(sortOrders []map[string]interface{}) error {
	return a.SmartQueryAPI.AssistantAPI.UpdateAssistantsSortOrder(sortOrders)
}

func (a *App) UpdateTopicsSortOrder(sortOrders []map[string]interface{}) error {
	return a.SmartQueryAPI.TopicAPI.UpdateTopicsSortOrder(sortOrders)
}

// Windows API 相关常量和函数
const (
	ERROR_ALREADY_EXISTS = 183
	MUTEX_NAME           = "Global\\SkymindAppMutex"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW        = kernel32.NewProc("CreateMutexW")
	procReleaseMutex        = kernel32.NewProc("ReleaseMutex")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	user32                  = syscall.NewLazyDLL("user32.dll")
	procFindWindowW         = user32.NewProc("FindWindowW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procIsIconic            = user32.NewProc("IsIconic")
	procOpenIcon            = user32.NewProc("OpenIcon")
)

// SW_RESTORE 常量
const SW_RESTORE = 9

// CheckSingleInstance 检查是否为单实例
// 返回 true 表示可以继续运行，false 表示已有实例在运行
func CheckSingleInstance(a *App) bool {
	// 创建命名互斥体
	mutexName, _ := syscall.UTF16PtrFromString(MUTEX_NAME)

	ret, _, err := procCreateMutexW.Call(
		0,
		0,
		uintptr(unsafe.Pointer(mutexName)),
	)

	if ret == 0 {
		// 创建互斥体失败
		return false
	}

	a.Mutex = syscall.Handle(ret)

	// 检查错误码
	if errno, ok := err.(syscall.Errno); ok && errno == ERROR_ALREADY_EXISTS {
		// 互斥体已存在，说明已有实例在运行
		procCloseHandle.Call(uintptr(a.Mutex))
		a.Mutex = 0

		// 尝试激活已有窗口
		bringExistingWindowToFront(a)
		return false
	}

	return true
}

// bringExistingWindowToFront 激活已有窗口到前台
func bringExistingWindowToFront(a *App) {
	// 查找窗口标题为"AI工作台"的窗口
	windowTitle, _ := syscall.UTF16PtrFromString("AI工作台")

	hwnd, _, _ := procFindWindowW.Call(
		0,
		uintptr(unsafe.Pointer(windowTitle)),
	)

	if hwnd != 0 {
		// 检查窗口是否最小化
		isIconic, _, _ := procIsIconic.Call(hwnd)
		if isIconic != 0 {
			// 如果窗口最小化，则恢复
			procOpenIcon.Call(hwnd)
		}

		// 显示窗口
		procShowWindow.Call(hwnd, SW_RESTORE)

		// 设置为前台窗口
		procSetForegroundWindow.Call(hwnd)
	} else {
		// 如果找不到窗口，尝试查找类名
		className, _ := syscall.UTF16PtrFromString("Chrome_WidgetWin_1")
		hwnd, _, _ = procFindWindowW.Call(
			uintptr(unsafe.Pointer(className)),
			0,
		)

		if hwnd != 0 {
			// 检查窗口是否最小化
			isIconic, _, _ := procIsIconic.Call(hwnd)
			if isIconic != 0 {
				// 如果窗口最小化，则恢复
				procOpenIcon.Call(hwnd)
			}

			// 显示窗口
			procShowWindow.Call(hwnd, SW_RESTORE)

			// 设置为前台窗口
			procSetForegroundWindow.Call(hwnd)
		}
	}
}

// ReleaseMutex 释放互斥体
func ReleaseMutex(a *App) {
	if a.Mutex != 0 {
		procReleaseMutex.Call(uintptr(a.Mutex))
		procCloseHandle.Call(uintptr(a.Mutex))
		a.Mutex = 0
	}
}
