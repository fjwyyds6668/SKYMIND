package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"skymind/global"
	"skymind/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDatabase GORM数据库结构体
type GormDatabase struct {
	db *gorm.DB
}

// NewGormDatabase 创建新的GORM数据库实例
func NewGormDatabase(dataPath string) (*GormDatabase, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(dataPath, "skymind.db")

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 可以根据需要调整日志级别
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 获取底层的sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 根据操作系统设置连接池参数（Windows 7兼容性）
	if runtime.GOOS == "windows" {
		// Windows 7兼容性配置：减少连接数，增加连接生命周期
		sqlDB.SetMaxOpenConns(5)                 // 减少最大连接数
		sqlDB.SetMaxIdleConns(2)                 // 减少空闲连接数
		sqlDB.SetConnMaxLifetime(24 * time.Hour) // 增加连接生命周期
	} else {
		// 其他操作系统使用原配置
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置全局数据库连接
	global.SetDB(db)

	gormDB := &GormDatabase{db: db}

	// 自动迁移数据库表
	if err := gormDB.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化全局雪花算法生成器（机器ID为1）
	if err := InitGlobalSnowflake(1); err != nil {
		return nil, fmt.Errorf("failed to initialize snowflake: %w", err)
	}

	return gormDB, nil
}

// Close 关闭数据库连接
func (d *GormDatabase) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// autoMigrate 自动迁移数据库表
func (d *GormDatabase) autoMigrate() error {
	// 定义需要迁移的模型
	models := []interface{}{
		&models.Assistant{},
		&models.Topic{},
		&models.Conversation{},
		&models.Message{},
		&models.Memory{},
		&models.MemoryHistory{},
		&models.File{},
	}

	// 执行自动迁移
	for _, model := range models {
		if err := d.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	// 创建额外的索引
	if err := d.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes 创建额外的索引
func (d *GormDatabase) createIndexes() error {
	// 为助手的name和active字段创建复合索引
	if err := d.db.Exec("CREATE INDEX IF NOT EXISTS idx_assistants_name_active ON assistants(name, is_active)").Error; err != nil {
		return fmt.Errorf("failed to create assistants name_active index: %w", err)
	}

	// 为话题的assistant_id和sort_order创建复合索引
	if err := d.db.Exec("CREATE INDEX IF NOT EXISTS idx_topics_assistant_sort ON topics(assistant_id, sort_order)").Error; err != nil {
		return fmt.Errorf("failed to create topics assistant_sort index: %w", err)
	}

	// 为对话的topic_id和created_at创建复合索引
	if err := d.db.Exec("CREATE INDEX IF NOT EXISTS idx_conversations_topic_created ON conversations(topic_id, created_at)").Error; err != nil {
		return fmt.Errorf("failed to create conversations topic_created index: %w", err)
	}

	// 为消息的conversation_id和created_at创建复合索引
	if err := d.db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at)").Error; err != nil {
		return fmt.Errorf("failed to create messages conversation_created index: %w", err)
	}

	return nil
}

// GetDB 获取GORM数据库连接
func (d *GormDatabase) GetDB() *gorm.DB {
	return d.db
}

// GetDefaultAssistantSettings 获取默认助手设置
func GetDefaultAssistantSettings() models.AssistantSettings {
	return models.AssistantSettings{
		Temperature:     1.0,
		ContextCount:    5,
		EnableMaxTokens: false,
		MaxTokens:       0,
		StreamOutput:    true,
		TopP:            1.0,
		ToolUseMode:     "prompt",
		ReasoningEffort: "low",
		QwenThinkMode:   false,
	}
}
