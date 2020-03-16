package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/kava-labs/kava/x/incentive/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

// PayoutClaim sends the timelocked claim coins to the input address
func (k Keeper) PayoutClaim(ctx sdk.Context, addr sdk.AccAddress, denom string, id uint64) sdk.Error {
	claim, found := k.GetClaim(ctx, addr, denom, id)
	if !found {
		return sdk.ErrInternal("msg string") // TODO ErrClaimNotFound
	}
	claimPeriod, found := k.GetClaimPeriod(ctx, id, denom)
	if !found {
		return sdk.ErrInternal("msg string") // TODO ErrClaimPeriodNotFound
	}
	err := k.SendCoinsFromModuleToVestingAccount(ctx, types.IncentiveMacc, addr, sdk.NewCoins(claim.Reward), int64(claimPeriod.TimeLock.Seconds()))
	if err != nil {
		return err
	}
	return nil

}

// SendCoinsFromModuleToVestingAccount sends time-locked coins from the input module account to the recipient. If the recipients account is not a vesting account, it is converted to a periodic vesting account and the coins are added to the vesting balance as a vesting period with the input length.
func (k Keeper) SendCoinsFromModuleToVestingAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins, length int64) sdk.Error {
	// TODO

	// 0. Get the account from the account keeper and do a type switch, throw if it's a validator vesting account (can TODO and make this work later if necessary)
	acc := k.accountKeeper.GetAccount(ctx, recipientAddr)

	_, ok := acc.(validatorvesting.ValidatorVestingAccount)
	if ok {
		return sdk.ErrInternal("msg string") // TODO ErrInvalidAccountType
	}
	// 1. Transfer coins using regular supply keeper module account to account method. This will update the Coins field on the account
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
	if err != nil {
		return err
	}
	// 2. Get the account from the account keeper and do a type switch
	acc = k.accountKeeper.GetAccount(ctx, recipientAddr)
	vacc, ok := acc.(vesting.PeriodicVestingAccount)
	if ok {
		// 2a. If it's a periodic vesting account, update the account:
		proposedEndTime := ctx.BlockTime().Unix() + length
		// 2a2. Update the original vesting coins. TODO Do I need to remove the coins from the 'Coins' field?
		vacc.OriginalVesting = vacc.OriginalVesting.Add(amt)
		// 2a3. Update the periods
		totalPeriodLength := types.GetTotalVestingPeriodLength(vacc.VestingPeriods)
		// in the case that the proposed length is longer than the sum of all previous period lengths, create a new period with length equal to the difference between the proposed length and the previous total length
		if totalPeriodLength < length {
			newPeriodLength := length - totalPeriodLength
			newPeriod := vesting.Period{Amount: amt, Length: newPeriodLength}
			vacc.VestingPeriods = append(vacc.VestingPeriods, newPeriod)
			// need to update the end time as well so that the sum of all period lengths equals endTime - startTime
			vacc.EndTime = proposedEndTime
		} else {
			// In the case that the proposed length is lessthan or equal to the sum of all previous period lengths, insert the period and update other periods as necessary.
			// EXAMPLE (l is length, a is amount)
			// Original Periods: {[l: 1 a: 1], [l: 2, a: 1], [l:8, a:3], [l: 5, a: 3]}
			// Period we want to insert [l: 5, a: x]
			// Expected result:
			// {[l: 1, a: 1], [l:2, a: 1], [l:2, a:x], [l:6, a:3], [l:5, a:3]}
			newPeriods := vesting.Periods{}
			lengthCounter := int64(0)
			appendRemaining := false
			for _, period := range vacc.VestingPeriods {
				if appendRemaining {
					newPeriods = append(newPeriods, period)
					continue
				}
				lengthCounter += period.Length
				if lengthCounter < length {
					newPeriods = append(newPeriods, period)
				} else if lengthCounter == length {
					newPeriod := vesting.Period{Length: period.Length, Amount: period.Amount.Add(amt)}
					newPeriods = append(newPeriods, newPeriod)
					appendRemaining = true
				} else {
					newPeriod := vesting.Period{
						Length: length - types.GetTotalVestingPeriodLength(newPeriods),
						Amount: amt,
					}
					previousPeriod := vesting.Period{
						Length: period.Length - newPeriod.Length,
						Amount: period.Amount,
					}
					newPeriods = append(newPeriods, newPeriod, previousPeriod)
					appendRemaining = true
				}
			}
		}
	} else {
		// 3b. If it's not a periodic vesting account (TODO make sure that we will have already thrown if the input account is a module account), transition the account to a periodic vesting account:
		bacc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
		newPeriods := vesting.Periods{
			vesting.Period{
				Length: length,
				Amount: amt,
			},
		}
		bva, err := vesting.NewBaseVestingAccount(bacc, amt, ctx.BlockTime().Unix()+length)
		if err != nil {
			return sdk.ErrInternal(err.Error()) // TODO
		}
		pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), newPeriods)
		k.accountKeeper.SetAccount(ctx, pva) // TODO does this actually replace the account (ie it saves to the same key)?

	}

	return nil
}