package types

import (
	context "context"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator interface{}, err error)
	SetValidator(ctx context.Context, validator interface{}) error
	DeleteValidator(ctx context.Context, validator interface{}) error
	GetAllValidators(ctx context.Context) (validators []interface{}, err error)
	ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator interface{}, err error)
	SetValidatorByConsAddr(ctx context.Context, validator interface{}) error
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error
	Unjail(ctx context.Context, consAddr sdk.ConsAddress) error
}