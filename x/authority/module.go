package authority

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store" // Use core store service
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"example.com/chain/x/authority/client/cli"
	"example.com/chain/x/authority/keeper"
	modulev1 "example.com/chain/x/authority/module/v1"
	"example.com/chain/x/authority/types"
)

// Interface Assertions
var (
	_ module.AppModuleBasic  = AppModuleBasic{}
	_ module.HasGenesis      = AppModule{}
	_ appmodule.AppModule    = AppModule{}
	_ module.HasServices     = AppModule{}
	_ module.HasABCIEndBlock = AppModule{}
)

// Register for depinject
func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
	// 	_ = err
	// }
}

func (amb AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}
	return ValidateGenesis(&genesis)
}

func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

func (AppModule) IsOnePerModuleType() {}
func (AppModule) IsAppModule()        {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)
	InitGenesis(ctx, am.keeper, &genesis)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genesis)
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.keeper.EndBlocker(sdkCtx)
}

// Depinject support

type ModuleInputs struct {
	depinject.In

	Config        *modulev1.Module
	Cdc           codec.Codec
	StoreService  store.KVStoreService // Changed from storetypes.StoreKey
	StakingKeeper keeper.StakingKeeper
	Logger        log.Logger
}

type ModuleOutputs struct {
	depinject.Out

	AuthorityKeeper keeper.Keeper
	Module          appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// Default authority to governance module address if not set
	authority := in.Config.Authority
	if authority == "" {
		authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService, // This now matches the updated keeper
		in.StakingKeeper,
		authority,
		in.Logger,
	)

	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{
		AuthorityKeeper: k,
		Module:          m,
	}
}
