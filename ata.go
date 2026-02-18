package udisks

type Ata struct {
	SmartSupported                    bool
	SmartEnabled                      bool
	SmartUpdated                      uint64
	SmartFailing                      bool
	SmartPowerOnSeconds               uint64
	SmartTemperature                  float64
	SmartNumAttributesFailing         int32
	SmartNumAttributesFailedInThePast int32
	SmartNumBadSectors                int64
	SmartSelftestStatus               string
	SmartSelftestPercentRemaining     int32
	PmSupported                       bool
	PmEnabled                         bool
	ApmSupported                      bool
	ApmEnabled                        bool
	AamSupported                      bool
	AamEnabled                        bool
	AamVendorRecommendedValue         int32
	WriteCacheSupported               bool
	WriteCacheEnabled                 bool
	ReadLookaheadSupported            bool
	ReadLookaheadEnabled              bool
	SecurityEraseUnitMinutes          int32
	SecurityEnhancedEraseUnitMinutes  int32
	SecurityFrozen                    bool
}
