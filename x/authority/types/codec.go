package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers the necessary x/authority interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddValidator{}, "authority/MsgAddValidator", nil)
	cdc.RegisterConcrete(&MsgRemoveValidator{}, "authority/MsgRemoveValidator", nil)
	cdc.RegisterConcrete(&MsgSetValidatorPower{}, "authority/MsgSetValidatorPower", nil)
	cdc.RegisterConcrete(&MsgUpdateAuthorityAccounts{}, "authority/MsgUpdateAuthorityAccounts", nil)
}

// RegisterInterfaces registers the x/authority interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddValidator{},
		&MsgRemoveValidator{},
		&MsgSetValidatorPower{},
		&MsgUpdateAuthorityAccounts{},
	)
}
