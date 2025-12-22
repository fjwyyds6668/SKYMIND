package database

import (
	"fmt"
	"sync"
	"time"
)

// Snowflake 短雪花算法生成器
// 生成不超过16位的ID，起始日期为2025年11月1日
// ID结构：时间戳(41位) + 机器ID(5位) + 序列号(8位) = 54位
// 最大值：2^54-1 = 18014398509481983 (17位，但实际使用中会小于16位)
type Snowflake struct {
	mu          sync.Mutex
	timestamp   int64    // 上次生成ID的时间戳
	machineID   int64    // 机器ID (0-31)
	sequence    int64    // 序列号 (0-255)
	epoch       int64    // 起始时间戳 (2025年11月1日)
}

const (
	// 位数分配
	timestampBits = 41 // 时间戳位数
	machineIDBits = 5  // 机器ID位数
	sequenceBits  = 8  // 序列号位数

	// 最大值
	maxMachineID = -1 ^ (-1 << machineIDBits) // 31
	maxSequence  = -1 ^ (-1 << sequenceBits)  // 255

	// 位移
	machineIDShift = sequenceBits                    // 8
	timestampShift = sequenceBits + machineIDBits     // 13
)

// NewSnowflake 创建新的雪花算法生成器
func NewSnowflake(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, fmt.Errorf("machine ID must be between 0 and %d", maxMachineID)
	}

	// 2025年11月1日 00:00:00 UTC
	epoch := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

	return &Snowflake{
		timestamp: 0,
		machineID: machineID,
		sequence:  0,
		epoch:     epoch,
	}, nil
}

// Generate 生成新的ID
func (s *Snowflake) Generate() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// 如果当前时间小于上次生成ID的时间，时钟回拨
	if now < s.timestamp {
		return 0, fmt.Errorf("clock moved backwards, refusing to generate id")
	}

	// 如果是同一毫秒内
	if now == s.timestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 序列号溢出，等待下一毫秒
		if s.sequence == 0 {
			now = s.waitNextMillis(now)
		}
	} else {
		// 不同毫秒，重置序列号
		s.sequence = 0
	}

	s.timestamp = now

	// 组装ID
	id := ((now - s.epoch) << timestampShift) |
		(s.machineID << machineIDShift) |
		s.sequence

	return id, nil
}

// GenerateString 生成字符串格式的ID
func (s *Snowflake) GenerateString() (string, error) {
	id, err := s.Generate()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// waitNextMillis 等待下一毫秒
func (s *Snowflake) waitNextMillis(currentTime int64) int64 {
	for {
		now := time.Now().UnixMilli()
		if now > currentTime {
			return now
		}
	}
}

// ParseID 解析ID获取时间信息
func (s *Snowflake) ParseID(id int64) (time.Time, int64, int64, error) {
	timestamp := (id >> timestampShift) + s.epoch
	machineID := (id >> machineIDShift) & maxMachineID
	sequence := id & maxSequence

	return time.UnixMilli(timestamp), machineID, sequence, nil
}

// 全局雪花生成器实例
var globalSnowflake *Snowflake
var snowflakeOnce sync.Once

// InitGlobalSnowflake 初始化全局雪花生成器
func InitGlobalSnowflake(machineID int64) error {
	var err error
	snowflakeOnce.Do(func() {
		globalSnowflake, err = NewSnowflake(machineID)
	})
	return err
}

// GenerateID 生成全局唯一ID
func GenerateID() (int64, error) {
	if globalSnowflake == nil {
		return 0, fmt.Errorf("global snowflake not initialized, call InitGlobalSnowflake first")
	}
	return globalSnowflake.Generate()
}

// GenerateIDString 生成字符串格式的全局唯一ID
func GenerateIDString() (string, error) {
	if globalSnowflake == nil {
		return "", fmt.Errorf("global snowflake not initialized, call InitGlobalSnowflake first")
	}
	return globalSnowflake.GenerateString()
}
