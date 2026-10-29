package app

import (
	"fmt"

	"cosmossdk.io/depinject"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	authorityante "example.com/chain/x/authority/ante"
	authoritykeeper "example.com/chain/x/authority/keeper"
)

// AnteHandlerInputs defines the inputs needed for creating the ante handler
type AnteHandlerInputs struct {
	depinject.In

	AccountKeeper ante.AccountKeeper
	BankKeeper    authtypes.BankKeeper
	// SignModeHandler *txsigning.HandlerMap
	// SigGasConsumer  func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	FeegrantKeeper  ante.FeegrantKeeper `optional:"true"`
	AuthorityKeeper authoritykeeper.Keeper
	TxConfig        client.TxConfig
}

// AnteHandlerOutputs defines the outputs from creating the ante handler
type AnteHandlerOutputs struct {
	depinject.Out

	AnteHandler sdk.AnteHandler
}

// ProvideAnteHandler provides the ante handler for the app
func ProvideAnteHandler(in AnteHandlerInputs) AnteHandlerOutputs {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			AccountKeeper:   in.AccountKeeper,
			BankKeeper:      in.BankKeeper,
			SignModeHandler: in.TxConfig.SignModeHandler(),
			FeegrantKeeper:  in.FeegrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			AuthorityKeeper: in.AuthorityKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	return AnteHandlerOutputs{
		AnteHandler: anteHandler,
	}
}

// HandlerOptions are the options required for constructing a default SDK AnteHandler
type HandlerOptions struct {
	AccountKeeper   ante.AccountKeeper
	BankKeeper      authtypes.BankKeeper
	SignModeHandler *txsigning.HandlerMap
	FeegrantKeeper  ante.FeegrantKeeper
	SigGasConsumer  func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
	AuthorityKeeper authoritykeeper.Keeper
}

// NewAnteHandler creates the ante handler chain
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, fmt.Errorf("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, fmt.Errorf("bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, fmt.Errorf("sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(nil),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, nil),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),

		// 🔥 ADD YOUR CUSTOM DECORATOR HERE
		authorityante.NewDisableStakingWithExceptionsDecorator(options.AuthorityKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
