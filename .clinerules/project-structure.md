# Wails 项目结构规则 (Project Structure Rule)

## 项目概述
这是一个基于 Wails v2 框架的桌面应用项目 "天灵AI工作台" (skymind)，使用 Go 后端和 Vue 3 前端。

## 目录结构说明

### 根目录文件
- `main.go` - 应用程序入口点，包含 Wails 应用配置
- `wails.json` - Wails 项目配置文件
- `go.mod` / `go.sum` - Go 模块依赖管理
- `package.json` - 前端依赖管理（根级别）
- `README.md` / `README.CN.md` - 项目说明文档

### `.clinerules/` - Cline 规则目录
- `icon-registration-rule.md` - 图标注册规则，定义 lucide-vue-next 图标的使用规范
- `project-structure.md` - 项目结构规则，描述整个 Wails 项目的目录结构和文件说明
- `development-workflow.md` - 开发工作流规则，定义代码修改后的编译测试和文档更新流程

### 后端目录结构

#### `api/` - API 层
- `api/enter.go` - API 入口文件
- `api/smart_query/` - 智能查询相关 API
  - `assistant.go` - 助手 API
  - `conversation.go` - 对话 API
  - `enter.go` - 智能查询入口
  - `message.go` - 消息 API
  - `topic.go` - 主题 API
  - `utils.go` - 工具函数

#### `appcore/` - 应用核心
- `app.go` - 应用核心逻辑
- `app_api.go` - 应用 API 方法
- `app_lifecycle.go` - 应用生命周期管理
- `system_tray.go` - 系统托盘功能
- `icon.ico` - 应用图标

#### `database/` - 数据库层
- `gorm.go` - GORM 数据库配置
- `snowflake.go` - 雪花算法 ID 生成

#### `global/` - 全局变量
- `global.go` - 全局变量定义

#### `models/` - 数据模型
- `assistant.go` - 助手模型
- `conversation.go` - 对话模型
- `memory.go` - 记忆模型
- `message.go` - 消息模型
- `topic.go` - 主题模型

#### `service/` - 业务逻辑层
- `service/enter.go` - 服务入口
- `service/smart_query/` - 智能查询服务
  - `assistant.go` - 助手服务
  - `conversation.go` - 对话服务
  - `enter.go` - 智能查询服务入口
  - `message.go` - 消息服务
  - `topic.go` - 主题服务

### 前端目录结构 (`frontend/`)

#### `frontend/src/` - 源代码目录
- `main.js` - Vue 应用入口，包含图标全局注册
- `App.vue` - 根组件
- `permission.js` - 路由权限控制
- `style.css` - 全局样式

#### `frontend/src/components/` - 公共组件
- `HelloWorld.vue` - 示例组件
- `draggableList.vue` - 拖拽列表组件

#### `frontend/src/layout/` - 布局组件
- `Layout.vue` - 主布局组件

#### `frontend/src/router/` - 路由配置
- `index.js` - 路由主配置
- `modules/wails.js` - Wails 相关路由模块

#### `frontend/src/store/` - 状态管理 (Pinia)
- `index.js` - Store 入口
- `modules/chat.js` - 聊天状态管理
- `modules/counter.js` - 计数器状态管理

#### `frontend/src/views/` - 页面组件
- `HomeView.vue` - 首页
- `FileInterpreter/index.vue` - 文件解释器页面
- `FullSearch/index.vue` - 全文搜索页面
- `SmartQuery/` - 智能查询模块
  - `index.vue` - 智能查询主页
  - `assistants.vue` - 助手管理
  - `chat.vue` - 聊天界面
  - `conversation.vue` - 对话管理
  - `action.vue` - 操作界面
  - `switch.vue` - 切换界面
  - `send.vue` - 用户消息组件，用于显示用户询问部分
  - `reply.vue` - AI回复组件，用于显示AI回复部分
- `SystemAdvisor/index.vue` - 系统使用顾问页面
- `TaskAssistant/index.vue` - 个人计划助理页面

#### `frontend/src/assets/` - 静态资源
- `fonts/` - 字体文件
- `images/` - 图片资源

#### `frontend/wailsjs/` - Wails 生成的 TypeScript 绑定
- `go/main/` - Go 后端 API 的前端绑定
- `runtime/` - Wails 运行时绑定

### 构建和脚本目录

#### `build/` - 构建输出
- `bin/` - 编译后的可执行文件
- `windows/` - Windows 特定构建文件

#### `scripts/` - 构建脚本
- `build-*.sh` - 各平台构建脚本
- `install-wails-cli.sh` - Wails CLI 安装脚本

## 技术栈

### 后端
- **Go** - 主要编程语言
- **Wails v2** - 桌面应用框架
- **GORM** - ORM 框架

### 前端
- **Vue 3** - 前端框架
- **Vite** - 构建工具
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **TDesign Vue Next** - UI 组件库
- **TDesign Chat** - 聊天组件
- **Axios** - HTTP 客户端
- **Lucide Vue Next** - 图标库

## 开发注意事项

1. **前后端通信**：通过 Wails 自动生成的绑定进行
2. **图标使用**：必须遵循 `icon-registration-rule.md` 中的规则
3. **状态管理**：使用 Pinia 进行前端状态管理
4. **路由**：使用 Vue Router 管理前端路由
5. **样式**：使用 TDesign 组件库和自定义 CSS
6. **构建**：使用 `wails dev` 进行开发，`wails build` 进行生产构建

## 模块说明

### SmartQuery 模块
项目的核心功能模块，包含：
- 助手管理 (assistants)
- 对话管理 (conversation)
- 聊天界面 (chat)
- 操作界面 (action)
- 切换界面 (switch)

### 其他功能模块
- FileInterpreter - 文件解释器
- FullSearch - 全文搜索
- SystemAdvisor - 系统使用顾问
- TaskAssistant - 个人计划助理
