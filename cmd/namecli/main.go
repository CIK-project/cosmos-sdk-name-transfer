package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/CIK-project/cosmos-sdk-name-transfer/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

const (
	storeAcc = "acc"
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Instantiate the codec for the command line application
	cdc := app.MakeCodec()

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("name", "namepub")
	config.SetBech32PrefixForValidator("nameval", "namevalpub")
	config.SetBech32PrefixForConsensusNode("namecons", "nameconspub")
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "gaiacli",
		Short: "Command line interface for interacting with gaiad",
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.InitClientCommand(),
		rpc.StatusCommand(),
		client.ConfigCmd(),
		queryCmd(cdc),
		txCmd(cdc),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// Add flags and prefix all env exposed with GA
	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := initConfig(rootCmd)
	if err != nil {
		panic(err)
	}

	err = executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(storeAcc, cdc),
	)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		bankcmd.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
