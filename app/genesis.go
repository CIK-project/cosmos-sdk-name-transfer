package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// State to Unmarshal
type GenesisState struct {
	Accounts []GenesisAccount `json:"accounts"`
}

func NewGenesisState(accounts []GenesisAccount) GenesisState {

	return GenesisState{
		Accounts: accounts,
	}
}

// nolint
type GenesisAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	Sequence      uint64         `json:"sequence_number"`
	AccountNumber uint64         `json:"account_number"`
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	return GenesisAccount{
		Address:       acc.GetAddress(),
		Coins:         acc.GetCoins(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address:       ga.Address,
		Coins:         ga.Coins.Sort(),
		AccountNumber: ga.AccountNumber,
		Sequence:      ga.Sequence,
	}
}

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState(addr sdk.AccAddress) GenesisState {
	return NewGenesisState([]GenesisAccount{NewDefaultGenesisAccount(addr)})
}

func NewDefaultGenesisAccount(addr sdk.AccAddress) GenesisAccount {
	accAuth := auth.NewBaseAccountWithAddress(addr)
	coins := sdk.Coins{
		sdk.NewCoin("name", sdk.NewInt(1000)),
	}

	coins.Sort()

	accAuth.Coins = coins
	return NewGenesisAccount(&accAuth)
}
