package types

import "github.com/cosmos/cosmos-sdk/codec"

// ModuleCdc module level codec
var ModuleCdc = codec.New()

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
}
