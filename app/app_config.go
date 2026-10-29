package app

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	consensusmodulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	distributionmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	govmodulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"

	"cosmossdk.io/depinject/appconfig"

	_ "cosmossdk.io/x/upgrade" // import for side-effects
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import for side-effects
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus" // import for side-effects
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import for side-effects
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // import for side-effects
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	_ "github.com/cosmos/cosmos-sdk/x/gov" //import for side effects
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	// _ "github.com/cosmos/cosmos-sdk/x/slashing" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import for side-effects
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	// Import your authority module
	_ "example.com/chain/x/authority"
	authoritymodulev1 "example.com/chain/x/authority/module/v1"
	authoritytypes "example.com/chain/x/authority/types"
)

var (
	moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
		{Account: authtypes.FeeCollectorName},
		{Account: distrtypes.ModuleName},
		{Account: stakingtypes.BondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: stakingtypes.NotBondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: govtypes.ModuleName, Permissions: []string{authtypes.Burner}},
		{Account: ibctransfertypes.ModuleName, Permissions: []string{authtypes.Minter, authtypes.Burner}},
		{Account: icatypes.ModuleName},
	}

	authority = sdk.MustBech32ifyAddressBytes(AccountAddressPrefix, authtypes.NewModuleAddress(govtypes.ModuleName))

	blockAccAddr = []string{
		authtypes.FeeCollectorName,
		stakingtypes.BondedPoolName,
		stakingtypes.NotBondedPoolName,
		distrtypes.ModuleName,
	}

	appConfig = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: runtime.ModuleName,
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: Name,

					PreBlockers: []string{
						authtypes.ModuleName,
						upgradetypes.ModuleName,
					},

					BeginBlockers: []string{
						upgradetypes.ModuleName,
						stakingtypes.ModuleName,
						distrtypes.ModuleName,
						ibcexported.ModuleName,
						authoritytypes.ModuleName,
					},

					EndBlockers: []string{
						govtypes.ModuleName,
						stakingtypes.ModuleName,
						authoritytypes.ModuleName, // Add authority module to EndBlockers
					},

					OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
						{
							ModuleName: authtypes.ModuleName,
							KvStoreKey: "acc",
						},
					},

					InitGenesis: []string{
						authtypes.ModuleName,
						banktypes.ModuleName,
						distrtypes.ModuleName,
						stakingtypes.ModuleName,
						govtypes.ModuleName,
						authoritytypes.ModuleName, // Add authority module to genesis
						genutiltypes.ModuleName,
						upgradetypes.ModuleName,
						//IBC
						ibcexported.ModuleName,
						ibctransfertypes.ModuleName,
						icatypes.ModuleName,
					},
				}),
			},
			{
				Name: authtypes.ModuleName,
				Config: appconfig.WrapAny(&authmodulev1.Module{
					Bech32Prefix:             AccountAddressPrefix,
					ModuleAccountPermissions: moduleAccPerms,
				}),
			},
			{
				Name: banktypes.ModuleName,
				Config: appconfig.WrapAny(&bankmodulev1.Module{
					BlockedModuleAccountsOverride: blockAccAddr,
				}),
			},
			{
				Name:   distrtypes.ModuleName,
				Config: appconfig.WrapAny(&distributionmodulev1.Module{}),
			},
			{
				Name:   stakingtypes.ModuleName,
				Config: appconfig.WrapAny(&stakingmodulev1.Module{}),
			},
			{
				Name: upgradetypes.ModuleName,
				Config: appconfig.WrapAny(&upgrademodulev1.Module{
					Authority: authority,
				}),
			},
			{
				Name:   genutiltypes.ModuleName,
				Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
			},
			{
				Name:   "tx",
				Config: appconfig.WrapAny(&txconfigv1.Config{}),
			},
			{
				Name:   consensustypes.ModuleName,
				Config: appconfig.WrapAny(&consensusmodulev1.Module{}),
			},
			{
				Name: govtypes.ModuleName,
				Config: appconfig.WrapAny(&govmodulev1.Module{
					Authority: authority,
				}),
			},
			// Add authority module configuration
			{
				Name: authoritytypes.ModuleName,
				Config: appconfig.WrapAny(&authoritymodulev1.Module{
					Authority: authority,
				}),
			},
		},
	})
)
