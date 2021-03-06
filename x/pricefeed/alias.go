// nolint
// autogenerated code using github.com/rigelrozanski/multitool
// aliases generated for the following subdirectories:
// ALIASGEN: github.com/kava-labs/kava/x/pricefeed/types/
// ALIASGEN: github.com/kava-labs/kava/x/pricefeed/keeper/
package pricefeed

import (
	"github.com/kava-labs/kava/x/pricefeed/keeper"
	"github.com/kava-labs/kava/x/pricefeed/types"
)

const (
	DefaultCodespace              = types.DefaultCodespace
	CodeEmptyInput                = types.CodeEmptyInput
	CodeExpired                   = types.CodeExpired
	CodeInvalidPrice              = types.CodeInvalidPrice
	CodeInvalidAsset              = types.CodeInvalidAsset
	CodeInvalidOracle             = types.CodeInvalidOracle
	EventTypeMarketPriceUpdated   = types.EventTypeMarketPriceUpdated
	EventTypeOracleUpdatedPrice   = types.EventTypeOracleUpdatedPrice
	EventTypeNoValidPrices        = types.EventTypeNoValidPrices
	AttributeValueCategory        = types.AttributeValueCategory
	AttributeMarketID             = types.AttributeMarketID
	AttributeMarketPrice          = types.AttributeMarketPrice
	AttributeOracle               = types.AttributeOracle
	AttributeExpiry               = types.AttributeExpiry
	AttributeKeyPriceUpdateFailed = types.AttributeKeyPriceUpdateFailed
	ModuleName                    = types.ModuleName
	StoreKey                      = types.StoreKey
	RouterKey                     = types.RouterKey
	QuerierRoute                  = types.QuerierRoute
	DefaultParamspace             = types.DefaultParamspace
	RawPriceFeedPrefix            = types.RawPriceFeedPrefix
	CurrentPricePrefix            = types.CurrentPricePrefix
	MarketPrefix                  = types.MarketPrefix
	OraclePrefix                  = types.OraclePrefix
	TypeMsgPostPrice              = types.TypeMsgPostPrice
	QueryPrice                    = types.QueryPrice
	QueryRawPrices                = types.QueryRawPrices
	QueryMarkets                  = types.QueryMarkets
)

var (
	// functions aliases
	RegisterCodec       = types.RegisterCodec
	ErrEmptyInput       = types.ErrEmptyInput
	ErrExpired          = types.ErrExpired
	ErrNoValidPrice     = types.ErrNoValidPrice
	ErrInvalidMarket    = types.ErrInvalidMarket
	ErrInvalidOracle    = types.ErrInvalidOracle
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	NewMsgPostPrice     = types.NewMsgPostPrice
	NewParams           = types.NewParams
	DefaultParams       = types.DefaultParams
	ParamKeyTable       = types.ParamKeyTable
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier

	// variable aliases
	ModuleCdc      = types.ModuleCdc
	KeyMarkets     = types.KeyMarkets
	DefaultMarkets = types.DefaultMarkets
)

type (
	GenesisState            = types.GenesisState
	Market                  = types.Market
	Markets                 = types.Markets
	CurrentPrice            = types.CurrentPrice
	PostedPrice             = types.PostedPrice
	SortDecs                = types.SortDecs
	MsgPostPrice            = types.MsgPostPrice
	Params                  = types.Params
	QueryWithMarketIDParams = types.QueryWithMarketIDParams
	Keeper                  = keeper.Keeper
)
