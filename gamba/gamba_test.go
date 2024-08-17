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
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/platforms/twitch"

	"github.com/google/go-cmp/cmp"
	"gorm.io/gorm"
)

func TestHasOutboundPendingDuels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc      string
		runBefore []func(testing.TB, *gorm.DB)
		want      int
	}{
		{
			desc:      "no outbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has outbound pending duels",
			runBefore: []func(testing.TB, *gorm.DB){
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			db := databasetest.New(t)
			if err := db.Where("1 = 1").Delete(&models.Duel{}).Error; err != nil {
				t.Fatal(err)
			}

			var user1 models.User
			err := db.First(&user1, models.User{
				TwitchID:   "user1",
				TwitchName: "user1",
			}).Error
			if err != nil {
				t.Fatalf("failed to find user1: %v", err)
			}

			for _, f := range tc.runBefore {
				f(t, db)
			}

			got, err := OutboundPendingDuels(&user1, 30*time.Second, db)
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
		runBefore []func(testing.TB, *gorm.DB)
		want      int
	}{
		{
			desc:      "no inbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has inbound pending duels",
			runBefore: []func(testing.TB, *gorm.DB){
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			db := databasetest.New(t)
			if err := db.Where("1 = 1").Delete(&models.Duel{}).Error; err != nil {
				t.Fatal(err)
			}

			var user2 models.User
			err := db.First(&user2, models.User{
				TwitchID:   "user2",
				TwitchName: "user2",
			}).Error
			if err != nil {
				t.Fatalf("failed to find user2: %v", err)
			}

			for _, f := range tc.runBefore {
				f(t, db)
			}

			got, err := InboundPendingDuels(&user2, 30*time.Second, db)
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
	db := databasetest.New(t)
	server := newTestServer()

	if err := db.Where("1 = 1").Delete(&models.User{}).Error; err != nil {
		t.Fatal(err)
	}

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	messages := []models.Message{
		{
			User:    user1,
			Channel: "channel1",
			Text:    "something",
			Time:    time.Now().Add(-1 * time.Minute),
		},
		{
			User:    user2,
			Channel: "channel1",
			Text:    "something else",
			Time:    time.Now().Add(-50 * time.Minute),
		},
	}
	for i, m := range messages {
		if err := db.Create(&m).Error; err != nil {
			t.Fatalf("failed to create message %d: %v", i, err)
		}
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTesting(server.URL, db),
	}

	grantPoints(ps, db)

	var transactions []models.GambaTransaction
	if err := db.Find(&transactions).Error; err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 gamba transactions, found %d: %v", len(transactions), transactions)
	}

	if transactions[0].Delta != int64(activeGrantAmount) {
		t.Errorf("transaction 0 should have granted %d points, but granted %d", activeGrantAmount, transactions[0].Delta)
	}
	if transactions[0].UserID != user1.ID {
		t.Errorf("transaction 0 should have been for user 1 but was for user %v", transactions[0].User)
	}
	if transactions[1].Delta != int64(inactiveGrantAmount) {
		t.Errorf("transaction 1 should have granted %d points, but granted %d", inactiveGrantAmount, transactions[1].Delta)
	}
	if transactions[1].UserID != user2.ID {
		t.Errorf("transaction 1 should have been for user 2 but was for user %v", transactions[1].User)
	}
}

func TestGetInactiveUsers(t *testing.T) {
	t.Parallel()
	db := databasetest.New(t)
	server := newTestServer()

	if err := db.Where("1 = 1").Delete(&models.User{}).Error; err != nil {
		t.Fatal(err)
	}

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTesting(server.URL, db),
	}

	got := getInactiveUsers(ps, db)
	want := []models.User{user1, user2}

	if len(got) != len(want) {
		t.Fatalf("getInactiveUsers() got %d users, want %d: diff (-want +got):\n%s", len(got), len(want), cmp.Diff(want, got))
	}
	if got[0].ID != want[0].ID {
		t.Errorf("getInactiveUsers()[0].ID = %d want %d", got[0].ID, want[0].ID)
	}
}

func TestGetActiveUsers(t *testing.T) {
	t.Parallel()
	db := databasetest.New(t)

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	messages := []models.Message{
		{
			User:    user1,
			Channel: "channel1",
			Text:    "something",
			Time:    time.Now().Add(-1 * time.Minute),
		},
		{
			User:    user2,
			Channel: "channel1",
			Text:    "something else",
			Time:    time.Now().Add(-50 * time.Minute),
		},
	}
	for i, m := range messages {
		if err := db.Create(&m).Error; err != nil {
			t.Fatalf("failed to create message %d: %v", i, err)
		}
	}

	got, err := getActiveUsers(db)
	if err != nil {
		t.Fatalf("getActiveUsers() unexpected error: %v", err)
	}
	want := []models.User{user1}

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
					User: models.User{
						ID: 1,
					},
					IsActive: true,
				},
				{
					User: models.User{
						ID: 2,
					},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User: models.User{
						ID: 1,
					},
					IsActive: true,
				},
				{
					User: models.User{
						ID: 2,
					},
					IsActive: true,
				},
			},
		},
		{
			desc: "all inactive",
			input: []grant{
				{
					User: models.User{
						ID: 1,
					},
					IsActive: false,
				},
				{
					User: models.User{
						ID: 2,
					},
					IsActive: false,
				},
			},
			want: []grant{
				{
					User: models.User{
						ID: 1,
					},
					IsActive: false,
				},
				{
					User: models.User{
						ID: 2,
					},
					IsActive: false,
				},
			},
		},
		{
			desc: "inactive and active unsorted",
			input: []grant{
				{
					User: models.User{
						ID: 1,
					},
					IsActive: false,
				},
				{
					User: models.User{
						ID: 2,
					},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User: models.User{
						ID: 2,
					},
					IsActive: true,
				},
				{
					User: models.User{
						ID: 1,
					},
					IsActive: false,
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := deduplicateByUser(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("deduplicateByUser() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func startDuel(t testing.TB, db *gorm.DB) {
	t.Helper()
	var user1, user2 models.User
	err := db.First(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	}).Error
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	err = db.First(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	}).Error
	if err != nil {
		t.Fatalf("Failed to find user2: %v", err)
	}
	err = db.Create(&models.Duel{
		UserID:   user1.ID,
		User:     user1,
		TargetID: user2.ID,
		Target:   user2,
		Amount:   25,
		Pending:  true,
		Accepted: false,
	}).Error
	if err != nil {
		t.Fatalf("Failed to create duel: %v", err)
	}
}

func add50PointsToUser1(t testing.TB, db *gorm.DB) {
	t.Helper()
	var user models.User
	err := db.First(&user, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	}).Error
	if err != nil {
		t.Fatalf("Failed to find user1: %v", err)
	}
	add50PointsToUser(t, user, db)
}

func add50PointsToUser2(t testing.TB, db *gorm.DB) {
	t.Helper()
	var user models.User
	err := db.First(&user, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	}).Error
	if err != nil {
		t.Fatalf("Failed to find/create user2: %v", err)
	}
	add50PointsToUser(t, user, db)
}

func add50PointsToUser(t testing.TB, user models.User, db *gorm.DB) {
	t.Helper()
	txn := models.GambaTransaction{
		Game:  "FAKE - TEST",
		User:  user,
		Delta: 50,
	}
	if err := db.Create(&txn).Error; err != nil {
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
