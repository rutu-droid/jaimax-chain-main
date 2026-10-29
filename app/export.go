package app

import (
	"encoding/json"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func (app *App) ExportStateAndValidators(forZeroHeight bool, jailAllowedAddr, modulesToExport []string) (servertypes.ExportedApp, error) {
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		return servertypes.ExportedApp{}, nil
	}

	genState, err := app.ModuleManager.ExportGenesisForModules(ctx, app.codec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", " ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)

	return servertypes.ExportedApp{
		AppState:   appState,
		Validators: validators,
		Height:     height,
	}, err

}
