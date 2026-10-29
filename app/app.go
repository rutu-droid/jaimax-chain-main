package app

import (
	// stdctx "context"

	"io"

	"cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	// upgradetypes "cosmossdk.io/x/upgrade/types"

	abci "github.com/cometbft/cometbft/abci/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	authoritykeeper "example.com/chain/x/authority/keeper"
	// icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	// icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	// ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	// ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

const (
	Name = "jaimax"

	AccountAddressPrefix = "jaimax"

	ChainCoinType = 118

	// Upgrade name - must match what you use in the upgrade proposal
	// UpgradeName = "ibc-v1"
)

var DefaultNodeHome string

var (
	_ servertypes.ABCI = (*App)(nil)
)

type App struct {
	*runtime.App
	codec            codec.Codec
	txConfig         client.TxConfig
	interfaceRegisry codectypes.InterfaceRegistry

	AuthKeeper         authkeeper.AccountKeeper
	BankKeeper         bankkeeper.Keeper
	StakingKeeper      *stakingkeeper.Keeper
	UpgradeKeeper      *upgradekeeper.Keeper
	DistributionKeeper distributionkeeper.Keeper
	GovKeeper          *govkeeper.Keeper
	ConsensusKeeper    consensuskeeper.Keeper

	// Add your authority keeper
	AuthorityKeeper authoritykeeper.Keeper

	// IBC Keepers
	IBCKeeper           *ibckeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
}

func init() {
	var err error
	helpers.EnvPrefix = Name
	DefaultNodeHome, err = helpers.GetNodeHomeDirectory("." + Name)
	// fmt.Println(DefaultNodeHome)
	if err != nil {
		panic(err)
	}
}

func AppConfig() depinject.Config {
	return depinject.Configs(
		appConfig,
		depinject.Supply(
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.AppModuleBasic{GenTxValidator: genutiltypes.DefaultMessageValidator},
			},
		),
		depinject.Provide(ProvideAnteHandler),
	)
}

func New(
	logger log.Logger,
	db db.DB,
	tracestore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	var (
		app        = &App{}
		appBuilder *runtime.AppBuilder

		appConfig = depinject.Configs(
			AppConfig(),
			depinject.Supply(
				appOpts,
				logger,

				app.GetIBC,
			),
		)
	)

	var appModules map[string]appmodule.AppModule
	var anteHandler sdk.AnteHandler

	if err := depinject.Inject(appConfig,
		&appBuilder,
		&appModules,
		&app.txConfig,
		&app.codec,
		&app.interfaceRegisry,
		&app.AuthKeeper,
		&app.BankKeeper,
		&app.ConsensusKeeper,
		&app.DistributionKeeper,
		&app.GovKeeper,
		&app.StakingKeeper,
		&app.UpgradeKeeper,
		&app.AuthorityKeeper,
		&anteHandler,
	); err != nil {
		panic(err)
	}

	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution(),
		func(bapp *baseapp.BaseApp) {
			bapp.SetAnteHandler(anteHandler)
		},
	)

	app.App = appBuilder.Build(db, tracestore, baseAppOptions...)

	if err := app.registerIBCModules(appOpts); err != nil {
		logger.Error("Failed to register IBC modules", "error", err)
		panic(err)
	}

	// upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	// if err != nil {
	// 	panic(err)
	// }

	// if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
	// 	logger.Info("Applying store upgrades", "name", upgradeInfo.Name, "height", upgradeInfo.Height)

	// 	app.SetStoreLoader(
	// 		upgradetypes.UpgradeStoreLoader(
	// 			upgradeInfo.Height,
	// 			&storetypes.StoreUpgrades{
	// 				Added: []string{
	// 					ibcexported.StoreKey,
	// 					ibctransfertypes.StoreKey,
	// 					icahosttypes.StoreKey,
	// 					icacontrollertypes.StoreKey,
	// 				},
	// 			},
	// 		),
	// 	)
	// }

	// Set InitChainer - this runs when the chain first starts
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
			return nil, err
		}
		return app.App.InitChainer(ctx, req)
	})

	// CRITICAL: Register the upgrade handler
	// This is what executes when the chain reaches the upgrade height
	// app.UpgradeKeeper.SetUpgradeHandler(
	// 	UpgradeName,
	// 	func(ctx stdctx.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
	// 		logger.Info("Starting IBC upgrade", "name", plan.Name, "height", plan.Height)

	// 		sdkCtx := sdk.UnwrapSDKContext(ctx)

	// TODO: When you're ready to actually add IBC modules:
	// 1. Uncomment the IBC registration code in ibc.go
	// 2. Uncomment the line below to register IBC modules
	// 3. Add the IBC keeper initialization
	//
	// Example:

	// Run migrations for existing modules
	// This ensures all modules are updated to their latest version
	// 		newVM, err := app.ModuleManager.RunMigrations(sdkCtx, app.Configurator(), vm)
	// 		if err != nil {
	// 			logger.Error("Failed to run module migrations", "error", err)
	// 			return nil, err
	// 		}

	// 		logger.Info("IBC upgrade completed successfully", "name", plan.Name)
	// 		return newVM, nil
	// 	},
	// )

	// Load the app state
	if err := app.Load(loadLatest); err != nil {
		logger.Error("Failed to load app", "error", err)
		panic(err)
	}

	// REMOVED: The problematic consensus params check
	// This was causing the nil pointer dereference because
	// the context cannot be created until after the chain starts

	return app
}

func (app *App) AppCodec() codec.Codec {
	return app.codec
}

func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegisry
}

func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	kvStoreKey, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)

	if !ok {
		return nil
	}

	return kvStoreKey
}

func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)

	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string)
	for _, perms := range moduleAccPerms {
		dup[perms.GetAccount()] = perms.GetPermissions()
	}

	return dup
}

func blockedAddresses() map[string]bool {
	result := make(map[string]bool)

	if len(blockAccAddr) > 0 {
		for _, addr := range blockAccAddr {
			result[addr] = true
		}
	} else {
		for addr := range GetMaccPerms() {
			result[addr] = true
		}
	}

	return result
}
