# Dify 文件上传和聊天请求流程说明

## 一、整体流程

```
用户上传文件 → 前端处理 → 后端上传到Dify → 获取Dify文件ID → 用户发送消息 → 构建消息（包含文件ID） → 发送到Dify API
```

## 二、前端流程

### 1. 文件上传阶段 (`handleUploadFile`)

**步骤：**
1. 用户选择文件
2. 前端处理文件（压缩图片等）
3. 将文件转换为 base64
4. 调用 `SaveFile` API 保存文件到后端

**文件对象结构：**
```javascript
{
  key: processedFile.uuid,        // 初始UUID（临时key）
  name: files[0].name,             // 文件名
  status: 'progress',             // 上传状态：'progress' | 'success' | 'error'
  description: '上传中',
  // ... 其他字段
}
```

**SaveFile API 调用：**
```javascript
const savedFile = await SaveFile(
  fileNameWithoutExt,      // 文件名（不含后缀）
  files[0].name,           // 原始文件名（含后缀）
  processedFile.fileSuffix, // 文件后缀
  processedFile.md5,       // MD5值
  processedFile.originalPath || '', // 本地路径（可为空）
  processedFile.size,       // 文件大小
  newConversationId.value || 'temp', // 关联ID
  fileContentBase64        // 文件内容（base64编码）
);
```

**后端返回：**
```javascript
{
  id: "本地UUID",           // 本地数据库文件ID
  originalPath: "Dify文件ID" // Dify返回的文件ID（UUID格式）
}
```

**更新文件状态：**
```javascript
// 将文件key更新为Dify文件ID
filesList.value = filesList.value.map((file) => {
  if (file.key === oldKey) {
    return {
      ...file,
      key: difyFileID,      // 更新为Dify文件ID
      localId: fileId,      // 保存本地UUID
      status: 'success',    // 更新状态为成功
    };
  }
  return file;
});
```

### 2. 构建消息阶段 (`buildChatMessages`)

**从当前文件列表提取文件ID：**
```javascript
const currentFiles = filesList.value
  .filter((f) => f?.status === 'success')  // 只包含成功上传的文件
  .map((f) => f?.key || f?.id)             // 提取Dify文件ID
  .filter((id) => typeof id === "string" && id.trim() !== "");
```

**构建用户消息：**
```javascript
const userMessage = {
  role: "user",
  content: inputValue,                    // 用户输入的问题
  files: currentFiles.length > 0 ? currentFiles : undefined  // 文件ID数组
};
```

**完整消息列表格式：**
```javascript
[
  {
    role: "system",
    content: "系统提示词..."
  },
  {
    role: "user",
    content: "历史消息1",
    files: ["dify-file-id-1"]  // 历史消息的附件（如果有）
  },
  {
    role: "assistant",
    content: "AI回复1"
  },
  {
    role: "user",
    content: "帮我分析这个文件",  // 当前用户输入
    files: ["c60d398b-d452-4c3a-8a58-0c5f6c8f0f7c"]  // 当前文件的Dify ID
  }
]
```

### 3. 发送消息阶段 (`startChatStream`)

**调用后端API：**
```javascript
await StreamChatCompletion(
  streamId,           // 流式输出ID
  StreamType.CHAT,    // 流类型
  conversationId,     // 对话ID
  messages,           // 消息列表（包含files字段）
  modelType           // 模型类型：'instruct' | 'thinking' | 'fast'
);
```

## 三、后端流程

### 1. 接收消息 (`StreamChatCompletion`)

**从消息列表中提取最后一条用户消息：**
```go
for i := len(messages) - 1; i >= 0; i-- {
    if role, ok := messages[i]["role"].(string); ok && role == "user" {
        // 提取用户问题
        query = content
        
        // 提取文件ID数组
        if files, ok := messages[i]["files"].([]interface{}); ok {
            for _, f := range files {
                if id, ok := f.(string); ok && strings.TrimSpace(id) != "" {
                    fileIDs[strings.TrimSpace(id)] = struct{}{}
                }
            }
        }
        break
    }
}
```

### 2. 构建 Dify API 请求

**请求数据结构：**
```go
requestData := map[string]interface{}{
    "inputs":        map[string]interface{}{},  // Dify输入变量（通常为空）
    "query":         query,                      // 用户问题
    "response_mode": "streaming",                // 流式响应
    "user":          streamID,                   // 用户标识
}

// 如果有文件ID，添加files数组
if len(fileIDs) > 0 {
    files := make([]string, 0, len(fileIDs))
    for id := range fileIDs {
        files = append(files, id)
    }
    requestData["files"] = files
}

// 如果有有效的conversation_id（UUID格式），添加它
if conversationID != "" {
    requestData["conversation_id"] = conversationID
}
```

**最终请求JSON格式：**
```json
{
  "inputs": {},
  "query": "帮我分析这个文件",
  "response_mode": "streaming",
  "user": "39200147915008",
  "files": [
    "c60d398b-d452-4c3a-8a58-0c5f6c8f0f7c"
  ],
  "conversation_id": "46368b5b-7bd5-4407-80ce-7aa3153437fc"  // 可选，必须是UUID格式
}
```

### 3. 发送HTTP请求

**请求配置：**
```go
POST http://192.168.100.39/v1/chat-messages
Content-Type: application/json
Authorization: Bearer app-ggympSzmvPpq9e4oGWWCxQ5q

{
  "inputs": {},
  "query": "帮我分析这个文件",
  "response_mode": "streaming",
  "user": "39200147915008",
  "files": ["c60d398b-d452-4c3a-8a58-0c5f6c8f0f7c"]
}
```

## 四、关键点说明

### 1. 文件ID的转换流程

```
前端临时UUID → 后端上传到Dify → Dify返回文件ID（UUID） → 更新filesList中的key → 构建消息时使用Dify文件ID
```

### 2. 文件状态管理

- `progress`: 文件正在上传中
- `success`: 文件上传成功，可以使用
- `error`: 文件上传失败

**只有 `status === 'success'` 的文件才会被包含在消息中发送给Dify。**

### 3. 消息构建时机

- `filesList.value` 必须在 `buildChatMessages` **之前**保持有效
- 在 `startChatStream` **之后**才清空 `filesList.value`

### 4. Dify API 要求

- `files` 字段必须是字符串数组，包含Dify文件ID（UUID格式）
- `conversation_id` 必须是有效的UUID格式，否则留空让Dify自动创建
- `query` 是用户的问题文本
- `response_mode` 设置为 `"streaming"` 用于流式响应

## 五、调试日志

### 前端日志
- `🔍 buildChatMessages - 文件列表状态:` - 显示文件列表和文件ID
- `🔍 buildChatMessages - 用户消息:` - 显示构建的用户消息（包含files字段）
- `🔍 buildChatMessages - 最终消息列表:` - 显示完整的消息列表

### 后端日志
- `找到用户消息中的files字段` - 成功提取文件ID
- `收集到文件ID` - 每个文件ID的提取
- `附加文件ID到Dify请求` - 文件ID已添加到请求中
- `用户消息中没有files字段或格式不正确` - 未找到文件ID（需要检查）

## 六、常见问题

### 1. 文件ID没有传递到Dify

**可能原因：**
- 文件状态没有正确更新为 `success`
- `filesList.value` 在 `buildChatMessages` 之前被清空
- 文件ID格式不正确

**解决方法：**
- 检查浏览器控制台日志，确认文件状态
- 确认 `buildChatMessages` 时 `filesList.value` 不为空
- 检查文件ID是否为有效的UUID格式

### 2. 文件上传成功但前端仍显示"上传中"

**可能原因：**
- 文件状态更新逻辑中的 `oldKey` 匹配失败
- `filesList.value` 更新没有触发响应式更新

**解决方法：**
- 检查浏览器控制台日志，确认是否找到匹配的文件
- 确认 `oldKey` 和 `filesList` 中的 `key` 是否一致

