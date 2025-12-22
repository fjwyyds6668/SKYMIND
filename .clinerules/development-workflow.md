# 开发工作流规则 (Development Workflow Rule)

## 规则描述

此规则定义了在 Wails 项目中进行开发时必须遵循的工作流程，确保代码质量和项目文档的同步更新。

## 开发工作流程

### 1. 代码修改后的编译测试

每次修改代码后，必须执行以下步骤进行编译测试：

#### 1.1 运行开发服务器

```bash
wails dev
```

#### 1.2 验证要点

- **后端编译**：确保 Go 代码无语法错误，能够正常编译
- **前端构建**：确保 Vue 3 前端能够正常构建和热重载
- **应用启动**：确保桌面应用能够正常启动
- **功能测试**：验证修改的功能是否正常工作
- **控制台检查**：检查浏览器控制台和终端是否有错误信息

#### 1.3 测试检查清单

- [ ] 应用能够正常启动
- [ ] 前端页面能够正常加载
- [ ] 修改的功能模块正常工作
- [ ] 没有编译错误或警告
- [ ] 控制台无错误信息
- [ ] 前后端通信正常

### 2. 创建新文件后的项目结构更新

#### 2.1 更新项目结构文档

当在项目中创建新文件或目录时，必须同步更新 `.clinerules/project-structure.md` 文件。

#### 2.2 更新要求

- **新增文件**：在相应的目录结构下添加文件说明
- **新增目录**：在相应的层级下添加目录说明和包含的文件列表
- **文件描述**：提供简洁明了的文件功能描述
- **保持格式**：遵循现有的文档格式和缩进规则

#### 2.3 更新示例

```
#### `frontend/src/components/` - 公共组件
- `HelloWorld.vue` - 示例组件
- `draggableList.vue` - 拖拽列表组件
+ `NewComponent.vue` - 新增组件描述
```

#### 2.4 特殊情况处理

- **Wails 生成文件**：`frontend/wailsjs/` 目录下的自动生成文件不需要手动更新
- **构建输出文件**：`build/` 目录下的构建输出文件不需要手动更新
- **依赖文件**：`node_modules/`、`package-lock.json` 等依赖文件不需要记录

### 3. 开发最佳实践

#### 3.1 修改前检查

- 确认修改范围和影响
- 备份重要文件（如需要）
- 查看相关的规则文档

#### 3.2 修改过程中

- 遵循项目的编码规范
- 保持代码的可读性和可维护性
- 及时保存修改

#### 3.3 修改后验证

- 执行 `wails dev` 进行编译测试
- 进行功能测试
- 更新相关文档

### 4. 常见问题处理

#### 4.1 编译失败

- 检查 Go 语法错误
- 检查 Vue 组件语法
- 检查依赖是否正确安装
- 查看错误日志定位问题

#### 4.2 前端构建失败

- 检查 `frontend/package.json` 依赖
- 运行 `npm install` 重新安装依赖
- 检查 Vite 配置文件

#### 4.3 应用启动失败

- 检查 Wails 配置文件
- 确认端口是否被占用
- 检查系统权限设置

### 5. 自动化建议

#### 5.1 开发环境启动

建议使用以下命令组合：

```bash
# 确保依赖已安装
npm install

# 启动开发服务器
wails dev
```

#### 5.2 生产构建测试

在重要修改后，建议进行生产构建测试：

```bash
wails build
```

### 6. 组件重构规则

#### 6.1 组件拆分原则

当组件变得复杂时，应遵循以下拆分原则：

- **单一职责**：每个组件只负责一个明确的功能
- **可复用性**：提取可复用的逻辑到独立组件
- **可维护性**：将复杂逻辑分解为更小的、易于管理的组件

#### 6.2 消息组件重构规范

在重构对话相关组件时，必须遵循以下规范：

- **用户消息组件**：使用 `MessageSend` 组件处理 `message.role === 'user'` 的消息
- **AI 回复组件**：使用 `MessageReply` 组件处理 `message.role === 'assistant'` 的消息
- **消息切换**：使用 `MessageSwitch` 组件处理多条消息的切换功能
- **配置管理**：在 `ConversationSettings` 中添加 `currentSendId` 和 `currentReplyId` 字段

#### 6.3 组件通信规范

- **Props 设计**：组件应接收 `messages` 数组和 `currentIndex` 索引，而不是单个 `message` 对象
- **事件传递**：使用 `emit` 向父组件传递操作事件和索引变化事件
- **状态管理**：当前显示的消息索引状态由父组件管理，子组件负责展示和交互

#### 6.4 重构后验证

组件重构完成后必须验证：

- [ ] 原有功能保持完整
- [ ] 消息显示正常
- [ ] 切换功能工作正常
- [ ] 操作按钮响应正确
- [ ] 配置保存和读取正常

### 7. Wails API 导入规范

#### 7.1 导入规则

在 Wails 项目中，所有后端 API 的导入必须遵循以下规范：

1. **统一导入位置**：所有 Wails 后端 API 必须在文件顶部的 `import` 语句中导入
2. **禁止动态导入**：禁止在函数或方法内部使用 `require()` 动态导入 Wails API
3. **导入路径规范**：使用相对路径 `../../../wailsjs/go/main/App` 导入后端 API

#### 7.2 正确示例

```javascript
// ✅ 正确：在顶部统一导入
import { GetConversations, CreateConversation, CreateMessage, UpdateConversationSettings } from "../../../wailsjs/go/main/App";

// 在方法中直接使用
const handleConfigChange = async (data) => {
  try {
    await UpdateConversationSettings(data.conversationId, data.settings);
    // ... 其他逻辑
  } catch (error) {
    // ... 错误处理
  }
};
```

#### 7.3 错误示例

```javascript
// ❌ 错误：在函数内部动态导入
const handleConfigChange = async (data) => {
  try {
    const { UpdateConversationSettings } = require("../../../wailsjs/go/main/App");
    await UpdateConversationSettings(data.conversationId, data.settings);
    // ... 其他逻辑
  } catch (error) {
    // ... 错误处理
  }
};
```

#### 7.4 规则说明

- **性能优化**：统一导入可以减少运行时的动态解析开销
- **代码可读性**：所有依赖关系在文件顶部一目了然
- **构建优化**：静态导入有利于构建工具进行 tree-shaking
- **错误排查**：统一的导入方式便于代码审查和问题定位

### 8. 数据库操作规范

#### 8.1 GORM 使用规则

在项目中进行数据库操作时，必须遵循以下规范：

1. **统一使用 GORM**：所有数据库查询和操作必须使用 GORM 框架
2. **全局数据库对象**：使用 `global.SLDB` 作为全局数据库连接对象
3. **错误处理**：所有数据库操作必须包含错误处理和日志记录
4. **事务管理**：复杂操作应使用事务确保数据一致性

#### 8.2 数据库连接对象

在 `global/global.go` 中定义的全局数据库对象：

```go
// 全局数据库连接对象
var SLDB *gorm.DB
```

#### 8.3 正确示例

```go
// ✅ 正确：使用全局 GORM 对象进行数据库操作
func (s *ConversationService) UpdateConversationSettings(id, settings string) error {
    if err := global.SLDB.Model(&models.Conversation{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "settings":   settings,
            "updated_at": time.Now(),
        }).Error; err != nil {
        return fmt.Errorf("failed to update conversation settings: %w", err)
    }
    return nil
}

// ✅ 正确：查询操作
func (s *ConversationService) GetConversationByID(id string) (*models.Conversation, error) {
    var conversation models.Conversation
    err := global.SLDB.Where("id = ?", id).First(&conversation).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("conversation not found")
        }
        return nil, fmt.Errorf("failed to query conversation: %w", err)
    }
    return &conversation, nil
}
```

#### 8.4 错误示例

```go
// ❌ 错误：不使用全局 GORM 对象
func (s *ConversationService) UpdateConversationSettings(id, settings string) error {
    db, err := gorm.Open("sqlite", "database.db")
    if err != nil {
        return err
    }
    // ... 错误的数据库操作
}

// ❌ 错误：缺少错误处理
func (s *ConversationService) GetConversationByID(id string) *models.Conversation {
    var conversation models.Conversation
    global.SLDB.Where("id = ?", id).First(&conversation) // 缺少错误处理
    return &conversation
}
```

#### 8.5 规则说明

- **统一性**：使用全局数据库对象确保连接池的一致性
- **性能优化**：避免重复创建数据库连接
- **错误追踪**：统一的错误处理便于问题定位
- **维护性**：集中的数据库配置便于管理

### 9. 后端 API 开发规范

#### 9.1 API 方法添加流程

在 Wails 项目中添加新的后端 API 方法时，必须遵循以下层次结构：

1. **Model 层**：在 `models/` 目录下定义数据结构体（如需要）
2. **Service 层**：在 `service/smart_query/` 目录下实现业务逻辑
3. **API 层**：在 `api/smart_query/` 目录下定义 API 接口
4. **App Core 层**：在 `appcore/app_api.go` 中注册 Wails API 方法

#### 9.2 各层职责说明

##### Model 层 (`models/`)

- 定义数据结构和数据库模型
- 包含 GORM 标签和 JSON 标签
- 定义数据验证规则

##### Service 层 (`service/smart_query/`)

- 实现具体的业务逻辑
- 处理数据库操作（使用 `global.SLDB`）
- 包含错误处理和数据验证
- 使用 GORM 进行所有数据库操作
- **必须添加日志记录**：所有关键操作都必须使用 `logger` 包记录日志

##### API 层 (`api/smart_query/`)

- 定义 HTTP 风格的 API 接口
- 处理参数转换和类型断言
- 调用 Service 层方法
- 返回标准化的响应格式

##### App Core 层 (`appcore/app_api.go`)

- 注册 Wails API 方法，供前端调用
- 提供统一的 API 入口
- 处理应用级别的逻辑

#### 9.3 Service 层日志记录规范

##### 9.3.1 日志记录要求

所有 Service 层的关键操作都必须添加日志记录，包括：

1. **创建操作**：记录创建的实体信息和相关参数
2. **更新操作**：记录更新的实体ID和变更内容
3. **删除操作**：记录删除的实体ID和相关上下文
4. **查询操作**：记录查询失败的情况
5. **错误处理**：记录所有错误的详细信息和上下文

##### 9.3.2 日志记录示例

```go
// 创建操作日志
logger.LogInfo("Creating new conversation", map[string]interface{}{
    "topicId": conversation.TopicID,
    "title":   conversation.Title,
})

// 错误日志
logger.LogError("Failed to create conversation", err, map[string]interface{}{
    "topicId": conversation.TopicID,
    "title":   conversation.Title,
})

// 数据库操作日志
logger.LogDatabaseOperation("create", "conversations", conversation.ID, nil)
```

##### 9.3.3 必须记录的操作

Service 层的所有数据库操作和外部请求都必须进行日志记录，包括：

- **创建操作**：所有实体的创建方法（Create*）
- **更新操作**：所有实体的更新方法（Update*）
- **删除操作**：所有实体的删除方法（Delete*）
- **外部请求**：向大模型发送请求的操作（如 StreamChatCompletion）
- **批量操作**：批量更新或删除的操作
- **事务操作**：涉及多个数据库操作的复杂业务逻辑

#### 9.4 添加新 API 方法的完整流程

假设要添加 `UpdateConversationSettings` 方法：

1. **Service 层实现** (`service/smart_query/conversation.go`)：

```go
// UpdateConversationSettings 更新对话设置
func (s *ConversationService) UpdateConversationSettings(id, settings string) error {
    logger.LogInfo("Updating conversation settings", map[string]interface{}{
        "conversationId": id,
        "settingsLength": len(settings),
    })
    
    if err := global.SLDB.Model(&models.Conversation{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "settings":   settings,
            "updated_at": time.Now(),
        }).Error; err != nil {
        logger.LogError("Failed to update conversation settings", err, map[string]interface{}{
            "conversationId": id,
        })
        return fmt.Errorf("failed to update conversation settings: %w", err)
    }
    
    logger.LogDatabaseOperation("update", "conversations", id, nil)
    return nil
}
```

2. **API 层定义** (`api/smart_query/conversation.go`)：

```go
// UpdateConversationSettings 更新对话设置
func (api *ConversationAPI) UpdateConversationSettings(id, settings string) error {
    return conversationService.UpdateConversationSettings(id, settings)
}
```

3. **App Core 层注册** (`appcore/app_api.go`)：

```go
// UpdateConversationSettings 更新对话设置 - Wails API方法
func UpdateConversationSettings(a *App, id, settings string) error {
    return a.SmartQueryAPI.ConversationAPI.UpdateConversationSettings(id, settings)
}
```

4. **根目录 App 层委托** (`app.go`)：

```go
// UpdateConversationSettings 更新对话设置 - 委托给appcore包
func (a *App) UpdateConversationSettings(id, settings string) error {
    return appcore.UpdateConversationSettings(a.App, id, settings)
}
```

5. **前端使用** (`frontend/src/views/SmartQuery/chat.vue`)：

```javascript
import { UpdateConversationSettings } from "../../../wailsjs/go/main/App";

const handleConfigChange = async (data) => {
  try {
    await UpdateConversationSettings(data.conversationId, data.settings);
    // ... 其他逻辑
  } catch (error) {
    // ... 错误处理
  }
};
```

#### 9.5 重要说明

- **必须按层次顺序**：Service → API → App Core → 根目录 App
- **根目录 App 层**：所有 Wails API 方法必须在根目录的 `app.go` 中有对应的委托方法
- **绑定生成**：只有在根目录 `app.go` 中定义的方法才会被 Wails 生成前端绑定
- **重新构建**：添加新方法后必须重新运行 `wails build` 或 `wails dev` 来重新生成绑定文件
- **日志记录**：Service 层的所有关键操作都必须添加适当的日志记录

#### 9.6 注意事项

- **必须按层次顺序**：Service → API → App Core
- **统一错误处理**：所有层都要有适当的错误处理
- **使用全局数据库对象**：Service 层必须使用 `global.SLDB`
- **参数验证**：API 层负责参数转换和基本验证
- **文档同步**：添加新方法后更新相关文档
- **日志记录**：Service 层必须包含完整的日志记录，便于问题排查

### 10. 组件定义规则

#### 10.1 组件 Props 处理规范

如果一个组件拥有 Array 或 Object 类型的 Props，或某一个 Props 参与 Computed 计算属性的计算，则必须按照以下方式来定义：

```javascript
// 初始化方法
const init = () => {
  // 初始化逻辑
};

// 组件初始化时调用init
init();
```

#### 10.2 中间值定义规则

- **命名规范**：中间值的名称为"props 的名称+Data"，驼峰命名法
- **深拷贝方式**：使用 cloneDeep()的方式深拷贝来接收 Props
- **监听变化**：对 init 中采用中间值接收的 Props 进行监听变化

#### 10.3 正确示例

```javascript
import { cloneDeep } from "lodash";
// Props
const props = defineProps({
  messages: {
    type: Array,
    required: true,
    default: () => [],
  },
  conversation: {
    type: Object,
    default: null,
  },
});

// 定义中间值来接收props
const messagesData = ref([]);
const conversationData = ref({});

// 初始化方法
const init = () => {
  messagesData.value = cloneDeep(props.messages);
  conversationData.value = cloneDeep(props.conversation);
};

// 组件初始化时调用init
init();

// 计算属性使用中间值
const currentIndex = computed(() => {
  if (!props.currentId || messagesData.value.length === 0) return 0;
  const index = messagesData.value.findIndex((msg) => msg.id === props.currentId);
  return index !== -1 ? index : 0;
});

// 监听props变化，更新中间值
watch(
  () => props.messages,
  (newMessages) => {
    messagesData.value = cloneDeep(newMessages);
  },
  { deep: true }
);

watch(
  () => props.conversation,
  (newConversation) => {
    conversationData.value = cloneDeep(newConversation);
  },
  { deep: true }
);
```

#### 10.4 规则说明

- **避免直接修改 props**：Vue 3 中无法直接修改 props，需要通过中间值进行操作
- **响应式保证**：使用 ref 定义中间值，确保响应式更新
- **数据隔离**：深拷贝确保 props 和中间值的数据隔离
- **变化监听**：监听 props 变化，及时更新中间值

#### 10.5 适用场景

- **Array 类型 Props**：当 props 包含数组，且需要在组件内部修改或计算时
- **Object 类型 Props**：当 props 包含对象，且需要在组件内部修改或计算时
- **Computed 依赖**：当 props 参与 computed 计算属性的计算时
- **复杂操作**：当需要对 props 进行复杂操作（过滤、排序、映射等）时

### 11. 按钮内容布局规则

#### 11.1 按钮内容容器规范

当按钮需要包含图标和文本时，必须使用 `class="button-content"` 作为内容容器的统一类名。

#### 11.2 样式规则

```css
.button-content {
  display: flex;
  align-items: center;
  gap: 4px;
}
```

#### 11.3 正确示例

```vue
<template>
  <t-button variant="text" @click="handleAction">
    <div class="button-content">
      <IconPlus class="add-icon-margin" :size="16" />
      <span>新增助手</span>
    </div>
  </t-button>
</template>

<style lang="less" scoped>
.button-content {
  display: flex;
  align-items: center;
  gap: 4px;
}

.add-icon-margin {
  margin-right: 4px;
}
</style>
```

#### 11.4 规则说明

- **统一布局**：所有按钮内容都使用相同的布局类名，确保视觉一致性
- **Flexbox 布局**：使用 flex 布局实现图标和文本的对齐
- **间距控制**：使用 gap 属性控制图标和文本之间的间距
- **图标边距**：图标可以单独设置 margin 来调整与文本的间距

#### 11.5 适用场景

- **图标+文本按钮**：当按钮包含图标和文本时
- **多元素按钮**：当按钮包含多个元素需要统一布局时
- **一致性要求**：需要保持整个应用的按钮样式一致性时

### 12. 文档维护

#### 12.1 规则文件更新

当开发流程发生变化时，需要更新此规则文件。

#### 12.2 项目结构同步

确保 `.clinerules/project-structure.md` 始终反映最新的项目结构。

## 执行顺序

1. **修改代码** → 2. **运行 `wails dev`** → 3. **验证功能** → 4. **更新文档** → 5. **完成**

## 注意事项

1. **必须执行**：每次代码修改后都必须运行 `wails dev` 进行测试
2. **及时更新**：创建新文件后必须立即更新项目结构文档
3. **完整验证**：不仅要验证编译成功，还要验证功能正常
4. **文档同步**：保持代码和文档的同步更新
5. **错误处理**：遇到编译错误时，必须解决后才能继续开发

## 违规后果

不遵循此规则可能导致：

- 代码质量下降
- 项目文档过时
- 团队协作困难
- 应用运行时错误
- 维护成本增加
