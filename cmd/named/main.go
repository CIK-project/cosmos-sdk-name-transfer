package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CIK-project/cosmos-sdk-name-transfer/app"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cfg "github.com/tendermint/tendermint/config"
)

const (
	flagOverwrite = "overwrite"
	flagMoniker   = "moniker"
)

type printInfo struct {
	Moniker    string          `json:"moniker"`
	ChainID    string          `json:"chain_id"`
	NodeID     string          `json:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir"`
	Secret     string          `json:"secret`
	AppMessage json.RawMessage `json:"app_message"`
}

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("name", "namepub")
	config.SetBech32PrefixForValidator("nameval", "namevalpub")
	config.SetBech32PrefixForConsensusNode("namecons", "nameconspub")
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "named",
		Short:             "Name Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	rootCmd.AddCommand(initCmd(ctx, cdc))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func initCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			// make block intervals short
			config.Consensus.TimeoutPropose = 600 * time.Millisecond
			config.Consensus.TimeoutPrevote = 200 * time.Millisecond
			config.Consensus.TimeoutPrecommit = 200 * time.Millisecond
			config.Consensus.TimeoutCommit = 1 * time.Second

			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}

			nodeID, pk, err := gaiaInit.InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			config.Moniker = viper.GetString(flagMoniker)

			var appState json.RawMessage
			genFile := config.GenesisFile()
			var secret string

			if appState, secret, err = initializeGenesis(cdc, genFile, chainID, viper.GetBool(flagOverwrite)); err != nil {
				return err
			}

			if err = gaiaInit.ExportGenesisFile(genFile, chainID, []tmtypes.GenesisValidator{
				tmtypes.GenesisValidator{
					Address: pk.Address(),
					PubKey:  pk,
					Power:   10,
					Name:    "root",
				},
			}, appState); err != nil {
				return err
			}

			toPrint := printInfo{
				ChainID:    chainID,
				Moniker:    config.Moniker,
				NodeID:     nodeID,
				Secret:     secret,
				AppMessage: appState,
			}

			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(flagMoniker, "", "set the validator's moniker")
	cmd.MarkFlagRequired(flagMoniker)

	return cmd
}

func initializeGenesis(
	cdc *codec.Codec, genFile, chainID string, overwrite bool,
) (appState json.RawMessage, secret string, err error) {

	if !overwrite && common.FileExists(genFile) {
		return nil, "", fmt.Errorf("genesis.json file already exists: %v", genFile)
	}

	addr, secret, err := server.GenerateSaveCoinKey(app.DefaultCLIHome, "root", "12345678", false)
	if err != nil {
		return nil, "", err
	}

	appState, err = codec.MarshalJSONIndent(cdc, app.NewDefaultGenesisState(addr))
	return
}

func displayInfo(cdc *codec.Codec, info printInfo) error {
	out, err := codec.MarshalJSONIndent(cdc, info)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s\n", string(out))
	return nil
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewNameApp(logger, db, traceStore,
		baseapp.SetPruning(viper.GetString("pruning")),
		baseapp.SetMinimumFees(viper.GetString("minimum_fees")),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	return nil, nil, errors.New("Not yet implemented")
}
