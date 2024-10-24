package beacon

type OperatingMode string

const (
	OperatingModeFull  OperatingMode = "full"
	OperatingModeLight OperatingMode = "light"
)

type LightClientMode string

const (
	LightClientModeProxy LightClientMode = "proxy"
)
