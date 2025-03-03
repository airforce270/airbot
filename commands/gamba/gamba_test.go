package gamba_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/commands/commandtest"
	"github.com/airforce270/airbot/commands/gamba"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/permission"
	"github.com/airforce270/airbot/utils/ptrs"
)

func TestGambaCommands(t *testing.T) {
	t.Parallel()
	tests := []commandtest.Case{
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "user1 won the duel with user2 and wins 25 points!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "user2 won the duel with user1 and wins 25 points!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$accept",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "There are no duels pending against you.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$decline",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Declined duel.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$decline",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "There are no duels pending against you.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "@user2, user1 has started a duel for 25 points! Type $accept or $decline in the next 30 seconds!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser2,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You don't have enough points for that duel (you have 0 points)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "user2 don't have enough points for that duel (they have 0 points)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You already have a duel pending.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 25",
					UserID:  "user3",
					User:    "user3",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
				add50PointsToUser2,
				add50PointsToUser3,
				startDuel,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "That chatter already has a duel pending.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user1 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You can't duel yourself Pepega",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 0",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You must duel at least 1 point.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$duel user2 xx",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $duel <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "user1 gave 10 points to user2 FeelsOkayMan <3",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 100",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You can't give more points than you have (you have 50 points)",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 0",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "You must give at least 1 point.",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2 xx",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints user2",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$givepoints",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "Usage: $givepoints <user> <amount>",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$points user1",
				"$p",
				"$p user1",
			},
			RunBefore: []commandtest.SetupFunc{
				add50PointsToUser1,
			},
			Want: []*base.Message{
				{
					Text:    "GAMBA user1 has 50 points",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points user1",
					UserID:  "user2",
					User:    "user2",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$p user1"},
			RunBefore: []commandtest.SetupFunc{
				add50PointsToUser1,
			},
			Want: []*base.Message{
				{
					Text:    "GAMBA user1 has 50 points",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$points rando",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$p rando"},
			Want: []*base.Message{
				{
					Text:    "rando has never been seen by fake-username",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$r 10",
				"$roulette 20%",
				"$r 20%",
			},
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "GAMBA user1 won 10 points in roulette and now has 60 points!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 10",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform: commandtest.TwitchPlatform,
			OtherTexts: []string{
				"$r 10",
				"$roulette 20%",
				"$r 20%",
			},
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "GAMBA user1 lost 10 points in roulette and now has 40 points!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette all",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$r all"},
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo1,
				add50PointsToUser1,
			},
			RunAfter: []commandtest.TeardownFunc{
				waitForTransactionsToSettle,
			},
			Want: []*base.Message{
				{
					Text:    "GAMBA user1 won 50 points in roulette and now has 100 points!",
					Channel: "user2",
				},
			},
		},
		{
			Input: base.IncomingMessage{
				Message: base.Message{
					Text:    "$roulette 60",
					UserID:  "user1",
					User:    "user1",
					Channel: "user2",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix:          "$",
				PermissionLevel: permission.Normal,
			},
			Platform:   commandtest.TwitchPlatform,
			OtherTexts: []string{"$r 60"},
			RunBefore: []commandtest.SetupFunc{
				deleteAllGambaTransactions,
				setRandValueTo0,
				add50PointsToUser1,
			},
			Want: []*base.Message{
				{
					Text:    "user1: You don't have enough points for that (current: 50)",
					Channel: "user2",
				},
			},
		},
	}

	commandtest.Run(t, tests)
}

func TestFetchUserPoints(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	db, queries := databasetest.New(t)
	user1, err := database.SelectOrCreateTwitchUser(t.Context(), db, queries, "user1" /* id */, "user1" /* name */)
	if err != nil {
		t.Fatalf("failed to find/create user1: %v", err)
	}
	user2, err := database.SelectOrCreateTwitchUser(t.Context(), db, queries, "user2" /* id */, "user2" /* name */)
	if err != nil {
		t.Fatalf("failed to find/create user2: %v", err)
	}

	tests := []struct {
		desc         string
		transactions []database.CreateGambaTransactionParams
		want         int64
	}{
		{
			desc:         "no transactions",
			transactions: []database.CreateGambaTransactionParams{},
			want:         0,
		},
		{
			desc: "single transaction for single user",
			transactions: []database.CreateGambaTransactionParams{
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: ptrs.Ptr[int64](1),
					Delta:  ptrs.Ptr[int64](50),
				},
			},
			want: 50,
		},
		{
			desc: "multiple transactions for single user",
			transactions: []database.CreateGambaTransactionParams{
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user1.ID,
					Delta:  ptrs.Ptr[int64](50),
				},
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user1.ID,
					Delta:  ptrs.Ptr[int64](-20),
				},
			},
			want: 30,
		},
		{
			desc: "multiple transactions +/- for single user",
			transactions: []database.CreateGambaTransactionParams{
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user1.ID),
					Delta:  ptrs.Int64Nil(50),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user1.ID),
					Delta:  ptrs.Int64Nil(-20),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user1.ID),
					Delta:  ptrs.Int64Nil(5),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user1.ID),
					Delta:  ptrs.Int64Nil(100),
				},
			},
			want: 135,
		},
		{
			desc: "single transaction for other user",
			transactions: []database.CreateGambaTransactionParams{
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user2.ID),
					Delta:  ptrs.Int64Nil(50),
				},
			},
			want: 0,
		},
		{
			desc: "multiple transactions +/- for multiple users",
			transactions: []database.CreateGambaTransactionParams{
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user1.ID,
					Delta:  ptrs.Ptr[int64](50),
				},
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user1.ID,
					Delta:  ptrs.Ptr[int64](-20),
				},
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: ptrs.Int64Nil(user1.ID),
					Delta:  ptrs.Ptr[int64](5),
				},
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user1.ID,
					Delta:  ptrs.Ptr[int64](100),
				},
				{
					Game:   ptrs.Ptr("FAKE - TEST"),
					UserID: &user2.ID,
					Delta:  ptrs.Ptr[int64](50),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: &user2.ID,
					Delta:  ptrs.Ptr[int64](-20),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: &user2.ID,
					Delta:  ptrs.Ptr[int64](5),
				},
				{
					Game:   ptrs.StringNil("FAKE - TEST"),
					UserID: &user2.ID,
					Delta:  ptrs.Ptr[int64](100),
				},
			},
			want: 135,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if err := queries.DeleteAllGambaTransactionsForTest(ctx); err != nil {
				t.Fatalf("failed to drop all gamba transactions: %v", err)
			}

			if _, err := database.CreateGambaTransactions(ctx, db, queries, tc.transactions); err != nil {
				t.Fatalf("failed to insert gamba transactions: %v", err)
			}

			got, err := gamba.FetchUserPoints(ctx, queries, user1)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("fetchUserPoints() = %d, want %d", got, tc.want)
			}
		})
	}
}

func setRandValueTo0(t testing.TB, r *base.Resources) {
	r.Rand.Reader = bytes.NewBuffer([]byte{0})
}

func setRandValueTo1(t testing.TB, r *base.Resources) {
	r.Rand.Reader = bytes.NewBuffer([]byte{1})
}

func waitForTransactionsToSettle(t testing.TB) {
	time.Sleep(20 * time.Millisecond)
}

func deleteAllGambaTransactions(t testing.TB, r *base.Resources) {
	t.Helper()
	if err := r.Queries.DeleteAllGambaTransactionsForTest(t.Context()); err != nil {
		t.Fatalf("Failed to delete all gamba txns: %v", err)
	}
}

func startDuel(t testing.TB, r *base.Resources) {
	t.Helper()

	user1, err := r.Queries.SelectTwitchUser(t.Context(), database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	user2, err := r.Queries.SelectTwitchUser(t.Context(), database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("Failed to find user2: %v", err)
	}
	_, err = r.Queries.CreateDuel(t.Context(), database.CreateDuelParams{
		UserID:   &user1.ID,
		TargetID: &user2.ID,
		Amount:   ptrs.Ptr[int64](25),
		Pending:  ptrs.TrueFloat,
		Accepted: ptrs.FalseFloat,
	})
	if err != nil {
		t.Fatalf("Failed to create duel: %v", err)
	}
}

func add50PointsToUser1(t testing.TB, r *base.Resources) {
	t.Helper()
	user, err := database.SelectOrCreateTwitchUser(t.Context(), r.DB, r.Queries, "user1" /* id */, "user1" /* name */)
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	add50PointsToUser(t, user, r.Queries)
}

func add50PointsToUser2(t testing.TB, r *base.Resources) {
	t.Helper()
	user, err := database.SelectOrCreateTwitchUser(t.Context(), r.DB, r.Queries, "user2" /* id */, "user2" /* name */)
	if err != nil {
		t.Fatalf("Failed to find/create user2: %v", err)
	}
	add50PointsToUser(t, user, r.Queries)
}

func add50PointsToUser3(t testing.TB, r *base.Resources) {
	t.Helper()
	user, err := database.SelectOrCreateTwitchUser(t.Context(), r.DB, r.Queries, "user3" /* id */, "user3" /* name */)
	if err != nil {
		t.Fatalf("Failed to find/create user3: %v", err)
	}
	add50PointsToUser(t, user, r.Queries)
}

func add50PointsToUser(t testing.TB, user database.User, queries *database.Queries) {
	t.Helper()
	txn := database.CreateGambaTransactionParams{
		Game:   ptrs.StringNil("FAKE - TEST"),
		UserID: ptrs.Int64Nil(user.ID),
		Delta:  ptrs.Int64Nil(50),
	}
	if _, err := queries.CreateGambaTransaction(t.Context(), txn); err != nil {
		t.Fatalf("Failed to insert gamba transaction: %v", err)
	}
}
