package smart_query

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"skymind/database"
	"skymind/global"
	"skymind/logger"
	"skymind/models"

	"github.com/google/uuid"
)

type FileService struct{}

// SaveFileMetadata 保存文件元数据到数据库（先上传到 Dify，再保存元数据）
func (s *FileService) SaveFileMetadata(fileRecord *models.File) (*models.File, error) {
	// 生成UUID如果为空
	if fileRecord.ID == "" {
		fileUUID := uuid.New().String()
		fileRecord.ID = fileUUID
	}

	// 先上传文件到 Dify
	difyFileID, err := s.uploadFileToDify(fileRecord)
	if err != nil {
		logger.LogError("上传文件到 Dify 失败", err, map[string]interface{}{
			"fileUUID":     fileRecord.ID,
			"originalName": fileRecord.OriginalName,
			"localPath":    fileRecord.OriginalPath,
		})
		return nil, fmt.Errorf("上传文件到 Dify 失败: %w", err)
	}

	// 将 Dify 文件 ID 保存到 OriginalPath 字段（或可以添加新字段）
	// 这里我们使用 OriginalPath 存储 Dify 文件 ID，因为本地路径可能不再需要
	fileRecord.OriginalPath = difyFileID

	// 保存到数据库
	if err := global.SLDB.Create(fileRecord).Error; err != nil {
		logger.LogError("保存文件元数据到数据库失败", err, map[string]interface{}{
			"fileUUID":     fileRecord.ID,
			"originalName": fileRecord.OriginalName,
			"difyFileID":   difyFileID,
		})
		return nil, fmt.Errorf("保存文件元数据到数据库失败: %w", err)
	}

	logger.LogInfo("文件上传到 Dify 并保存元数据成功", map[string]interface{}{
		"fileUUID":     fileRecord.ID,
		"originalName": fileRecord.OriginalName,
		"fileSize":     fileRecord.FileSize,
		"relatedID":    fileRecord.RelatedID,
		"difyFileID":   difyFileID,
	})

	logger.LogDatabaseOperation("create", "files", fileRecord.ID, nil)

	return fileRecord, nil
}

// uploadFileToDify 上传文件到 Dify
func (s *FileService) uploadFileToDify(fileRecord *models.File) (string, error) {
	// 读取本地文件
	fileData, err := os.ReadFile(fileRecord.OriginalPath)
	if err != nil {
		return "", fmt.Errorf("读取本地文件失败: %w", err)
	}

	// 获取 Dify 配置（使用 instruct model 配置）
	config := database.GetInstructModelConfig()
	apiBase := config.ApiBase
	apiKey := config.ApiKey

	// 构建 Dify API 端点
	// 如果 apiBase 已经包含 /v1，则不重复添加
	apiEndpoint := "/files/upload"
	if strings.HasSuffix(apiBase, "/v1") || strings.HasSuffix(apiBase, "/v1/") {
		apiEndpoint = "/files/upload"
	} else {
		apiEndpoint = "/v1/files/upload"
	}

	// 创建 multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件字段（确保文件名格式正确，避免重复扩展名）
	fileName := fileRecord.OriginalName
	if !strings.HasSuffix(strings.ToLower(fileName), "."+strings.ToLower(fileRecord.FileSuffix)) {
		fileName = fileName + "." + fileRecord.FileSuffix
	}

	// 获取正确的 MIME 类型
	mimeType := mime.TypeByExtension("." + strings.ToLower(fileRecord.FileSuffix))
	if mimeType == "" {
		// 如果无法识别，根据常见扩展名设置默认值
		mimeType = s.getMimeTypeByExtension(fileRecord.FileSuffix)
	}

	// 手动创建 multipart 字段以便设置正确的 Content-Type
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	fileHeader.Set("Content-Type", mimeType)

	fileField, err := writer.CreatePart(fileHeader)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}

	if _, err := fileField.Write(fileData); err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}

	// 添加用户标识字段（使用 relatedID 或 fileRecord.ID）
	userID := fileRecord.RelatedID
	if userID == "" {
		userID = fileRecord.ID
	}
	if err := writer.WriteField("user", userID); err != nil {
		return "", fmt.Errorf("写入用户字段失败: %w", err)
	}

	// 关闭 writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭 multipart writer 失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiBase+apiEndpoint, &requestBody)
	if err != nil {
		return "", fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码（201 Created 或 200 OK 都表示成功）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Dify API 返回错误: status=%d, message=%s", resp.StatusCode, string(responseBody))
	}

	// 解析响应
	var difyResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &difyResponse); err != nil {
		return "", fmt.Errorf("解析 Dify 响应失败: %w", err)
	}

	// 提取文件 ID
	fileID, ok := difyResponse["id"].(string)
	if !ok {
		// 尝试其他可能的字段名
		if fileIDStr, ok := difyResponse["file_id"].(string); ok {
			fileID = fileIDStr
		} else {
			return "", fmt.Errorf("Dify 响应中未找到文件 ID: %v", difyResponse)
		}
	}

	logger.LogInfo("文件上传到 Dify 成功", map[string]interface{}{
		"difyFileID":   fileID,
		"originalName": fileRecord.OriginalName,
		"fileSize":     fileRecord.FileSize,
		"response":     difyResponse,
	})

	return fileID, nil
}

// SaveBase64ToTempFile 将 base64 编码的文件内容保存到临时文件
func (s *FileService) SaveBase64ToTempFile(base64Content, fileName string) (string, error) {
	// 解码 base64
	fileData, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", fmt.Errorf("解码 base64 失败: %w", err)
	}

	// 获取用户配置目录
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = os.Getenv("HOME")
	}
	tempDir := filepath.Join(userProfile, "skymind", "temp")

	// 确保临时目录存在
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 生成临时文件路径
	tempFilePath := filepath.Join(tempDir, fileName)

	// 保存文件
	if err := os.WriteFile(tempFilePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("保存临时文件失败: %w", err)
	}

	logger.LogInfo("临时文件保存成功", map[string]interface{}{
		"tempPath": tempFilePath,
		"fileName": fileName,
		"size":     len(fileData),
	})

	return tempFilePath, nil
}

// SaveFile 保存文件到本地并记录到数据库
func (s *FileService) SaveFile(file multipart.File, header *multipart.FileHeader, relatedID string) (*models.File, error) {
	defer file.Close()

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.LogError("读取文件内容失败", err, map[string]interface{}{
			"fileName":  header.Filename,
			"relatedID": relatedID,
		})
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	// 计算MD5
	md5Hash := md5.Sum(fileBytes)
	md5Str := fmt.Sprintf("%x", md5Hash)

	// 获取文件信息
	fileName := header.Filename
	fileSuffix := ""
	if dotIndex := strings.LastIndex(fileName, "."); dotIndex != -1 {
		fileSuffix = strings.ToLower(fileName[dotIndex+1:])
		fileName = fileName[:dotIndex]
	} else {
		// 如果没有找到点号，整个文件名作为fileName，fileSuffix为空
		fileSuffix = ""
	}

	// 生成UUID
	fileUUID := uuid.New().String()

	// 获取用户配置目录
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = os.Getenv("HOME")
	}
	cacheDir := filepath.Join(userProfile, "skymind", "cache")

	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		logger.LogError("创建缓存目录失败", err, map[string]interface{}{
			"cacheDir": cacheDir,
		})
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 构建保存路径
	saveFileName := fmt.Sprintf("%s.%s", fileUUID, fileSuffix)
	savePath := filepath.Join(cacheDir, saveFileName)

	// 保存文件到本地
	if err := os.WriteFile(savePath, fileBytes, 0644); err != nil {
		logger.LogError("保存文件失败", err, map[string]interface{}{
			"savePath": savePath,
			"fileName": header.Filename,
		})
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 创建文件记录
	fileRecord := &models.File{
		ID:           fileUUID,
		OriginalPath: header.Filename,
		OriginalMD5:  md5Str,
		OriginalName: fileName,
		FileSuffix:   fileSuffix,
		FileSize:     int64(len(fileBytes)),
		RelatedID:    relatedID,
	}

	// 保存到数据库
	if err := global.SLDB.Create(fileRecord).Error; err != nil {
		logger.LogError("保存文件记录到数据库失败", err, map[string]interface{}{
			"fileUUID": fileUUID,
			"fileName": header.Filename,
		})
		// 如果数据库保存失败，删除已保存的文件
		os.Remove(savePath)
		return nil, fmt.Errorf("保存文件记录到数据库失败: %w", err)
	}

	logger.LogInfo("文件保存成功", map[string]interface{}{
		"fileUUID":     fileUUID,
		"originalName": header.Filename,
		"fileSize":     len(fileBytes),
		"savePath":     savePath,
		"relatedID":    relatedID,
	})

	logger.LogDatabaseOperation("create", "files", fileUUID, nil)

	return fileRecord, nil
}

// GetFileByID 根据ID获取文件信息
func (s *FileService) GetFileByID(id string) (*models.File, error) {
	var file models.File
	err := global.SLDB.Where("id = ?", id).First(&file).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, fmt.Errorf("文件不存在")
		}
		logger.LogError("查询文件失败", err, map[string]interface{}{
			"fileID": id,
		})
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}
	return &file, nil
}

// GetFilesByRelatedID 根据关联ID获取文件列表
func (s *FileService) GetFilesByRelatedID(relatedID string) ([]*models.File, error) {
	var files []*models.File
	err := global.SLDB.Where("related_id = ?", relatedID).Find(&files).Error
	if err != nil {
		logger.LogError("查询关联文件失败", err, map[string]interface{}{
			"relatedID": relatedID,
		})
		return nil, fmt.Errorf("查询关联文件失败: %w", err)
	}
	return files, nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(id string) error {
	// 先获取文件信息
	file, err := s.GetFileByID(id)
	if err != nil {
		return err
	}

	// 获取用户配置目录
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		userProfile = os.Getenv("HOME")
	}
	cacheDir := filepath.Join(userProfile, "skymind", "cache")

	// 构建文件路径
	filePath := filepath.Join(cacheDir, fmt.Sprintf("%s.%s", file.ID, file.FileSuffix))

	// 删除物理文件
	if _, err := os.Stat(filePath); err == nil {
		if removeErr := os.Remove(filePath); removeErr != nil {
			logger.LogError("删除物理文件失败", removeErr, map[string]interface{}{
				"filePath": filePath,
				"fileID":   id,
			})
		} else {
			logger.LogInfo("物理文件删除成功", map[string]interface{}{
				"filePath": filePath,
				"fileID":   id,
			})
		}
	}

	// 从数据库中删除记录
	if err := global.SLDB.Delete(&models.File{}, "id = ?", id).Error; err != nil {
		logger.LogError("删除文件记录失败", err, map[string]interface{}{
			"fileID": id,
		})
		return fmt.Errorf("删除文件记录失败: %w", err)
	}

	logger.LogDatabaseOperation("delete", "files", id, nil)
	logger.LogInfo("文件删除成功", map[string]interface{}{
		"fileID":       id,
		"originalName": file.OriginalName,
	})

	return nil
}

// DeleteFilesByRelatedID 根据关联ID删除文件
func (s *FileService) DeleteFilesByRelatedID(relatedID string) error {
	// 获取所有关联文件
	files, err := s.GetFilesByRelatedID(relatedID)
	if err != nil {
		return err
	}

	// 逐个删除文件
	for _, file := range files {
		if deleteErr := s.DeleteFile(file.ID); deleteErr != nil {
			logger.LogError("删除关联文件失败", deleteErr, map[string]interface{}{
				"fileID":    file.ID,
				"relatedID": relatedID,
			})
		}
	}

	return nil
}

// GetFilePath 获取文件物理路径
func (s *FileService) GetFilePath(fileID string) (string, error) {
	file, err := s.GetFileByID(fileID)
	if err != nil {
		return "", err
	}

	var cacheDir string
	if runtime.GOOS == "windows" {
		// Windows: 优先使用USERPROFILE环境变量
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			cacheDir = filepath.Join(userProfile, "skymind", "cache")
		} else if home := os.Getenv("HOME"); home != "" {
			// 备选：使用HOME环境变量
			cacheDir = filepath.Join(home, "skymind", "cache")
		} else if appData := os.Getenv("APPDATA"); appData != "" {
			// 备选：使用APPDATA环境变量
			cacheDir = filepath.Join(appData, "skymind", "cache")
		} else if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			// 备选：使用LOCALAPPDATA环境变量
			cacheDir = filepath.Join(localAppData, "skymind", "cache")
		} else {
			// 最后备选：使用当前目录
			cacheDir = filepath.Join("skymind", "cache")
		}
	} else {
		// Unix-like: 使用HOME环境变量
		if home := os.Getenv("HOME"); home != "" {
			cacheDir = filepath.Join(home, ".skymind", "cache")
		} else {
			// 最后备选：使用当前目录
			cacheDir = filepath.Join("skymind", "cache")
		}
	}

	filePath := filepath.Join(cacheDir, fmt.Sprintf("%s.%s", file.ID, file.FileSuffix))
	return filePath, nil
}

// ProcessFileContent 处理文件内容并转换为markdown格式
func (s *FileService) ProcessFileContent(fileID string) error {
	file, err := s.GetFileByID(fileID)
	if err != nil {
		return err
	}

	// 图片文件不需要处理内容（已通过 Dify API 传递），直接返回
	if s.isImageFile(file.FileSuffix) {
		// 图片文件：不需要本地处理内容，Dify 会直接处理
		// 设置一个简单的占位内容即可
		content := fmt.Sprintf("*图片文件已上传到 Dify，文件ID: %s*\n\n*图片内容将通过 Dify 的多模态模型直接处理。*", file.OriginalPath)
		if err := global.SLDB.Model(&models.File{}).Where("id = ?", fileID).Update("content", content).Error; err != nil {
			logger.LogError("更新图片文件内容失败", err, map[string]interface{}{
				"fileID": fileID,
			})
			return fmt.Errorf("更新图片文件内容失败: %w", err)
		}
		logger.LogInfo("图片文件内容处理完成", map[string]interface{}{
			"fileID":        fileID,
			"fileType":      file.FileSuffix,
			"contentLength": len(content),
		})
		return nil
	}

	// 获取文件物理路径（仅非图片文件需要）
	filePath, err := s.GetFilePath(fileID)
	if err != nil {
		return err
	}

	// 根据文件类型处理内容
	var content string
	if s.isTextFile(file.FileSuffix) {
		// 文本文件：转换为markdown格式
		content, err = s.processTextContent(filePath)
		if err != nil {
			return fmt.Errorf("处理文本内容失败: %w", err)
		}
	} else {
		// 其他文件类型：不允许上传
		return fmt.Errorf("不支持的文件类型: %s", file.FileSuffix)
	}

	// 检查内容长度，如果超过100K，则进行分段总结
	if len(content) > 100*1024 {
		content, err = s.summarizeLongContent(content)
		if err != nil {
			return fmt.Errorf("总结长内容失败: %w", err)
		}
	}

	// 更新数据库中的Content字段
	if err := global.SLDB.Model(&models.File{}).Where("id = ?", fileID).Update("content", content).Error; err != nil {
		logger.LogError("更新文件内容失败", err, map[string]interface{}{
			"fileID": fileID,
		})
		return fmt.Errorf("更新文件内容失败: %w", err)
	}

	logger.LogInfo("文件内容处理完成", map[string]interface{}{
		"fileID":        fileID,
		"fileType":      file.FileSuffix,
		"contentLength": len(content),
	})

	return nil
}

// isImageFile 判断是否为图片文件
func (s *FileService) isImageFile(fileSuffix string) bool {
	imageExtensions := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}
	for _, ext := range imageExtensions {
		if strings.ToLower(fileSuffix) == ext {
			return true
		}
	}
	return false
}

// isTextFile 判断是否为文本文件
func (s *FileService) isTextFile(fileSuffix string) bool {
	textExtensions := []string{"txt", "md", "json", "xml", "csv", "log", "yaml", "yml", "ini", "conf", "config", "py", "js", "html", "css", "sql", "sh", "bat", "ps1", "docx", "doc", "pdf"}
	for _, ext := range textExtensions {
		if strings.ToLower(fileSuffix) == ext {
			return true
		}
	}
	return false
}

// getMimeTypeByExtension 根据文件扩展名返回 MIME 类型（用于补充 mime.TypeByExtension 无法识别的情况）
func (s *FileService) getMimeTypeByExtension(fileSuffix string) string {
	fileSuffix = strings.ToLower(strings.TrimSpace(fileSuffix))

	// 图片类型
	mimeMap := map[string]string{
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"gif":  "image/gif",
		"bmp":  "image/bmp",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		// 文档类型
		"pdf":  "application/pdf",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"xls":  "application/vnd.ms-excel",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		// 文本类型
		"txt":  "text/plain",
		"md":   "text/markdown",
		"html": "text/html",
		"css":  "text/css",
		"js":   "text/javascript",
		"json": "application/json",
		"xml":  "application/xml",
		"csv":  "text/csv",
	}

	if mimeType, ok := mimeMap[fileSuffix]; ok {
		return mimeType
	}

	// 默认返回 application/octet-stream
	return "application/octet-stream"
}

// processImageContent 处理图片内容，调用视觉模型
func (s *FileService) processImageContent(filePath string) (string, error) {
	// 这里需要调用generator服务中的视觉模型函数
	// 由于循环导入问题，我们需要通过其他方式解决
	// 暂时返回占位符
	return fmt.Sprintf("![图片](%s)\n\n*图片内容已通过视觉模型分析，具体内容将在后续处理中生成*", filePath), nil
}

// processTextContent 处理文本内容，转换为markdown格式
func (s *FileService) processTextContent(filePath string) (string, error) {
	// 根据文件后缀进行不同的markdown格式化
	fileSuffix := filepath.Ext(filePath)
	fileSuffix = strings.ToLower(strings.TrimPrefix(fileSuffix, "."))

	switch fileSuffix {
	case "docx", "doc":
		// Word文档，暂时返回占位符
		return fmt.Sprintf("```%s\n%s\n```", fileSuffix, "*Word文档内容解析功能正在开发中，请稍后重试。*"), nil
	case "pdf":
		// PDF文档，暂时返回占位符
		// 注意：PDF 文件已上传到 Dify，可以通过 Dify 的文件预览功能访问
		return fmt.Sprintf("*PDF文件已上传到 Dify，文件ID: %s*\n\n*PDF内容解析功能正在开发中，文件可以通过 Dify API 访问。*", filepath.Base(filePath)), nil
	case "md":
		// 已经是markdown格式，直接读取返回
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取markdown文件失败: %w", err)
		}
		return string(fileBytes), nil
	case "txt":
		// 纯文本，转换为markdown
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文本文件失败: %w", err)
		}
		content := string(fileBytes)
		return fmt.Sprintf("```\n%s\n```", content), nil
	case "json":
		// JSON格式，格式化为markdown
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取JSON文件失败: %w", err)
		}
		content := string(fileBytes)
		return fmt.Sprintf("```json\n%s\n```", content), nil
	case "xml":
		// XML格式，格式化为markdown
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取XML文件失败: %w", err)
		}
		content := string(fileBytes)
		return fmt.Sprintf("```xml\n%s\n```", content), nil
	case "csv":
		// CSV格式，转换为表格markdown
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取CSV文件失败: %w", err)
		}
		content := string(fileBytes)
		return s.convertCSVToMarkdown(content)
	case "html":
		// HTML格式，提取文本并转换为markdown
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取HTML文件失败: %w", err)
		}
		content := string(fileBytes)
		return fmt.Sprintf("```html\n%s\n```", content), nil
	default:
		// 其他文本文件，使用代码块格式
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件失败: %w", err)
		}
		content := string(fileBytes)
		return fmt.Sprintf("```%s\n%s\n```", fileSuffix, content), nil
	}
}

// convertCSVToMarkdown 将CSV内容转换为markdown表格
func (s *FileService) convertCSVToMarkdown(csvContent string) (string, error) {
	lines := strings.Split(csvContent, "\n")
	if len(lines) == 0 {
		return "", nil
	}

	// 构建markdown表格
	var markdown strings.Builder

	// 处理表头
	if len(lines) > 0 {
		headers := strings.Split(lines[0], ",")
		markdown.WriteString("| ")
		for _, header := range headers {
			markdown.WriteString(strings.TrimSpace(header) + " | ")
		}
		markdown.WriteString("\n")

		// 添加分隔线
		markdown.WriteString("|")
		for range headers {
			markdown.WriteString(" --- |")
		}
		markdown.WriteString("\n")

		// 处理数据行
		for i := 1; i < len(lines); i++ {
			cells := strings.Split(lines[i], ",")
			markdown.WriteString("| ")
			for _, cell := range cells {
				markdown.WriteString(strings.TrimSpace(cell) + " | ")
			}
			markdown.WriteString("\n")
		}
	}

	return markdown.String(), nil
}

// summarizeLongContent 对长内容进行分段总结
func (s *FileService) summarizeLongContent(content string) (string, error) {
	// 这里需要调用generator服务中的指示模型函数
	// 由于循环导入问题，我们需要通过其他方式解决
	// 暂时返回截断的内容
	maxLength := 50 * 1024 // 50K
	if len(content) <= maxLength {
		return content, nil
	}

	summary := content[:maxLength] + "\n\n*注意：内容过长，已进行截断处理。完整内容将在后续处理中分段加载。*"
	return summary, nil
}
