package eth

type ID string

const (
	IDInvalid   ID = "invalid"
	IDHead      ID = "head"
	IDGenesis   ID = "genesis"
	IDFinalized ID = "finalized"
	IDSlot      ID = "slot"
	IDRoot      ID = "root"
)
