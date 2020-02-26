package keeper_test

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/types"
)

type AuctionType int

const (
	Invalid    AuctionType = 0
	Surplus    AuctionType = 1
	Debt       AuctionType = 2
	Collateral AuctionType = 3
)

func TestAuctionBidding(t *testing.T) {
	someTime := time.Date(0001, time.January, 1, 0, 0, 0, 0, time.UTC)

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	buyer := addrs[0]
	secondBuyer := addrs[1]
	modName := "liquidator"
	collateralAddrs := addrs[2:]
	collateralWeights := is(30, 20, 10)

	type auctionArgs struct {
		auctionType AuctionType
		seller      string
		lot         sdk.Coin
		bid         sdk.Coin
		debt        sdk.Coin
		addresses   []sdk.AccAddress
		weights     []sdk.Int
	}

	type bidArgs struct {
		bidder sdk.AccAddress
		amount sdk.Coin
	}

	tests := []struct {
		name            string
		auctionArgs     auctionArgs
		setupBids       []bidArgs
		bidArgs         bidArgs
		expectedError   sdk.CodeType
		expectedEndTime time.Time
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectpass      bool
	}{
		{
			"basic: auction doesn't exist",
			auctionArgs{Surplus, "", c("token1", 1), c("token2", 1), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			types.CodeAuctionNotFound,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"surplus: normal",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"surplus: second bidder",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{secondBuyer, c("token2", 11)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
		},
		{
			"surplus: invalid bid denom",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.CodeInvalidBidDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"surplus: invalid bid (equal)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 0)},
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"surplus: invalid bid (less than)",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 0), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			[]bidArgs{{buyer, c("token2", 100)}},
			bidArgs{buyer, c("token2", 99)},
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"debt: normal",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			nil,
			bidArgs{buyer, c("token1", 10)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 100),
			true,
		},
		{
			"debt: second bidder",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			[]bidArgs{{buyer, c("token1", 10)}},
			bidArgs{secondBuyer, c("token1", 9)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 100),
			true,
		},
		{
			"debt: invalid lot denom",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.CodeInvalidLotDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"debt: invalid lot size (larger)",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token1", 21)},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"debt: invalid lot size (equal)",
			auctionArgs{Debt, modName, c("token1", 20), c("token2", 100), c("debt", 20), []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token1", 20)},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"collateral [forward]: normal",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 10)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"collateral [forward]: second bidder",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{secondBuyer, c("token2", 11)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 11),
			true,
		},
		{
			"collateral [forward]: invalid bid denom",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("badtoken", 10)},
			types.CodeInvalidBidDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [forward]: invalid bid size (smaller)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 10)}},
			bidArgs{buyer, c("token2", 9)},
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [forward]: invalid bid size (equal)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 0)},
			types.CodeBidTooSmall,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [forward]: invalid bid size (greater than max)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 100), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			nil,
			bidArgs{buyer, c("token2", 101)},
			types.CodeBidTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"collateral [reverse]: normal",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}}, // put auction into reverse phase
			bidArgs{buyer, c("token1", 15)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			true,
		},
		{
			"collateral [reverse]: second bidder",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}, {buyer, c("token1", 15)}},                                                         // put auction into reverse phase, and add a reverse phase bid
			bidArgs{secondBuyer, c("token1", 14)},
			sdk.CodeType(0),
			someTime.Add(types.DefaultBidDuration),
			secondBuyer,
			c("token2", 50),
			true,
		},
		{
			"collateral [reverse]: invalid lot denom",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}}, // put auction into reverse phase
			bidArgs{buyer, c("badtoken", 15)},
			types.CodeInvalidLotDenom,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"collateral [reverse]: invalid lot size (equal)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}},                                                                                   // put auction into reverse phase
			bidArgs{buyer, c("token1", 20)},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"collateral [reverse]: invalid lot size (greater)",
			auctionArgs{Collateral, modName, c("token1", 20), c("token2", 50), c("debt", 50), collateralAddrs, collateralWeights}, // lot, max bid
			[]bidArgs{{buyer, c("token2", 50)}},                                                                                   // put auction into reverse phase
			bidArgs{buyer, c("token1", 21)},
			types.CodeLotTooLarge,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		{
			"basic: closed auction",
			auctionArgs{Surplus, modName, c("token1", 100), c("token2", 10), sdk.Coin{}, []sdk.AccAddress{}, []sdk.Int{}},
			nil,
			bidArgs{buyer, c("token2", 10)},
			types.CodeAuctionHasExpired,
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test
			tApp := app.NewTestApp()
			// Set up seller account
			sellerAcc := supply.NewEmptyModuleAccount(modName, supply.Minter, supply.Burner)
			require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 1000), c("token2", 1000), c("debt", 1000))))
			// Initialize genesis accounts
			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{
					auth.NewBaseAccount(buyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(secondBuyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[0], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[1], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					auth.NewBaseAccount(collateralAddrs[2], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
					sellerAcc,
				}),
			)
			ctx := tApp.NewContext(false, abci.Header{})
			keeper := tApp.GetAuctionKeeper()

			// Start Auction
			var id uint64
			var err sdk.Error
			switch tc.auctionArgs.auctionType {
			case Surplus:
				id, _ = keeper.StartSurplusAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid.Denom)
			case Debt:
				id, _ = keeper.StartDebtAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.bid, tc.auctionArgs.lot, tc.auctionArgs.debt)
			case Collateral:
				id, _ = keeper.StartCollateralAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid, tc.auctionArgs.addresses, tc.auctionArgs.weights, tc.auctionArgs.debt) // seller, lot, maxBid, otherPerson
			default:
				t.Fail()
			}
			// Place setup bids
			for _, b := range tc.setupBids {
				require.NoError(t, keeper.PlaceBid(ctx, id, b.bidder, b.amount))
			}

			// Close the auction early to test late bidding (if applicable)
			if strings.Contains(tc.name, "closed") {
				ctx = ctx.WithBlockTime(types.DistantFuture.Add(1))
			}

			// Place bid on auction
			err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.bidArgs.amount)

			// Check success/failure
			if tc.expectpass {
				require.Nil(t, err)
				// Get auction from store
				auction, found := keeper.GetAuction(ctx, id)
				require.True(t, found)
				// Check auction values
				require.Equal(t, modName, auction.GetInitiator())
				require.Equal(t, tc.expectedBidder, auction.GetBidder())
				require.Equal(t, tc.expectedBid, auction.GetBid())
				require.Equal(t, tc.expectedEndTime, auction.GetEndTime())
			} else {
				// Check expected error code type
				require.NotNil(t, err) // catch nil values before they cause a panic below
				require.Equal(t, tc.expectedError, err.Result().Code)
			}
		})
	}
}