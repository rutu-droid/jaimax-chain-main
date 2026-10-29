package keeper

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"example.com/chain/x/authority/types"
)

type StakingKeeper interface {
	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	SetValidator(ctx context.Context, validator stakingtypes.Validator) error
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error
	Unjail(ctx context.Context, consAddr sdk.ConsAddress) error
	SetValidatorByConsAddr(ctx context.Context, validator stakingtypes.Validator) error
	RemoveValidator(ctx context.Context, addr sdk.ValAddress) error
}

type Keeper struct {
	cdc           codec.BinaryCodec
	storeService  store.KVStoreService
	stakingKeeper StakingKeeper
	authority     string
	logger        log.Logger
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	stakingKeeper StakingKeeper,
	authority string,
	logger log.Logger,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeService:  storeService,
		stakingKeeper: stakingKeeper,
		authority:     authority,
		logger:        logger,
	}
}

func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/authority")
}

var ParamsKey = []byte("params")
var PendingValidatorUpdatesKey = []byte("pending_validator_updates")

func (k Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(params)
	if err := store.Set(ParamsKey, bz); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetParams(ctx sdk.Context) *types.Params {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(ParamsKey)
	if err != nil {
		return types.DefaultParams()
	}
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return &params
}

// Track which validators need updates sent to CometBFT
func (k Keeper) SetPendingValidatorUpdate(ctx sdk.Context, valAddr sdk.ValAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(PendingValidatorUpdatesKey, valAddr.Bytes()...)
	return store.Set(key, []byte{1})
}

func (k Keeper) GetPendingValidatorUpdates(ctx sdk.Context) ([]sdk.ValAddress, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Create end key by incrementing the last byte of the prefix
	endKey := make([]byte, len(PendingValidatorUpdatesKey))
	copy(endKey, PendingValidatorUpdatesKey)
	endKey[len(endKey)-1]++

	iterator, err := store.Iterator(PendingValidatorUpdatesKey, endKey)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var addresses []sdk.ValAddress
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Skip the prefix to get the address
		addrBytes := key[len(PendingValidatorUpdatesKey):]
		addresses = append(addresses, sdk.ValAddress(addrBytes))
	}
	return addresses, nil
}

func (k Keeper) DeletePendingValidatorUpdate(ctx sdk.Context, valAddr sdk.ValAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(PendingValidatorUpdatesKey, valAddr.Bytes()...)
	return store.Delete(key)
}

func (k Keeper) ClearPendingValidatorUpdates(ctx sdk.Context) error {
	store := k.storeService.OpenKVStore(ctx)

	// Create end key by incrementing the last byte of the prefix
	endKey := make([]byte, len(PendingValidatorUpdatesKey))
	copy(endKey, PendingValidatorUpdatesKey)
	endKey[len(endKey)-1]++

	iterator, err := store.Iterator(PendingValidatorUpdatesKey, endKey)
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if err := store.Delete(iterator.Key()); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) IsAuthority(ctx sdk.Context, addr sdk.AccAddress) bool {
	params := k.GetParams(ctx)
	addrStr := addr.String()
	for _, authAddr := range params.AuthorityAccounts {
		if authAddr == addrStr {
			return true
		}
	}
	return false
}

func (k Keeper) RequireAuthority(ctx sdk.Context, addr sdk.AccAddress) error {
	if !k.IsAuthority(ctx, addr) {
		return types.ErrUnauthorized.Wrapf("address %s is not an authority account", addr.String())
	}
	return nil
}

func (k Keeper) AddValidator(
	ctx sdk.Context,
	admin sdk.AccAddress,
	valAddr sdk.ValAddress,
	pubKey *codectypes.Any,
	moniker string,
	power int64,
) error {
	if err := k.RequireAuthority(ctx, admin); err != nil {
		return err
	}

	if power <= 0 {
		return types.ErrInvalidPower.Wrapf("power must be positive, got %d", power)
	}

	k.logger.Info("Adding PoA validator",
		"admin", admin.String(),
		"validator", valAddr.String(),
		"power", power,
		"moniker", moniker,
	)

	_, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err == nil {
		return types.ErrValidatorExists.Wrapf("validator %s already exists", valAddr.String())
	}

	var pk cryptotypes.PubKey
	if err := k.cdc.UnpackAny(pubKey, &pk); err != nil {
		return types.ErrInvalidPubKey.Wrapf("failed to unpack pubkey: %s", err)
	}

	if pk == nil {
		return types.ErrInvalidPubKey.Wrap("pubkey is nil after unpacking")
	}

	description := stakingtypes.Description{
		Moniker:         moniker,
		Identity:        "",
		Website:         "",
		SecurityContact: "",
		Details:         "PoA Validator",
	}

	commission := stakingtypes.CommissionRates{
		Rate:          math.LegacyZeroDec(),
		MaxRate:       math.LegacyZeroDec(),
		MaxChangeRate: math.LegacyZeroDec(),
	}

	valAddrStr := valAddr.String()
	validator, err := stakingtypes.NewValidator(valAddrStr, pk, description)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to create validator: %s", err)
	}

	tokens := k.TokensFromPower(power)
	validator.Tokens = tokens
	validator.DelegatorShares = math.LegacyNewDecFromInt(tokens)
	validator.Status = stakingtypes.Bonded
	validator.Commission = stakingtypes.NewCommission(
		commission.Rate,
		commission.MaxRate,
		commission.MaxChangeRate,
	)
	validator.MinSelfDelegation = math.NewInt(1)

	if err := k.stakingKeeper.SetValidator(ctx, validator); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to set validator: %s", err)
	}

	if err := k.stakingKeeper.SetValidatorByConsAddr(ctx, validator); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to set validator by cons addr: %s", err)
	}

	// Mark this validator for update in EndBlocker
	if err := k.SetPendingValidatorUpdate(ctx, valAddr); err != nil {
		return err
	}

	k.logger.Info("Successfully added PoA validator",
		"validator", valAddr.String(),
		"tokens", tokens.String(),
		"power", power,
	)

	return nil
}

func (k Keeper) SetValidatorPower(
	ctx sdk.Context,
	admin sdk.AccAddress,
	valAddr sdk.ValAddress,
	power int64,
) error {
	if err := k.RequireAuthority(ctx, admin); err != nil {
		return err
	}

	if power < 0 {
		return types.ErrInvalidPower.Wrapf("power cannot be negative, got %d", power)
	}

	k.logger.Info("Setting validator power",
		"admin", admin.String(),
		"validator", valAddr.String(),
		"power", power,
	)

	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return types.ErrValidatorNotFound.Wrapf("validator %s not found: %s", valAddr.String(), err)
	}

	oldTokens := validator.Tokens
	newTokens := k.TokensFromPower(power)

	k.logger.Info("Power change details",
		"old_tokens", oldTokens.String(),
		"new_tokens", newTokens.String(),
		"old_power", k.PowerFromTokens(oldTokens),
		"new_power", power,
	)

	validator.Tokens = newTokens
	validator.DelegatorShares = math.LegacyNewDecFromInt(newTokens)

	if power == 0 {
		validator.Status = stakingtypes.Unbonded
		k.logger.Info("Validator unbonded due to zero power", "validator", valAddr.String())
	} else if validator.Status != stakingtypes.Bonded {
		validator.Status = stakingtypes.Bonded
		k.logger.Info("Validator bonded", "validator", valAddr.String())
	}

	if err := k.stakingKeeper.SetValidator(ctx, validator); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to update validator: %s", err)
	}

	// Mark this validator for update in EndBlocker
	if err := k.SetPendingValidatorUpdate(ctx, valAddr); err != nil {
		return err
	}

	k.logger.Info("Successfully updated validator power",
		"validator", valAddr.String(),
		"new_power", power,
	)

	return nil
}

func (k Keeper) RemoveValidator(
	ctx sdk.Context,
	admin sdk.AccAddress,
	valAddr sdk.ValAddress,
) error {
	if err := k.RequireAuthority(ctx, admin); err != nil {
		return err
	}

	k.logger.Info("Removing validator",
		"admin", admin.String(),
		"validator", valAddr.String(),
	)

	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return types.ErrValidatorNotFound.Wrapf("validator %s not found: %s", valAddr.String(), err)
	}

	// Set power to 0 and unbond
	validator.Tokens = math.ZeroInt()
	validator.DelegatorShares = math.LegacyZeroDec()
	validator.Status = stakingtypes.Unbonded

	if err := k.stakingKeeper.SetValidator(ctx, validator); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("failed to unbond validator: %s", err)
	}

	// Mark this validator for update in EndBlocker
	if err := k.SetPendingValidatorUpdate(ctx, valAddr); err != nil {
		return err
	}

	k.logger.Info("Successfully set validator to be removed (power=0)",
		"validator", valAddr.String(),
		"note", "Validator will be removed from active set in next block",
	)

	return nil
}

func (k Keeper) UpdateAuthorityAccounts(
	ctx sdk.Context,
	admin sdk.AccAddress,
	newAccounts []string,
) error {
	if err := k.RequireAuthority(ctx, admin); err != nil {
		return err
	}

	if len(newAccounts) == 0 {
		return types.ErrEmptyAuthorityList
	}

	params := k.GetParams(ctx)
	params.AuthorityAccounts = newAccounts

	k.logger.Info("Updating authority accounts",
		"admin", admin.String(),
		"new_accounts", newAccounts,
	)

	return k.SetParams(ctx, params)
}

func (k Keeper) TokensFromPower(power int64) math.Int {
	return math.NewInt(power).Mul(math.NewInt(1_000_000))
}

func (k Keeper) PowerFromTokens(tokens math.Int) int64 {
	return tokens.Quo(math.NewInt(1_000_000)).Int64()
}

func (k Keeper) ValidatePowerChange(oldPower, newPower int64, maxChangePercent int64) error {
	if oldPower == 0 {
		return nil
	}

	diff := newPower - oldPower
	if diff < 0 {
		diff = -diff
	}

	maxChange := (oldPower * maxChangePercent) / 100
	if diff > maxChange {
		return types.ErrPowerChangeExceeded.Wrapf(
			"power change %d exceeds maximum allowed %d (%d%%)",
			diff, maxChange, maxChangePercent,
		)
	}

	return nil
}

func (k Keeper) EndBlocker(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	// Only send updates for validators that have changed
	pendingAddrs, err := k.GetPendingValidatorUpdates(ctx)
	if err != nil {
		k.logger.Error("failed to get pending validator updates", "error", err)
		return nil, err
	}

	if len(pendingAddrs) == 0 {
		// No updates needed
		return []abci.ValidatorUpdate{}, nil
	}

	var updates []abci.ValidatorUpdate
	var toRemove []sdk.ValAddress

	for _, valAddr := range pendingAddrs {
		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			k.logger.Error("failed to get validator for update", "validator", valAddr.String(), "error", err)
			continue
		}

		power := k.PowerFromTokens(validator.Tokens)

		tmPk, err := validator.CmtConsPublicKey()
		if err != nil {
			k.logger.Error("failed to get validator consensus pubkey", "error", err)
			continue
		}

		update := abci.ValidatorUpdate{
			PubKey: tmPk,
			Power:  power,
		}

		updates = append(updates, update)

		k.logger.Info("EndBlock validator update",
			"validator", validator.OperatorAddress,
			"power", power,
			"status", validator.Status,
		)

		// If power is 0, we can remove the validator from state in the NEXT block
		if power == 0 {
			toRemove = append(toRemove, valAddr)
		}
	}

	// Clear pending updates for this block
	if err := k.ClearPendingValidatorUpdates(ctx); err != nil {
		k.logger.Error("failed to clear pending validator updates", "error", err)
	}

	// Schedule removal of validators with power=0 for next block's BeginBlocker
	for _, valAddr := range toRemove {
		// You could store these in a separate key to be processed in BeginBlocker
		// For now, we'll just log it
		k.logger.Info("Validator scheduled for removal in next block", "validator", valAddr.String())
	}

	k.logger.Info("EndBlock returning validator updates", "count", len(updates))
	return updates, nil
}
