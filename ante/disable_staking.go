package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DisableStakingDecorator blocks all public staking transactions (except during genesis)
type DisableStakingDecorator struct{}

// NewDisableStakingDecorator creates a new DisableStakingDecorator
func NewDisableStakingDecorator() DisableStakingDecorator {
	return DisableStakingDecorator{}
}

// AnteHandle implements the AnteDecorator interface
func (d DisableStakingDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	// ✅ ALLOW STAKING DURING GENESIS (block height 0)
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		switch msg.(type) {
		case *stakingtypes.MsgDelegate:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"public delegation is disabled in PoA mode",
			)
		case *stakingtypes.MsgUndelegate:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"public undelegation is disabled in PoA mode",
			)
		case *stakingtypes.MsgBeginRedelegate:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"public redelegation is disabled in PoA mode",
			)
		case *stakingtypes.MsgCancelUnbondingDelegation:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"canceling unbonding is disabled in PoA mode",
			)
		case *stakingtypes.MsgCreateValidator:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"public validator creation is disabled - use authority module",
			)
		}
	}

	return next(ctx, tx, simulate)
}

// Optional version with authority exceptions
type AuthorityKeeper interface {
	IsAuthority(ctx sdk.Context, addr sdk.AccAddress) bool
}

type DisableStakingWithExceptionsDecorator struct {
	authorityKeeper AuthorityKeeper
}

func NewDisableStakingWithExceptionsDecorator(ak AuthorityKeeper) DisableStakingWithExceptionsDecorator {
	return DisableStakingWithExceptionsDecorator{
		authorityKeeper: ak,
	}
}

func (d DisableStakingWithExceptionsDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	// ✅ ALLOW STAKING DURING GENESIS (block height 0)
	if ctx.BlockHeight() == 0 {
		return next(ctx, tx, simulate)
	}

	// Get signers from the messages themselves
	msgs := tx.GetMsgs()
	if len(msgs) > 0 {
		// Try to get the first signer from the first message
		if msg, ok := msgs[0].(interface{ GetSigners() []sdk.AccAddress }); ok {
			signers := msg.GetSigners()
			if len(signers) > 0 && d.authorityKeeper.IsAuthority(ctx, signers[0]) {
				return next(ctx, tx, simulate)
			}
		}
	}

	// Block staking messages for non-authority accounts
	for _, msg := range msgs {
		switch msg.(type) {
		case *stakingtypes.MsgDelegate,
			*stakingtypes.MsgUndelegate,
			*stakingtypes.MsgBeginRedelegate,
			*stakingtypes.MsgCancelUnbondingDelegation,
			*stakingtypes.MsgCreateValidator:
			return ctx, sdkerrors.ErrUnauthorized.Wrap(
				"staking operations restricted to authority accounts",
			)
		}
	}

	return next(ctx, tx, simulate)
}