package hwdriver

import "github.com/tarm/serial"

type SerialProtocol struct {
	Port         string // eg: "COM3" or "/dev/ttyS0"
	BaudRate     int
	ByteSize     int
	Parity       byte
	StopBits     byte
	Timeout      int // in seconds
	WriteTimeout int // in seconds
}

// getSerialConfig constructs a serial.Config from SerialProtocol
func (s *SerialProtocol) getSerialConfig() *serial.Config {
	return &serial.Config{
		Name:     s.Port,
		Baud:     s.BaudRate,
		Size:     byte(s.ByteSize),
		Parity:   serial.Parity(s.Parity),
		StopBits: serial.StopBits(s.StopBits),
	}
}

// Connect opens the serial port based on the SerialProtocol configuration
func (s *SerialProtocol) Connect() (*serial.Port, error) {
	config := s.getSerialConfig()
	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}
	return port, nil
}
