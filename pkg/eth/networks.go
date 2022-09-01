package eth

func DefaultNetworkIDMap() map[uint64]string {
	return map[uint64]string{
		1:        "mainnet",
		3:        "ropsten",
		4:        "rinkeby",
		5:        "goerli",
		1337802:  "kiln",
		11155111: "sepolia",
	}
}

func GetNetworkName(networkID uint64) string {
	name, exists := DefaultNetworkIDMap()[networkID]
	if !exists {
		return "unknown"
	}

	return name
}
