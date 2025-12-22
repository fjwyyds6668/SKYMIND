package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"log"

	"skymind/models"

	"gopkg.in/yaml.v3"
)

// ModelConfigYAML YAML配置文件结构
type ModelConfigYAML struct {
	InstructModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"instruct_model"`
	ThinkingModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"thinking_model"`
	FastModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"fast_model"`
	VisualModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"visual_model"`
	EmbeddingModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"embedding_model"`
	RerankerModel struct {
		ID      string `yaml:"id"`
		ApiBase string `yaml:"api_base"`
		ApiKey  string `yaml:"api_key"`
		Name    string `yaml:"name"`
	} `yaml:"reranker_model"`
}

// getUserConfigDir 获取用户配置目录
func getUserConfigDir() string {
	if runtime.GOOS == "windows" {
		// Windows: 优先使用USERPROFILE环境变量
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			configDir := filepath.Join(userProfile, "skymind")
			log.Printf("Using USERPROFILE: %s", configDir)
			return configDir
		}
		// 备选：使用HOME环境变量
		if home := os.Getenv("HOME"); home != "" {
			configDir := filepath.Join(home, "skymind")
			log.Printf("Using HOME: %s", configDir)
			return configDir
		}
		// 备选：使用APPDATA环境变量
		if appData := os.Getenv("APPDATA"); appData != "" {
			configDir := filepath.Join(appData, "skymind")
			log.Printf("Using APPDATA: %s", configDir)
			return configDir
		}
		// 备选：使用LOCALAPPDATA环境变量
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			configDir := filepath.Join(localAppData, "skymind")
			log.Printf("Using LOCALAPPDATA: %s", configDir)
			return configDir
		}
	} else {
		// Unix-like: 使用HOME环境变量
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".skymind")
		}
	}

	// 最后备选：使用当前目录
	log.Printf("Using current directory as fallback")
	return "skymind"
}

// EnsureUserConfigFile 确保用户配置文件存在
func EnsureUserConfigFile() error {
	userConfigDir := getUserConfigDir()
	userConfigPath := filepath.Join(userConfigDir, "model_config.yml")

	log.Printf("Checking user config file at: %s", userConfigPath)

	// 检查用户配置文件是否存在
	if _, err := os.Stat(userConfigPath); err == nil {
		log.Printf("User config file already exists")
		return nil // 文件已存在
	}

	log.Printf("User config file does not exist, creating...")

	// 创建用户配置目录
	if err := os.MkdirAll(userConfigDir, 0755); err != nil {
		log.Printf("Failed to create user config directory: %v", err)
		return fmt.Errorf("failed to create user config directory: %w", err)
	}

	log.Printf("Created user config directory: %s", userConfigDir)

	// 尝试从多个可能的路径复制配置文件
	possiblePaths := []string{
		filepath.Join("appcore", "model_config.yml"),
		filepath.Join(".", "appcore", "model_config.yml"),
		filepath.Join("..", "appcore", "model_config.yml"),
	}

	var sourceFile *os.File
	var err error
	var usedPath string

	for _, path := range possiblePaths {
		log.Printf("Trying to find config file at: %s", path)
		if _, err := os.Stat(path); err == nil {
			sourceFile, err = os.Open(path)
			if err == nil {
				usedPath = path
				log.Printf("Found config file at: %s", usedPath)
				break
			}
		}
	}

	// 如果找不到外部配置文件，创建默认配置
	if sourceFile == nil {
		log.Printf("No external config file found, creating default config")
		return createDefaultConfigFile(userConfigPath)
	}

	defer sourceFile.Close()

	destFile, err := os.Create(userConfigPath)
	if err != nil {
		log.Printf("Failed to create destination config file: %v", err)
		return fmt.Errorf("failed to create destination config file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		log.Printf("Failed to copy config file: %v", err)
		return fmt.Errorf("failed to copy config file: %w", err)
	}

	log.Printf("Successfully copied config file from %s to %s", usedPath, userConfigPath)
	return nil
}

// createDefaultConfigFile 创建默认配置文件
func createDefaultConfigFile(userConfigPath string) error {
	defaultConfig := `# 模型配置文件
instruct_model:
  id: "Qwen/Qwen3-235B-A22B-Instruct-2507"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "Qwen/Qwen3-235B-A22B-Instruct-2507"

thinking_model:
  id: "Qwen/Qwen3-235B-A22B-Thinking-2507"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "通义千问3-235B-思考"

fast_model:
  id: "Qwen/Qwen3-32B"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "通义千问3-32B"

visual_model:
  id: "Qwen/Qwen3-VL-235B-A22B-Instruct"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "通义千问3-235B-视觉"

embedding_model:
  id: "Qwen/Qwen3-Embedding-8B"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "通义千问3-8B-嵌入"

reranker_model:
  id: "Qwen/Qwen3-Reranker-8B"
  api_base: "https://api-inference.modelscope.cn/v1"
  api_key: "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388"
  name: "通义千问3-8B-排序"
`

	err := os.WriteFile(userConfigPath, []byte(defaultConfig), 0644)
	if err != nil {
		log.Printf("Failed to create default config file: %v", err)
		return fmt.Errorf("failed to create default config file: %w", err)
	}

	log.Printf("Successfully created default config file: %s", userConfigPath)
	return nil
}

// getUserModelConfigPath 获取用户模型配置文件路径
func getUserModelConfigPath() string {
	return filepath.Join(getUserConfigDir(), "model_config.yml")
}

// readUserModelConfig 从用户配置文件读取模型配置
func readUserModelConfig() (*ModelConfigYAML, error) {
	// 确保用户配置文件存在
	if err := EnsureUserConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to ensure user config file: %w", err)
	}

	userConfigPath := getUserModelConfigPath()
	yamlFile, err := os.ReadFile(userConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user config file: %w", err)
	}

	var yamlConfig ModelConfigYAML
	if err := yaml.Unmarshal(yamlFile, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse user config file: %w", err)
	}

	return &yamlConfig, nil
}

// GetInstructModelConfig 获取指示模型配置
func GetInstructModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		// 读取失败，返回硬编码的默认配置
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-235B-A22B-Instruct-2507",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "Qwen/Qwen3-235B-A22B-Instruct-2507",
		}
	}

	// 成功读取用户配置，返回instruct_model配置
	return models.ModelConfig{
		ID:      yamlConfig.InstructModel.ID,
		ApiBase: yamlConfig.InstructModel.ApiBase,
		ApiKey:  yamlConfig.InstructModel.ApiKey,
		Name:    yamlConfig.InstructModel.Name,
	}
}

// GetThinkingModelConfig 获取思考模型配置
func GetThinkingModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-235B-A22B-Thinking-2507",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "通义千问3-235B-思考",
		}
	}

	return models.ModelConfig{
		ID:      yamlConfig.ThinkingModel.ID,
		ApiBase: yamlConfig.ThinkingModel.ApiBase,
		ApiKey:  yamlConfig.ThinkingModel.ApiKey,
		Name:    yamlConfig.ThinkingModel.Name,
	}
}

// GetFastModelConfig 获取快速模型配置
func GetFastModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-32B",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "通义千问3-32B",
		}
	}

	return models.ModelConfig{
		ID:      yamlConfig.FastModel.ID,
		ApiBase: yamlConfig.FastModel.ApiBase,
		ApiKey:  yamlConfig.FastModel.ApiKey,
		Name:    yamlConfig.FastModel.Name,
	}
}

// GetVisualModelConfig 获取视觉模型配置
func GetVisualModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-VL-235B-A22B-Instruct",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "通义千问3-235B-视觉",
		}
	}

	return models.ModelConfig{
		ID:      yamlConfig.VisualModel.ID,
		ApiBase: yamlConfig.VisualModel.ApiBase,
		ApiKey:  yamlConfig.VisualModel.ApiKey,
		Name:    yamlConfig.VisualModel.Name,
	}
}

// GetEmbeddingModelConfig 获取嵌入模型配置
func GetEmbeddingModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-Embedding-8B",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "通义千问3-8B-嵌入",
		}
	}

	return models.ModelConfig{
		ID:      yamlConfig.EmbeddingModel.ID,
		ApiBase: yamlConfig.EmbeddingModel.ApiBase,
		ApiKey:  yamlConfig.EmbeddingModel.ApiKey,
		Name:    yamlConfig.EmbeddingModel.Name,
	}
}

// GetRerankerModelConfig 获取排序模型配置
func GetRerankerModelConfig() models.ModelConfig {
	yamlConfig, err := readUserModelConfig()
	if err != nil {
		return models.ModelConfig{
			ID:      "Qwen/Qwen3-Reranker-8B",
			ApiBase: "https://api-inference.modelscope.cn/v1",
			ApiKey:  "ms-b7f877a2-4d7d-4846-8576-e7aec9b40388",
			Name:    "通义千问3-8B-排序",
		}
	}

	return models.ModelConfig{
		ID:      yamlConfig.RerankerModel.ID,
		ApiBase: yamlConfig.RerankerModel.ApiBase,
		ApiKey:  yamlConfig.RerankerModel.ApiKey,
		Name:    yamlConfig.RerankerModel.Name,
	}
}
