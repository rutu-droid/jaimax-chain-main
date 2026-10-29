package authority

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"example.com/chain/x/authority/keeper"
	"example.com/chain/x/authority/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	k.Logger(ctx).Info("Authority module genesis initialized",
		"authority_accounts", data.Params.AuthorityAccounts,
	)
}

func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

func DefaultGenesis() *types.GenesisState {
	return types.DefaultGenesisState()
}

func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
