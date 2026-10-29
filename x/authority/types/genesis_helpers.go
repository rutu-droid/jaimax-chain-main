package types

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic validation of genesis data
func (gs *GenesisState) Validate() error {
	if gs == nil || gs.Params == nil {
		return nil
	}
	return gs.Params.Validate()
}
