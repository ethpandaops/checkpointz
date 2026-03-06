package eth

type ID string

const (
	IDInvalid    ID = "invalid"
	IDHead       ID = "head"
	IDGenesis    ID = "genesis"
	IDFinalized  ID = "finalized"
	IDCheckpoint ID = "checkpoint"
	IDSlot       ID = "slot"
	IDRoot       ID = "root"
)
