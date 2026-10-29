package app

import (
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	address "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	codec             codec.Codec
	TxConfig          client.TxConfig
}

func MakeEncodingConfig() EncodingConfig {
	interfaceRegisry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdktypes.GetConfig().GetBech32ConsensusAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdktypes.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})

	if err != nil {
		panic(err)
	}

	codec := codec.NewProtoCodec(interfaceRegisry)

	txconfig := authtx.NewTxConfig(codec, authtx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegisry,
		codec:             codec,
		TxConfig:          txconfig,
	}
}

func GetEncodingConfig() EncodingConfig {
	return MakeEncodingConfig()
}
