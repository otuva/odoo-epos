package hwdriver

func (bd *BaseDriver) GetStatus() HWStatus {
	panic("GetStatus must be implemented by subclass")
}
