package hwdriver

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
}

func (s *SerialScaleDriver) GetStatus() HWStatus {
	return s.status
}

// get weight in kg
func (s *SerialScaleDriver) GetWeight() float64 {
	return s.weight
}

// get weight in g
func (s *SerialScaleDriver) GetWeightInGrams() float64 {
	return s.weight * 1000
}

// get weight in lb
func (s *SerialScaleDriver) GetWeightInPounds() float64 {
	return s.weight * 2.20462
}
