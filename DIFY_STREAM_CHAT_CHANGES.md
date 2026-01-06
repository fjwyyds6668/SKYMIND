# Dify 对话流前后端改造说明

本文件汇总当前项目从「本地/OpenAI 兼容接口」改造为「完全通过 Dify 对话流 + 文件上传」所做的前后端改动，便于以后维护或再次迁移。

---

## 一、整体架构变化

- **模型调用入口不变**：前端仍通过 Wails 绑定调用 `StreamChatCompletion`，但后端实现从原来的 OpenAI 兼容接口改为 Dify 的 `/v1/chat-messages`。
- **文件上传链路**：前端上传 → Go 后端保存临时文件 → 调用 Dify `/v1/files/upload` → 保存 Dify 文件 ID 到本地数据库 → 对话时携带到 Dify。
- **会话管理**：继续使用本地 `topics / conversations / messages` 表，`conversation_id` 只在 Dify 需要 UUID 时才传入，否则让 Dify 自动创建。
- **所有提示词与标题生成**：`GenerateSystemPrompt / OptimizeUserPrompt / GenerateConversationTitle / GenerateTopicTitle` 全部改为通过 Dify 的 blocking 接口实现，不再本地拼提示词。

---

## 二、后端改动（Go）

### 1. `service/smart_query/generator.go`

- **核心函数 `StreamChatCompletion` 改造为调用 Dify：**
  - 从 `database` 读取对应模型配置，使用 `api_base=http://192.168.100.39/v1`、`api_key=app-...`。
  - 从 `messages []map[string]interface{}` 中：
    - 找到最后一条 `role == "user"` 的消息，取其 `content` 作为 `query`。
    - 读取该消息的 `files` 字段（前端传入的 Dify 文件 ID 数组）。
  - 对 `relatedID` 做 UUID 校验：
    - 若是有效 UUID，则作为 `conversation_id` 传给 Dify。
    - 否则不传，让 Dify 自动创建新会话并在响应中返回。

- **Dify 请求体构造：**

  ```go
  requestData := map[string]interface{}{
    "inputs":        map[string]interface{}{},
    "query":         query,
    "response_mode": "streaming",
    "user":          streamID,
    // 可选: "conversation_id": <uuid>,
    // 可选: "files": []map[string]interface{}{...},
  }
  ```

- **`files` 字段改为 Dify 要求的对象数组：**
  - 先把最后一条用户消息中的 `files` 解析成 `fileIDs` 集合（Dify 文件 UUID）。
  - 根据 Dify 文件 ID 到本地 `files` 表中查询记录（`original_path` 存 Dify 文件 ID）。
  - 根据 `FileSuffix` 判定类型：
    - 图片扩展名（`png/jpg/jpeg/gif/webp/bmp/svg/ico`）→ `type = "image"`。
    - 其他默认 `type = "document"`。
  - 组装：

    ```go
    fileObj := map[string]interface{}{
      "type":            fileType,       // "image" or "document"
      "transfer_method": "local_file",
      "upload_file_id":  difyFileID,
    }
    ```

  - 最终：

    ```go
    requestData["files"] = []map[string]interface{}{ fileObj, ... }
    ```

- **流式请求发送 `sendStreamRequestWithCancel`：**
  - 自动处理 `api_base` 是否已包含 `/v1`。
  - 超时时间调整为 **5 分钟**，适配长流式对话。

- **流式响应处理 `handleStreamResponseWithWebSocket`：**
  - 处理 Dify SSE 格式：逐行读取 `data: ...`。
  - 新增：
    - 使用 `bufio.Scanner` 并调用 `scanner.Buffer(buf, 1MB)`，解决 `token too long` 错误。
    - 识别 Dify 的结束事件：`message_end`、`workflow_finished`，收到后发出 `"websocket-stream-end"` 事件。
    - 将 Dify 的事件数据通过 `convertDifyResponseToOpenAIFormat` 转为前端已使用的 OpenAI 风格增量结构（`choices[].delta.content`），保持前端兼容。

- **辅助 Dify 调用统一封装 `callDifyAPI`：**
  - 用于：
    - `GenerateSystemPrompt`
    - `OptimizeUserPrompt`
    - `GenerateConversationTitle`
    - `GenerateTopicTitle`
  - 使用 `response_mode: "blocking"` 单次返回，内部从 `answer` 或 `choices[].message.content` 中取文本。

- **其它：**
  - 新增 `getDifyFileType`：根据扩展名返回 `"image"` 或 `"document"`。
  - 保留 `isValidUUID` 用于 `conversation_id` 判断。

### 2. `service/smart_query/file.go`

- **上传文件到 Dify：`uploadFileToDify`**
  - 使用 `multipart/form-data` 请求 `POST /v1/files/upload`：
    - 字段 `file`：本地临时文件。
    - 字段 `user`：当前用户/会话标识（统一用本地生成的 UUID）。
  - 接受 `200 OK` 和 `201 Created` 为成功状态。
  - 解析响应，取 `id` 作为 Dify 文件 ID。

- **保存文件元数据：`SaveFileMetadata`**
  - 若有本地路径或 base64 内容，先保存/生成临时文件。
  - 调用 `uploadFileToDify` 获取 Dify 文件 ID。
  - 在本地 `files` 表中：
    - `ID`：本地 UUID。
    - `OriginalPath`：**保存 Dify 文件 ID（uuid）**。
    - `FileSuffix / FileSize / RelatedID / Content` 等照旧。

- **支持 base64 上传：`SaveBase64ToTempFile`**
  - 将前端传来的 base64 内容还原成文件，存储到用户目录下的 `skymind/temp` 目录。
  - 供 `uploadFileToDify` 使用。

- **内容处理：`ProcessFileContent`**
  - 扩展 `isTextFile` 支持 `pdf` 等。
  - 对 pdf 做占位处理（后续可接入真正的 PDF 文本抽取）。

### 3. 数据库与模型

- **`models/file.go`**
  - `File` 模型增加 `Content string` 字段，保存解析后的 Markdown 内容。

- **`database/gorm.go`**
  - `autoMigrate` 中新增 `&models.File{}`，确保 `files` 表创建并包含新字段。

### 4. 配置相关

- **`database/config.go` + `appcore/model_config.yml`**
  - 默认与用户配置中均将模型 `api_base`/`api_key` 指向 Dify：
    - `api_base`: `http://192.168.100.39/v1`
    - `api_key`: `app-...`

---

## 三、前端改动（Vue / Wails）

### 1. `frontend/src/views/SmartQuery/chat.vue` – 聊天主界面

#### 1.1 文件上传改造

- **上传入口**：继续使用 `t-chat-sender` 的 `@file-select="handleUploadFile"`。
- **`handleUploadFile` 逻辑：**
  1. 调用 `processFile` 对文件进行预处理（压缩等）。
  2. 在 `filesList` 中插入一条状态为 `progress` 的文件记录：
     - `key`: 本地 UUID。
     - `status`: `'progress'`。
  3. 使用 `FileReader.readAsDataURL` 将处理后的文件转为 base64 字符串。
  4. 调用 Wails 绑定的 `SaveFile(...)`，传入：
     - `fileName`（不含后缀）、
     - `originalName`、
     - `fileSuffix`、
     - `md5`、
     - `localPath`（可以为空）、
     - `fileSize`、
     - `relatedID`（对话或临时 ID）、
     - `fileContentBase64`。
  5. 解析返回的本地 `id` 和 `originalPath`（Dify 文件 ID），按 `oldKey` 匹配更新：
     - `key` 改为 Dify 文件 ID。
     - 保存 `localId` = 本地 UUID。
     - `status` 设为 `'success'`。
  6. 调用 `ProcessFileContent(fileId)` 处理文件内容。

- **上传状态与按钮禁用：**
  - 计算属性 `hasUploadingFiles`：`filesList` 中存在 `status === 'progress'` 的文件。
  - `t-chat-sender`：
    - `:loading="isStreamingChat || isOptimizingPrompt || hasUploadingFiles"`。
    - `:disabled="isStreamingChat || !selectedAssistantData || !selectedTopicData || hasUploadingFiles"`。
  - 确保文件未上传完时无法发送消息，且输入框有 loading 态。

#### 1.2 构建消息并携带文件 ID

- **`buildChatMessages(inputValue)`：**
  - 遍历历史对话中的 `messages`：
    - 对 `msg.attachments` 提取 `key`/`id` 作为历史文件 ID。
    - 仅保留 `status === 'success'` 或无 `status` 的附件（兼容旧数据）。
    - 生成：

      ```js
      {
        role: msg.role,
        content: msg.content,
        files: msgFiles.length > 0 ? msgFiles : undefined,
      }
      ```

  - 从当前 `filesList` 中提取本次要发送的文件 ID：

    ```js
    const currentFiles = filesList.value
      .filter(f => f?.status === 'success')
      .map(f => f?.key || f?.id)
      .filter(id => typeof id === "string" && id.trim() !== "");
    ```

  - 在末尾追加当前用户消息：

    ```js
    {
      role: "user",
      content: inputValue,
      files: currentFiles.length > 0 ? currentFiles : undefined,
    }
    ```

- **发送流式请求 `startChatStream`：**
  - 调用：

    ```js
    await StreamChatCompletion(
      streamId,
      StreamType.CHAT,
      conversationId,
      messages,   // 内含 files 数组
      modelType
    );
    ```

  - 在流启动后再清空 `filesList.value`，避免丢失待发送文件 ID。

#### 1.3 其它 UI 与逻辑调整

- **布局修复：**

  - 使用 `.input-section` 与 `:deep(.t-chat-sender)` 等选择器，保证附件按钮不溢出、输入区宽度 100%、滚动正常。

- **文件删除 / 点击：**
  - `handleRemoveFile`：从 `filesList` 中删除对应 `key` 的文件。
  - `handleFileClick`：预留文件点击逻辑（可扩展为预览等）。

### 2. 设定与流管理

- **`frontend/src/views/SmartQuery/settings.vue`**
  - `handleGeneratePrompt` 改为直接调用非流式的 `GenerateSystemPrompt`（经 Dify blocking API），不再通过 `StreamChatCompletion` 生成系统提示词。

- **`frontend/src/store/modules/stream.js`**
  - 会话标题与主题标题生成：
    - 改为先调用 `GenerateConversationTitle` / `GenerateTopicTitle`（Dify blocking），拿到结果后再调用后端 `UpdateConversationTitle` / `UpdateTopicTitle`。
  - 保持聊天主流仍为 SSE 流式。

---

## 四、行为验证与调试要点

- **上传成功 & Dify 成功接收：**
  - 后端日志应出现：
    - `文件上传到 Dify 并保存元数据成功`（含 `difyFileID`）。
    - `找到用户消息中的files字段`。
    - `收集到文件ID`。
    - `附加文件ID到Dify请求`，并显示形如：

      ```text
      fileCount=1 files="[map[transfer_method:local_file type:document upload_file_id:...]]"
      ```

  - Dify workflow 的 `sys.files` 中应能看到该文件。

- **流式响应：**

  - 若 Dify 返回内容较长，`bufio.Scanner: token too long` 已通过扩大缓冲区至 1MB 解决。

- **常见问题排查：**
  - `sys.files` 为空：
    - 检查前端是否把状态更新为 `success`。
    - 检查 `buildChatMessages` 是否真正把 `files` 字段带进最后一条用户消息。
    - 检查后端日志是否有 `找到用户消息中的files字段`。
  - 前端一直显示“上传中”：
    - 看控制台日志：确认 `oldKey` 匹配正确、`filesList` 的 `status` 是否已改为 `success`。
  - Dify 返回 500：
    - 现在请求体已与 Dify 文档一致，需到 Dify 控制台查看 workflow 内部报错。

---

> 本文档对应代码改动主要集中在：`service/smart_query/generator.go`、`service/smart_query/file.go`、`frontend/src/views/SmartQuery/chat.vue`、`frontend/src/views/SmartQuery/settings.vue`、`frontend/src/store/modules/stream.js`、`database/config.go`、`models/file.go`、`database/gorm.go` 等文件，可结合 Git 历史进一步查看具体 diff。

