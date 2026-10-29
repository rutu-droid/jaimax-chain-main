package cmd

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	return cfg
}

func initAppConfig() (string, any) {
	type CustomAppConfig struct {
		serverconfig.Config `mapstructure:",squash"`
	}

	srvcfg := serverconfig.DefaultConfig()

	customAppConfig := CustomAppConfig{
		Config: *srvcfg,
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}
