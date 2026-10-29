package types

const (
	// ModuleName defines the module name
	ModuleName = "authority"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_authority"
)

// KVStore keys
var (
	ParamsKeyPrefix = []byte{0x01} // key prefix for module parameters
)