package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default parameters
func DefaultParams() *Params {
	return &Params{
		AuthorityAccounts: []string{},
	}
}

// Validate validates the set of params
func (p *Params) Validate() error {
	if p == nil {
		return nil
	}
	
	for _, addr := range p.AuthorityAccounts {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return err
		}
	}
	return nil
}
