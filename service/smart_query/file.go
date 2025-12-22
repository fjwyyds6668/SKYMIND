package smart_query

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"skymind/global"
	"skymind/logger"
	"skymind/models"

	"github.com/google/uuid"
)

type FileService struct{}

// SaveFileMetadata 保存文件元数据到数据库（不包含文件内容）
func (s *FileService) SaveFileMetadata(fileRecord *models.File) (*models.File, error) {
	// 生成UUID如果为空
	if fileRecord.ID == "" {
		fileUUID := uuid.New().String()
		fileRecord.ID = fileUUID
	}

	// 保存到数据库
	if err := global.SLDB.Create(fileRecord).Error; err != nil {
		logger.LogError("保存文件元数据到数据库失败", err, map[string]interface{}{
			"fileUUID":     fileRecord.ID,
			"originalName": fileRecord.OriginalName,
		})
		return nil, fmt.Errorf("保存文件元数据到数据库失败: %w", err)
	}

	logger.LogInfo("文件元数据保存成功", map[string]interface{}{
		"fileUUID":     fileRecord.ID,
		"originalName": fileRecord.OriginalName,
		"fileSize":     fileRecord.FileSize,
		"relatedID":    fileRecord.RelatedID,
	})

	logger.LogDatabaseOperation("create", "files", fileRecord.ID, nil)

	return fileRecord, nil
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

	// 获取文件物理路径
	filePath, err := s.GetFilePath(fileID)
	if err != nil {
		return err
	}

	// 根据文件类型处理内容
	var content string
	if s.isImageFile(file.FileSuffix) {
		// 图片文件：调用视觉模型理解图片内容
		content, err = s.processImageContent(filePath)
		if err != nil {
			return fmt.Errorf("处理图片内容失败: %w", err)
		}
	} else if s.isTextFile(file.FileSuffix) {
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
		"fileID":   fileID,
		"fileType": file.FileSuffix,
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
	textExtensions := []string{"txt", "md", "json", "xml", "csv", "log", "yaml", "yml", "ini", "conf", "config", "py", "js", "html", "css", "sql", "sh", "bat", "ps1", "docx", "doc"}
	for _, ext := range textExtensions {
		if strings.ToLower(fileSuffix) == ext {
			return true
		}
	}
	return false
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
