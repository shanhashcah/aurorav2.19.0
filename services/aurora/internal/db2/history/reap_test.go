package history_test

import (
	"testing"

	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/services/aurora/internal/ledger"
	"github.com/hcnet/go/services/aurora/internal/reap"
	"github.com/hcnet/go/services/aurora/internal/test"
)

func TestReapLookupTables(t *testing.T) {
	tt := test.Start(t)
	defer tt.Finish()
	ledgerState := &ledger.State{}
	ledgerState.SetStatus(tt.Scenario("kahuna"))

	db := tt.AuroraSession()

	sys := reap.New(0, db, ledgerState)

	var (
		prevLedgers, curLedgers                     int
		prevClaimableBalances, curClaimableBalances int
		prevLiquidityPools, curLiquidityPools       int
	)

	// Prev
	{
		err := db.GetRaw(tt.Ctx, &prevLedgers, `SELECT COUNT(*) FROM history_ledgers`)
		tt.Require.NoError(err)
		err = db.GetRaw(tt.Ctx, &prevClaimableBalances, `SELECT COUNT(*) FROM history_claimable_balances`)
		tt.Require.NoError(err)
		err = db.GetRaw(tt.Ctx, &prevLiquidityPools, `SELECT COUNT(*) FROM history_liquidity_pools`)
		tt.Require.NoError(err)
	}

	ledgerState.SetStatus(tt.LoadLedgerStatus())
	sys.RetentionCount = 1
	err := sys.DeleteUnretainedHistory(tt.Ctx)
	tt.Require.NoError(err)

	q := &history.Q{tt.AuroraSession()}

	err = q.Begin()
	tt.Require.NoError(err)

	newOffsets, err := q.ReapLookupTables(tt.Ctx, nil)
	tt.Require.NoError(err)

	err = q.Commit()
	tt.Require.NoError(err)

	// cur
	{
		err := db.GetRaw(tt.Ctx, &curLedgers, `SELECT COUNT(*) FROM history_ledgers`)
		tt.Require.NoError(err)
		err = db.GetRaw(tt.Ctx, &curClaimableBalances, `SELECT COUNT(*) FROM history_claimable_balances`)
		tt.Require.NoError(err)
		err = db.GetRaw(tt.Ctx, &curLiquidityPools, `SELECT COUNT(*) FROM history_liquidity_pools`)
		tt.Require.NoError(err)
	}

	tt.Assert.Equal(61, prevLedgers, "prevLedgers")
	tt.Assert.Equal(1, curLedgers, "curLedgers")
	tt.Assert.Equal(1, prevClaimableBalances, "prevClaimableBalances")
	tt.Assert.Equal(0, curClaimableBalances, "curClaimableBalances")
	tt.Assert.Equal(1, prevLiquidityPools, "prevLiquidityPools")
	tt.Assert.Equal(0, curLiquidityPools, "curLiquidityPools")

	tt.Assert.Len(newOffsets, 2)
	tt.Assert.Equal(int64(0), newOffsets["history_claimable_balances"])
	tt.Assert.Equal(int64(0), newOffsets["history_liquidity_pools"])
}
