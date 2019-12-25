package types

import (
	"bytes"
	"fmt"
	"time"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params            Params    `json:"params" yaml:"params"`
	CDPs              CDPs      `json:"cdps" yaml:"cdps"`
	StartingCdpID     uint64    `json:"starting_cdp_id" yaml:"starting_cdp_id"`
	DebtDenom         string    `json:"debt_denom" yaml:"debt_denom"`
	PreviousBlockTime time.Time `json:"previous_block_time" yaml:"previous_block_time"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:            DefaultParams(),
		CDPs:              CDPs{},
		StartingCdpID:     DefaultCdpStartingID,
		DebtDenom:         DefaultDebtDenom,
		PreviousBlockTime: DefaultPreviousBlockTime,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {

	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.PreviousBlockTime.Equal(time.Time{}) {
		return fmt.Errorf("previous block time not set")
	}

	if gs.DebtDenom == "" {
		return fmt.Errorf("debt denom not set")

	}

	return nil
}

// Equal checks whether two gov GenesisState structs are equivalent
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(data)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (data GenesisState) IsEmpty() bool {
	return data.Equal(GenesisState{})
}
