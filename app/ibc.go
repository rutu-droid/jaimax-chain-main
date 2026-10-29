package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ica "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"

	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferv2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/v2"
	ibcapi "github.com/cosmos/ibc-go/v10/modules/core/api"
	solomachine "github.com/cosmos/ibc-go/v10/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

func (app *App) registerIBCModules(appOpts servertypes.AppOptions) error {
	// setup module store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(ibctransfertypes.StoreKey),
		storetypes.NewKVStoreKey(icahosttypes.StoreKey),
		storetypes.NewKVStoreKey(icacontrollertypes.StoreKey),
	); err != nil {
		return err
	}

	authority := authtypes.NewModuleAddress("ibc-admin").String()

	app.IBCKeeper = ibckeeper.NewKeeper(
		app.codec,
		runtime.NewKVStoreService(app.GetKey(ibcexported.StoreKey)),
		nil,
		app.UpgradeKeeper,
		authority,
	)

	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		app.codec,
		runtime.NewKVStoreService(app.GetKey(ibctransfertypes.StoreKey)),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AuthKeeper,
		app.BankKeeper,
		authority,
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		app.codec,
		runtime.NewKVStoreService(app.GetKey(icahosttypes.StoreKey)),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.AuthKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		authority,
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		app.codec,
		runtime.NewKVStoreService(app.GetKey(icacontrollertypes.StoreKey)),
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		authority,
	)

	var (
		transferStack      porttypes.IBCModule  = ibctransfer.NewIBCModule(app.TransferKeeper)
		transferStackV2    ibcapi.IBCModule     = ibctransferv2.NewIBCModule(app.TransferKeeper)
		icaControllerStack porttypes.Middleware = icacontroller.NewIBCMiddleware(app.ICAControllerKeeper)
		icaHostStack       porttypes.IBCModule  = icahost.NewIBCModule(app.ICAHostKeeper)
	)

	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack)

	ibcRouterV2 := ibcapi.NewRouter().AddRoute(ibctransfertypes.ModuleName, transferStackV2)

	app.IBCKeeper.SetRouter(ibcRouter)
	app.IBCKeeper.SetRouterV2(ibcRouterV2)

	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := clientKeeper.GetStoreProvider()

	tmLightClient := ibctm.NewLightClientModule(app.codec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, tmLightClient)

	soloLightClient := solomachine.NewLightClientModule(app.codec, storeProvider)
	clientKeeper.AddRoute(solomachine.ModuleName, soloLightClient)

	if err := app.RegisterModules(
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(app.TransferKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		ibctm.NewAppModule(tmLightClient),
		solomachine.NewAppModule(soloLightClient),
	); err != nil {
		return err
	}

	return nil
}

// Registering IBC modules since it doesn't suuport depinject

func RegisterIBC(cdc codec.Codec) map[string]appmodule.AppModule {
	modules := map[string]appmodule.AppModule{
		ibcexported.ModuleName:      ibc.NewAppModule(&ibckeeper.Keeper{}),
		ibctransfertypes.ModuleName: ibctransfer.NewAppModule(ibctransferkeeper.Keeper{}),
		icatypes.ModuleName:         ica.NewAppModule(&icacontrollerkeeper.Keeper{}, &icahostkeeper.Keeper{}),
		ibctm.ModuleName:            ibctm.NewAppModule(ibctm.NewLightClientModule(cdc, ibcclienttypes.StoreProvider{})),
		solomachine.ModuleName:      solomachine.NewAppModule(solomachine.NewLightClientModule(cdc, ibcclienttypes.StoreProvider{})),
	}

	for _, m := range modules {
		if mr, ok := m.(module.AppModuleBasic); ok {
			mr.RegisterInterfaces(cdc.InterfaceRegistry())
		}
	}

	return modules
}

func (app *App) GetIBC() *ibckeeper.Keeper {
	return app.IBCKeeper
}
