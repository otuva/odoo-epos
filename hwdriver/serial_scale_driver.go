package hwdriver

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type SerialScaleDriver struct {
	BaseDriver
	SerialProtocol
	MeasureRegexp     string
	StatusRegexp      string
	CommandTerminator string
	CommandDelay      int // in milliseconds
	MeasureDelay      int // in milliseconds
	NewMeasureDelay   int // in milliseconds
	MeasureCommand    string
	EmptyAnswerValid  bool
	weight            float64 // in kg
	status            HWStatus
	RetryCount        int
	RetryInterval     int // in milliseconds
	Debug             bool
	mu                sync.RWMutex // 保护weight和status的读写
	connMu            sync.Mutex   // 保护串口连接操作
	serialPort        *serial.Port
}

func (s *SerialScaleDriver) GetStatus() HWStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// get weight in kg
func (s *SerialScaleDriver) GetWeight() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weight
}

// get weight in g
func (s *SerialScaleDriver) GetWeightInGrams() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weight * 1000
}

// get weight in lb
func (s *SerialScaleDriver) GetWeightInPounds() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weight * 2.20462
}

// setStatus 内部方法，用于更新电子秤状态（调用前需要持有锁）
func (s *SerialScaleDriver) setStatus(status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = HWStatus{
		Status:  status,
		Message: message,
	}
	if s.Debug {
		log.Printf("[%s] Status updated: %s - %s", s.DeviceIdentifier, status, message)
	}
}

// setWeight 内部方法，用于更新电子秤重量（调用前需要持有锁）
func (s *SerialScaleDriver) setWeight(weight float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.weight = weight
	if s.Debug {
		log.Printf("[%s] Weight updated: %.3f kg", s.DeviceIdentifier, weight)
	}
}

// ensureConnection 确保串口连接可用，如果断开则重新连接
func (s *SerialScaleDriver) ensureConnection() error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	// 如果连接已存在，先测试是否有效
	if s.serialPort != nil {
		// 简单测试：尝试写入空字节
		_, err := s.serialPort.Write([]byte{})
		if err == nil {
			return nil // 连接有效
		}
		// 连接无效，关闭
		s.serialPort.Close()
		s.serialPort = nil
	}

	// 建立新连接
	config := s.SerialProtocol.getSerialConfig()
	port, err := serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}
	s.serialPort = port

	if s.Debug {
		log.Printf("[%s] Serial port connected: %s", s.DeviceIdentifier, s.Port)
	}

	return nil
}

// Disconnect 断开串口连接
func (s *SerialScaleDriver) Disconnect() error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.serialPort != nil {
		err := s.serialPort.Close()
		s.serialPort = nil
		if err != nil {
			return err
		}
		if s.Debug {
			log.Printf("[%s] Serial port disconnected", s.DeviceIdentifier)
		}
	}
	return nil
}

// clearSerialBuffer 清空串口缓冲区，只保留最新的数据
func (s *SerialScaleDriver) clearSerialBuffer() error {
	// 设置短暂的读超时来清空缓冲区
	buf := make([]byte, 1024)
	for {
		// 使用非阻塞方式读取
		n, err := s.serialPort.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// 其他错误也认为缓冲区已清空
			break
		}
		if n == 0 {
			break
		}
		if s.Debug {
			log.Printf("[%s] Cleared %d bytes from buffer", s.DeviceIdentifier, n)
		}
	}
	return nil
}

// convertWeightToKg 将重量值转换为千克
func convertWeightToKg(value float64, unit string) float64 {
	switch unit {
	case "g", "G":
		return value / 1000.0
	case "lb", "LB", "lbs":
		return value / 2.20462
	case "kg", "KG", "Kg":
		return value
	default:
		// 默认假定为kg
		return value
	}
}

// ReadWeight 读取电子秤重量，返回错误（如果有）
func (s *SerialScaleDriver) ReadWeight() error {
	var lastErr error

	// 尝试连接和读取，包含重试机制
	for attempt := 0; attempt <= s.RetryCount; attempt++ {
		if attempt > 0 {
			// 重试时设置状态为connecting
			s.setStatus(StatusConnecting, fmt.Sprintf("Retry %d/%d", attempt, s.RetryCount))
			time.Sleep(time.Duration(s.RetryInterval) * time.Millisecond)
		}

		// 确保串口连接
		if err := s.ensureConnection(); err != nil {
			lastErr = err
			if s.Debug {
				log.Printf("[%s] Connection attempt %d failed: %v", s.DeviceIdentifier, attempt+1, err)
			}
			continue
		}

		// 清空串口缓冲区，避免读取历史数据
		s.clearSerialBuffer()

		// 添加命令延迟
		if s.CommandDelay > 0 {
			time.Sleep(time.Duration(s.CommandDelay) * time.Millisecond)
		}

		// 发送测量命令
		if s.MeasureCommand != "" {
			command := s.MeasureCommand + s.CommandTerminator
			_, err := s.serialPort.Write([]byte(command))
			if err != nil {
				lastErr = fmt.Errorf("failed to write command: %w", err)
				if s.Debug {
					log.Printf("[%s] Write error: %v", s.DeviceIdentifier, err)
				}
				// 写入失败，关闭连接以便下次重连
				s.Disconnect()
				continue
			}
		}

		// 等待电子秤响应
		if s.MeasureDelay > 0 {
			time.Sleep(time.Duration(s.MeasureDelay) * time.Millisecond)
		}

		// 读取响应
		reader := bufio.NewReader(s.serialPort)
		line, err := reader.ReadString('\n')
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			if s.Debug {
				log.Printf("[%s] Read error: %v", s.DeviceIdentifier, err)
			}
			// 读取失败，关闭连接以便下次重连
			s.Disconnect()
			continue
		}

		if s.Debug {
			log.Printf("[%s] Received: %q", s.DeviceIdentifier, line)
		}

		// 解析重量数据
		re := regexp.MustCompile(s.MeasureRegexp)
		matches := re.FindStringSubmatch(line)

		if len(matches) < 2 {
			// 如果空响应有效，则设置重量为0
			if s.EmptyAnswerValid {
				s.setWeight(0)
				s.setStatus(StatusConnected, "Connected (no weight)")
				return nil
			}
			lastErr = fmt.Errorf("failed to parse weight from response: %q", line)
			if s.Debug {
				log.Printf("[%s] Parse error: %v", s.DeviceIdentifier, lastErr)
			}
			continue
		}

		// 提取重量值
		weightStr := matches[1]
		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			lastErr = fmt.Errorf("failed to convert weight: %w", err)
			if s.Debug {
				log.Printf("[%s] Convert error: %v", s.DeviceIdentifier, err)
			}
			continue
		}

		// 提取单位（如果有）
		unit := "kg" // 默认单位
		if len(matches) >= 3 && matches[2] != "" {
			unit = matches[2]
		}

		// 转换为kg
		weightKg := convertWeightToKg(weight, unit)

		// 更新重量和状态
		s.setWeight(weightKg)
		s.setStatus(StatusConnected, "Connected")

		if s.Debug {
			log.Printf("[%s] Successfully read weight: %.3f %s = %.3f kg",
				s.DeviceIdentifier, weight, unit, weightKg)
		}

		return nil
	}

	// 所有重试都失败，设置为断开状态
	s.setStatus(StatusDisconnected, fmt.Sprintf("Failed after %d retries: %v", s.RetryCount, lastErr))
	return lastErr
}
