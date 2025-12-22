package appcore

import (
	"embed"
	"skymind/logger"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconFS embed.FS

// Windows API 相关常量和函数
const (
	MOD_ALT         = 0x0001
	VK_A            = 0x41
	WM_HOTKEY       = 0x0312
	HOTKEY_ID       = 1
)

var (
	trayUser32           = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey = trayUser32.NewProc("RegisterHotKey")
	procUnregisterHotKey = trayUser32.NewProc("UnregisterHotKey")
)

// InitTray 初始化系统托盘
func InitTray(a *App) {
	logger.LogInfo("Initializing system tray")

	// 注册全局热键 Alt+A
	go func() {
		if registerGlobalHotKey() {
			logger.LogInfo("Global hotkey Alt+A registered successfully")
			// 启动热键监听
			go listenHotKey(a)
		} else {
			logger.LogError("Failed to register global hotkey Alt+A", nil)
		}
	}()

	go func() {
		systray.Run(onTrayReady(a), onTrayExit(a))
		logger.LogInfo("System tray initialized successfully")
	}()
}

// onTrayReady 托盘准备就绪时的回调函数
func onTrayReady(a *App) func() {
	return func() {
		// 设置托盘图标和标题
		systray.SetIcon(getIconData())
		systray.SetTitle("天灵AI工作台")
		systray.SetTooltip("天灵AI工作台")

		// 添加显示窗口菜单项
		showWindow := systray.AddMenuItem("显示窗口", "显示主窗口")

		// 添加分隔线
		systray.AddSeparator()

		// 添加退出菜单项
		quitApp := systray.AddMenuItem("退出", "退出应用")

		// 监听菜单项点击事件
		go func() {
			for {
				select {
				case <-showWindow.ClickedCh:
					logger.LogInfo("Show window clicked from system tray")
					ShowWindow(a)
				case <-quitApp.ClickedCh:
					logger.LogInfo("Quit application clicked from system tray")
					TerminateProcess(a)
				}
			}
		}()
	}
}

// onTrayExit 托盘退出时的回调函数
func onTrayExit(a *App) func() {
	return func() {
		// 清理资源
		a.Quitting = true
	}
}

// getIconData 获取托盘图标数据
func getIconData() []byte {
	// 从嵌入的文件系统读取图标
	if iconData, err := iconFS.ReadFile("icon.ico"); err == nil {
		return iconData
	}

	// 如果嵌入失败，创建一个简单的16x16像素的蓝色图标作为备用
	icon := make([]byte, 16*16*4)
	for i := 0; i < len(icon); i += 4 {
		icon[i] = 0x00   // R
		icon[i+1] = 0x78 // G
		icon[i+2] = 0xD6 // B
		icon[i+3] = 0xFF // A
	}
	return icon
}

// registerGlobalHotKey 注册全局热键 Alt+A
func registerGlobalHotKey() bool {
	ret, _, err := procRegisterHotKey.Call(
		0,                    // hWnd: 0 表示关联到当前线程
		HOTKEY_ID,           // 热键ID
		MOD_ALT,             // 修饰键: Alt
		VK_A,                // 虚拟键码: A
	)
	
	if ret != 0 {
		return true
	}
	
	logger.LogError("RegisterHotKey failed", err)
	return false
}

// unregisterGlobalHotKey 注销全局热键
func unregisterGlobalHotKey() {
	procUnregisterHotKey.Call(
		0,        // hWnd: 0 表示关联到当前线程
		HOTKEY_ID, // 热键ID
	)
}

// listenHotKey 监听热键消息
func listenHotKey(a *App) {
	logger.LogInfo("Starting hotkey listener")
	
	// 创建一个通道来监听应用退出
	done := make(chan struct{})
	
	// 检查 context 是否为 nil
	if a.Ctx != nil {
		go func() {
			<-a.Ctx.Done()
			close(done)
		}()
	} else {
		// 如果 context 为 nil，创建一个永远不会关闭的通道
		go func() {
			select {}
		}()
	}
	
	// 记录上一次的按键状态，防止重复触发
	var wasPressed bool
	
	// 使用轮询方式监听热键
	for {
		select {
		case <-done:
			// 应用退出时清理热键
			unregisterGlobalHotKey()
			logger.LogInfo("Hotkey listener stopped")
			return
		default:
			// 检查热键状态
			isPressed := isHotKeyPressed()
			
			// 只在按键从未按下变为按下时触发（防止重复触发）
			if isPressed && !wasPressed {
				if isWindowVisible(a) {
					logger.LogInfo("Alt+A hotkey detected, hiding window")
					HideWindow(a)
				} else {
					logger.LogInfo("Alt+A hotkey detected, showing window")
					ShowWindow(a)
				}
				// 等待一段时间防止重复触发
				time.Sleep(500 * time.Millisecond)
			}
			
			wasPressed = isPressed
			time.Sleep(50 * time.Millisecond) // 50ms 轮询间隔，提高响应速度
		}
	}
}

// isHotKeyPressed 检查 Alt+A 是否被按下
func isHotKeyPressed() bool {
	// 获取 Alt 键状态（检查左右 Alt 键）
	leftAltState := getKeyState(0xA4) // VK_LMENU
	rightAltState := getKeyState(0xA5) // VK_RMENU
	
	altKeyPressed := (leftAltState < 0) || (rightAltState < 0)
	if !altKeyPressed {
		return false
	}
	
	// 获取 A 键状态
	aKeyState := getKeyState(0x41) // VK_A
	if aKeyState >= 0 {
		return false
	}
	
	return true
}

// getKeyState 获取按键状态
func getKeyState(virtualKey int) int16 {
	procGetKeyState := trayUser32.NewProc("GetKeyState")
	ret, _, _ := procGetKeyState.Call(uintptr(virtualKey))
	return int16(ret)
}

// isWindowVisible 检查窗口是否可见
func isWindowVisible(a *App) bool {
	if a.Ctx == nil {
		return false
	}
	
	// 简化实现：基于应用的 InTray 状态来判断窗口可见性
	// 如果应用在托盘中，说明窗口被隐藏了
	return !a.InTray
}
