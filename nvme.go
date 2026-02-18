package udisks

type NVMeController struct {
	State                         string
	ControllerID                  uint16
	SubsystemNQN                  []byte
	FGUID                         string
	NVMeRevision                  string
	UnallocatedCapacity           uint64
	SmartUpdated                  uint64
	SmartCriticalWarning          []string
	SmartPowerOnHours             uint64
	SmartTemperature              uint16
	SmartSelftestStatus           string
	SmartSelftestPercentRemaining int32
	SanitizeStatus                string
	SanitizePercentRemaining      int32
}
