package gamba

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/airforce270/airbot/apiclients/twitchtest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/utils/ptrs"

	"github.com/google/go-cmp/cmp"
)

func TestHasOutboundPendingDuels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc      string
		runBefore []func(testing.TB, *database.Queries)
		want      int
	}{
		{
			desc:      "no outbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has outbound pending duels",
			runBefore: []func(testing.TB, *database.Queries){
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			_, queries := databasetest.New(t)
			if err := queries.DeleteAllDuelsForTest(ctx); err != nil {
				t.Fatal(err)
			}

			user1, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
				TwitchID:   ptrs.Ptr("user1"),
				TwitchName: ptrs.Ptr("user1"),
			})
			if err != nil {
				t.Fatalf("failed to find user1: %v", err)
			}

			for _, f := range tc.runBefore {
				f(t, queries)
			}

			got, err := OutboundPendingDuels(ctx, user1, 30*time.Second, queries)
			if err != nil {
				t.Fatalf("OutboundPendingDuels() unexpected err: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("OutboundPendingDuels() len = %d, want %d", len(got), tc.want)
			}
		})
	}
}

func TestInboundPendingDuels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc      string
		runBefore []func(testing.TB, *database.Queries)
		want      int
	}{
		{
			desc:      "no inbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has inbound pending duels",
			runBefore: []func(testing.TB, *database.Queries){
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			_, queries := databasetest.New(t)
			if err := queries.DeleteAllDuelsForTest(ctx); err != nil {
				t.Fatal(err)
			}

			user2, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
				TwitchID:   ptrs.Ptr("user2"),
				TwitchName: ptrs.Ptr("user2"),
			})
			if err != nil {
				t.Fatalf("failed to find user2: %v", err)
			}

			for _, f := range tc.runBefore {
				f(t, queries)
			}

			got, err := InboundPendingDuels(ctx, user2, 30*time.Second, queries)
			if err != nil {
				t.Fatalf("InboundPendingDuels() unexpected err: %v", err)
			}
			if len(got) != tc.want {
				t.Errorf("InboundPendingDuels() len = %d, want %d", len(got), tc.want)
			}
		})
	}
}

func TestGrantPoints(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	db, queries := databasetest.New(t)
	server := newTestServer()

	if err := queries.DeleteAllUsersForTest(ctx); err != nil {
		t.Fatal(err)
	}

	user1, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user2"),
		TwitchName: ptrs.Ptr("user2"),
	})
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	messages := []database.CreateMessageParams{
		{
			UserID:  &user1.ID,
			Channel: ptrs.Ptr("channel1"),
			Text:    ptrs.Ptr("something"),
			Time:    ptrs.Ptr(time.Now().Add(-1 * time.Minute)),
		},
		{
			UserID:  &user2.ID,
			Channel: ptrs.Ptr("channel1"),
			Text:    ptrs.Ptr("something else"),
			Time:    ptrs.Ptr(time.Now().Add(-50 * time.Minute)),
		},
	}
	for i, m := range messages {
		if _, err := queries.CreateMessage(ctx, m); err != nil {
			t.Fatalf("failed to create message %d: %v", i, err)
		}
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTestingWithDB(t, server.URL, db, queries),
	}

	grantPoints(ctx, ps, queries)

	transactions, err := queries.SelectAllGambaTransactions(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 gamba transactions, found %d: %v", len(transactions), transactions)
	}

	if *transactions[0].Delta != activeGrantAmount {
		t.Errorf("transaction 0 should have granted %d points, but granted %d", activeGrantAmount, transactions[0].Delta)
	}
	if *transactions[0].UserID != user1.ID {
		t.Errorf("transaction 0 should have been for user 1 but was for user %d", transactions[0].UserID)
	}
	if *transactions[1].Delta != inactiveGrantAmount {
		t.Errorf("transaction 1 should have granted %d points, but granted %d", inactiveGrantAmount, transactions[1].Delta)
	}
	if *transactions[1].UserID != user2.ID {
		t.Errorf("transaction 1 should have been for user 2 but was for user %d", transactions[1].UserID)
	}
}

func TestGetInactiveUsers(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	db, queries := databasetest.New(t)
	server := newTestServer()

	if err := queries.DeleteAllUsersForTest(ctx); err != nil {
		t.Fatal(err)
	}

	user1, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user2"),
		TwitchName: ptrs.Ptr("user2"),
	})
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTestingWithDB(t, server.URL, db, queries),
	}

	got := getInactiveUsers(ctx, ps)
	want := []database.User{user1, user2}

	if len(got) != len(want) {
		t.Fatalf("getInactiveUsers() got %d users, want %d: diff (-want +got):\n%s", len(got), len(want), cmp.Diff(want, got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("getInactiveUsers()[0].ID = %d want %d", got[0].ID, want[0].ID)
	}
}

func TestGetActiveUsers(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	_, queries := databasetest.New(t)

	user1, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2, err := queries.CreateTwitchUser(ctx, database.CreateTwitchUserParams{
		TwitchID:   ptrs.Ptr("user2"),
		TwitchName: ptrs.Ptr("user2"),
	})
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	messages := []database.CreateMessageParams{
		{
			UserID:  &user1.ID,
			Channel: ptrs.Ptr("channel1"),
			Text:    ptrs.Ptr("something"),
			Time:    ptrs.Ptr(time.Now().Add(-1 * time.Minute)),
		},
		{
			UserID:  &user2.ID,
			Channel: ptrs.Ptr("channel1"),
			Text:    ptrs.Ptr("something else"),
			Time:    ptrs.Ptr(time.Now().Add(-50 * time.Minute)),
		},
	}
	for i, m := range messages {
		if _, err := queries.CreateMessage(ctx, m); err != nil {
			t.Fatalf("failed to create message %d: %v", i, err)
		}
	}

	got, err := getActiveUsers(ctx, queries)
	if err != nil {
		t.Fatalf("getActiveUsers() unexpected error: %v", err)
	}
	want := []database.User{user1}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("getActiveUsers() diff (-want +got):\n%s", diff)
	}
}

func TestDeduplicateByUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc  string
		input []grant
		want  []grant
	}{
		{
			desc: "all active",
			input: []grant{
				{
					User:     database.User{ID: 1},
					IsActive: true,
				},
				{
					User:     database.User{ID: 2},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User:     database.User{ID: 1},
					IsActive: true,
				},
				{
					User:     database.User{ID: 2},
					IsActive: true,
				},
			},
		},
		{
			desc: "all inactive",
			input: []grant{
				{
					User:     database.User{ID: 1},
					IsActive: false,
				},
				{
					User:     database.User{ID: 2},
					IsActive: false,
				},
			},
			want: []grant{
				{
					User:     database.User{ID: 1},
					IsActive: false,
				},
				{
					User:     database.User{ID: 2},
					IsActive: false,
				},
			},
		},
		{
			desc: "inactive and active unsorted",
			input: []grant{
				{
					User:     database.User{ID: 1},
					IsActive: false,
				},
				{
					User:     database.User{ID: 2},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User:     database.User{ID: 2},
					IsActive: true,
				},
				{
					User:     database.User{ID: 1},
					IsActive: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := deduplicateByUser(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("deduplicateByUser() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func startDuel(t testing.TB, queries *database.Queries) {
	t.Helper()
	ctx := t.Context()
	user1, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	user2, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user2"),
		TwitchName: ptrs.Ptr("user2"),
	})
	if err != nil {
		t.Fatalf("Failed to find user2: %v", err)
	}
	_, err = queries.CreateDuel(ctx, database.CreateDuelParams{
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

func add50PointsToUser1(t testing.TB, queries *database.Queries) {
	t.Helper()
	ctx := t.Context()
	user, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user1"),
		TwitchName: ptrs.Ptr("user1"),
	})
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	add50PointsToUser(t, user, queries)
}

func add50PointsToUser2(t testing.TB, queries *database.Queries) {
	t.Helper()
	ctx := t.Context()
	user, err := queries.SelectTwitchUser(ctx, database.SelectTwitchUserParams{
		TwitchID:   ptrs.Ptr("user2"),
		TwitchName: ptrs.Ptr("user2"),
	})
	if err != nil {
		t.Fatalf("Failed to find/create user2: %v", err)
	}
	add50PointsToUser(t, user, queries)
}

func add50PointsToUser(t testing.TB, user database.User, queries *database.Queries) {
	t.Helper()
	ctx := t.Context()
	_, err := queries.CreateGambaTransaction(ctx, database.CreateGambaTransactionParams{
		Game:   ptrs.Ptr("FAKE - TEST"),
		UserID: &user.ID,
		Delta:  ptrs.Ptr[int64](50),
	})
	if err != nil {
		t.Fatalf("Failed to insert gamba transaction: %v", err)
	}
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/chat/chatters") {
			fmt.Fprint(w, twitchtest.GetChannelChatChattersResp)
		} else if strings.Contains(r.URL.Path, "/users") {
			fmt.Fprint(w, twitchtest.GetUsersResp)
		} else {
			log.Printf("Unknown URL sent to test server: %s", r.URL.Path)
		}
	}))
}
