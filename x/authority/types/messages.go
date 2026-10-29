package types

import (
	errorsmod "cosmossdk.io/errors"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message type URLs
const (
	TypeMsgAddValidator             = "add_validator"
	TypeMsgSetValidatorPower        = "set_validator_power"
	TypeMsgRemoveValidator          = "remove_validator"
	TypeMsgUpdateAuthorityAccounts  = "update_authority_accounts"
)

// ValidateBasic performs basic validation of MsgAddValidator
func (m *MsgAddValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address: %s", err)
	}
	
	if _, err := sdk.ValAddressFromBech32(m.ValidatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}
	
	if m.PubKey == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "public key cannot be empty")
	}
	
	if m.Moniker == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "moniker cannot be empty")
	}
	
	if m.Power <= 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "power must be positive, got %d", m.Power)
	}
	
	return nil
}

// ValidateBasic performs basic validation of MsgSetValidatorPower
func (m *MsgSetValidatorPower) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address: %s", err)
	}
	
	if _, err := sdk.ValAddressFromBech32(m.ValidatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}
	
	if m.Power < 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "power cannot be negative, got %d", m.Power)
	}
	
	return nil
}

// ValidateBasic performs basic validation of MsgRemoveValidator
func (m *MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address: %s", err)
	}
	
	if _, err := sdk.ValAddressFromBech32(m.ValidatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address: %s", err)
	}
	
	return nil
}

// ValidateBasic performs basic validation of MsgUpdateAuthorityAccounts
func (m *MsgUpdateAuthorityAccounts) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address: %s", err)
	}
	
	if len(m.NewAuthorityAccounts) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "authority accounts list cannot be empty")
	}
	
	// Validate each authority account address
	for _, addr := range m.NewAuthorityAccounts {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority account address %s: %s", addr, err)
		}
	}
	
	return nil
}