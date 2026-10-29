package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/authority module sentinel errors
var (
	ErrUnauthorized         = sdkerrors.Register(ModuleName, 1, "unauthorized")
	ErrInvalidValidator     = sdkerrors.Register(ModuleName, 2, "invalid validator")
	ErrValidatorNotFound    = sdkerrors.Register(ModuleName, 3, "validator not found")
	ErrInvalidPower         = sdkerrors.Register(ModuleName, 4, "invalid power")
	ErrInvalidPubKey        = sdkerrors.Register(ModuleName, 5, "invalid public key")
	ErrValidatorExists      = sdkerrors.Register(ModuleName, 6, "validator already exists")
	ErrEmptyAuthorityList   = sdkerrors.Register(ModuleName, 7, "authority list cannot be empty")
	ErrPowerChangeExceeded  = sdkerrors.Register(ModuleName, 8, "power change exceeded maximum allowed")
)