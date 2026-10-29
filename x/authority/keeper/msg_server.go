package keeper

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"example.com/chain/x/authority/types"
)

type msgServer struct {
	Keeper
	types.UnimplementedMsgServer
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// Convert anypb.Any to codectypes.Any
	var pubKey *codectypes.Any
	if msg.PubKey != nil {
		pubKey = &codectypes.Any{
			TypeUrl: msg.PubKey.TypeUrl,
			Value:   msg.PubKey.Value,
		}
	}

	if err := k.Keeper.AddValidator(ctx, admin, valAddr, pubKey, msg.Moniker, msg.Power); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"authority_add_validator",
			sdk.NewAttribute("admin", msg.Admin),
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("moniker", msg.Moniker),
		),
	)

	return &types.MsgAddValidatorResponse{}, nil
}

func (k msgServer) SetValidatorPower(goCtx context.Context, msg *types.MsgSetValidatorPower) (*types.MsgSetValidatorPowerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.SetValidatorPower(ctx, admin, valAddr, msg.Power); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"authority_set_validator_power",
			sdk.NewAttribute("admin", msg.Admin),
			sdk.NewAttribute("validator", msg.ValidatorAddress),
		),
	)

	return &types.MsgSetValidatorPowerResponse{}, nil
}

func (k msgServer) RemoveValidator(goCtx context.Context, msg *types.MsgRemoveValidator) (*types.MsgRemoveValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.RemoveValidator(ctx, admin, valAddr); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"authority_remove_validator",
			sdk.NewAttribute("admin", msg.Admin),
			sdk.NewAttribute("validator", msg.ValidatorAddress),
		),
	)

	return &types.MsgRemoveValidatorResponse{}, nil
}

func (k msgServer) UpdateAuthorityAccounts(goCtx context.Context, msg *types.MsgUpdateAuthorityAccounts) (*types.MsgUpdateAuthorityAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.UpdateAuthorityAccounts(ctx, admin, msg.NewAuthorityAccounts); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"authority_update_accounts",
			sdk.NewAttribute("admin", msg.Admin),
		),
	)

	return &types.MsgUpdateAuthorityAccountsResponse{}, nil
}
