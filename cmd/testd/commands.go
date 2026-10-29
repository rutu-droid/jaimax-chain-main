package cmd

import (
	"errors"
	"io"

	"example.com/chain/app"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	basicModules map[string]appmodule.AppModule,
) {
	// Build a basic manager, but we need to handle genutil specially
	basicManager := module.NewBasicManagerFromManager(module.NewManagerFromMap(basicModules), map[string]module.AppModuleBasic{
		genutiltypes.ModuleName: genutil.AppModuleBasic{GenTxValidator: genutiltypes.DefaultMessageValidator},
	})

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	server.AddCommandsWithStartCmdOptions(rootCmd, app.DefaultNodeHome, newApp, appExport, server.StartCmdOptions{
		AddFlags: addModuleInitFlags,
	})

	rootCmd.AddCommand(
		server.StatusCommand(),
		genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlockResultsCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlocksCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transaction subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

func newApp(
	logger log.Logger,
	db db.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.New(
		logger, db, traceStore, true, appOpts, baseappOptions...,
	)
}

func appExport(
	logger log.Logger,
	db db.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddress []string,
	appOpts servertypes.AppOptions,
	moduleToExport []string,
) (servertypes.ExportedApp, error) {
	var bApp *app.App

	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home is not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	if height != -1 {
		bApp = app.New(logger, db, traceStore, false, viperAppOpts)
		if err := bApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		bApp = app.New(logger, db, traceStore, true, viperAppOpts)
	}

	return bApp.ExportStateAndValidators(forZeroHeight, jailAllowedAddress, moduleToExport)
}
