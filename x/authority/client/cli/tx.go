package cli

import (
	//"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	//codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	//cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/anypb"

	"example.com/chain/x/authority/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdAddValidator(),
		CmdSetValidatorPower(),
		CmdRemoveValidator(),
		CmdUpdateAuthorityAccounts(),
	)

	return cmd
}

// func CmdAddValidator() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "add-validator [validator-address] [pubkey-json] [moniker] [power]",
// 		Short: "Add a new PoA validator",
// 		Args:  cobra.ExactArgs(4),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			valAddr, err := sdk.ValAddressFromBech32(args[0])
// 			if err != nil {
// 				return fmt.Errorf("invalid validator address: %w", err)
// 			}

// 			power, err := strconv.ParseInt(args[3], 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid power: %w", err)
// 			}

// 			// Parse the pubkey JSON
// 			var pkAny codectypes.Any
// 			if err := json.Unmarshal([]byte(args[1]), &pkAny); err != nil {
// 				return fmt.Errorf("failed to unmarshal pubkey JSON: %w", err)
// 			}

// 			// Unpack to verify it's valid
// 			var pk cryptotypes.PubKey
// 			if err := clientCtx.Codec.UnpackAny(&pkAny, &pk); err != nil {
// 				return fmt.Errorf("failed to unpack pubkey: %w", err)
// 			}

// 			// Convert to anypb.Any for the message
// 			pubKeyAny := &anypb.Any{
// 				TypeUrl: pkAny.TypeUrl,
// 				Value:   pkAny.Value,
// 			}

// 			msg := &types.MsgAddValidator{
// 				Admin:            clientCtx.GetFromAddress().String(),
// 				ValidatorAddress: valAddr.String(),
// 				PubKey:           pubKeyAny,
// 				Moniker:          args[2],
// 				Power:            power,
// 			}

// 			if err := msg.ValidateBasic(); err != nil {
// 				return err
// 			}

// 			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
// 		},
// 	}

// 	flags.AddTxFlagsToCmd(cmd)
// 	return cmd
// }

func CmdAddValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validator [validator-address] [pubkey-json] [moniker] [power]",
		Short: "Add a new PoA validator",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid validator address: %w", err)
			}

			power, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid power: %w", err)
			}

			// Parse pubkey using the codec's JSON unmarshaler
			var pubKeyAny anypb.Any
			if err := clientCtx.Codec.UnmarshalJSON([]byte(args[1]), &pubKeyAny); err != nil {
				return fmt.Errorf("failed to unmarshal pubkey JSON: %w", err)
			}

			msg := &types.MsgAddValidator{
				Admin:            clientCtx.GetFromAddress().String(),
				ValidatorAddress: valAddr.String(),
				PubKey:           &pubKeyAny,
				Moniker:          args[2],
				Power:            power,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
func CmdSetValidatorPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-validator-power [validator-address] [power]",
		Short: "Set a validator's voting power",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid validator address: %w", err)
			}

			power, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid power: %w", err)
			}

			msg := &types.MsgSetValidatorPower{
				Admin:            clientCtx.GetFromAddress().String(),
				ValidatorAddress: valAddr.String(),
				Power:            power,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRemoveValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-validator [validator-address]",
		Short: "Remove a validator from the PoA network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid validator address: %w", err)
			}

			msg := &types.MsgRemoveValidator{
				Admin:            clientCtx.GetFromAddress().String(),
				ValidatorAddress: valAddr.String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateAuthorityAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-authority-accounts [account1,account2,...]",
		Short: "Update the list of authority accounts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			accounts := strings.Split(args[0], ",")
			for i, acc := range accounts {
				accounts[i] = strings.TrimSpace(acc)
				if _, err := sdk.AccAddressFromBech32(accounts[i]); err != nil {
					return fmt.Errorf("invalid account address %s: %w", accounts[i], err)
				}
			}

			msg := &types.MsgUpdateAuthorityAccounts{
				Admin:                clientCtx.GetFromAddress().String(),
				NewAuthorityAccounts: accounts,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
