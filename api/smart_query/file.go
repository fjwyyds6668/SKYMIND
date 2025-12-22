package smart_query

import (
	"skymind/models"
)

// FileAPI 文件API
type FileAPI struct {
}

// SaveFile 保存文件信息（不包含文件内容，只保存元数据）
func (api *FileAPI) SaveFile(fileName, originalName, fileSuffix, md5, localPath string, fileSize int64, relatedID string) (*models.File, error) {
	// 创建文件记录（不保存文件内容，只保存元数据）
	fileRecord := &models.File{
		ID:           "", // 将在Service层生成
		OriginalPath: localPath,
		OriginalMD5:  md5,
		OriginalName: originalName,
		FileSuffix:   fileSuffix,
		FileSize:     fileSize,
		RelatedID:    relatedID,
	}
	
	return fileService.SaveFileMetadata(fileRecord)
}

// GetFileByID 根据ID获取文件信息
func (api *FileAPI) GetFileByID(id string) (*models.File, error) {
	return fileService.GetFileByID(id)
}

// GetFilesByRelatedID 根据关联ID获取文件列表
func (api *FileAPI) GetFilesByRelatedID(relatedID string) ([]*models.File, error) {
	return fileService.GetFilesByRelatedID(relatedID)
}

// DeleteFile 删除文件
func (api *FileAPI) DeleteFile(id string) error {
	return fileService.DeleteFile(id)
}

// DeleteFilesByRelatedID 根据关联ID删除文件
func (api *FileAPI) DeleteFilesByRelatedID(relatedID string) error {
	return fileService.DeleteFilesByRelatedID(relatedID)
}

// GetFilePath 获取文件物理路径
func (api *FileAPI) GetFilePath(fileID string) (string, error) {
	return fileService.GetFilePath(fileID)
}

// ProcessFileContent 处理文件内容并转换为markdown格式
func (api *FileAPI) ProcessFileContent(fileID string) error {
	return fileService.ProcessFileContent(fileID)
}
