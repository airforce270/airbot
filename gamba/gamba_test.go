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
	"github.com/airforce270/airbot/apiclients/twitchtmi"
	"github.com/airforce270/airbot/apiclients/twitchtmitest"
	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/database/models"
	"github.com/airforce270/airbot/platforms/twitch"
	"github.com/airforce270/airbot/testing/databasetest"

	"github.com/google/go-cmp/cmp"
	"gorm.io/gorm"
)

func TestHasOutboundPendingDuels(t *testing.T) {
	db := databasetest.NewFakeDB()
	db.Where("1 = 1").Delete(&models.Duel{})

	var user1 models.User
	result := db.First(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	})
	if result.Error != nil {
		t.Fatalf("failed to find user1: %v", result.Error)
	}

	tests := []struct {
		desc      string
		runBefore []func() error
		want      int
	}{
		{
			desc:      "no outbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has outbound pending duels",
			runBefore: []func() error{
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			for i, f := range tc.runBefore {
				if err := f(); err != nil {
					t.Fatalf("runBefore[%d] failed: %v", i, err)
				}
			}

			got, err := OutboundPendingDuels(&user1, time.Duration(30)*time.Second, db)
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
	db := databasetest.NewFakeDB()
	db.Where("1 = 1").Delete(&models.Duel{})

	var user2 models.User
	result := db.First(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	})
	if result.Error != nil {
		t.Fatalf("failed to find user2: %v", result.Error)
	}

	tests := []struct {
		desc      string
		runBefore []func() error
		want      int
	}{
		{
			desc:      "no inbound pending duels",
			runBefore: nil,
			want:      0,
		},
		{
			desc: "has inbound pending duels",
			runBefore: []func() error{
				add50PointsToUser1,
				add50PointsToUser2,
				startDuel,
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			for i, f := range tc.runBefore {
				if err := f(); err != nil {
					t.Fatalf("runBefore[%d] failed: %v", i, err)
				}
			}

			got, err := InboundPendingDuels(&user2, time.Duration(30)*time.Second, db)
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
	db := databasetest.NewFakeDB()
	server := newTestServer()
	setFakes(server.URL)
	defer resetFakes()

	db.Where("1 = 1").Delete(&models.User{})

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	result := db.Create(&user1)
	if result.Error != nil {
		t.Fatalf("failed to create user1: %v", result.Error)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	result = db.Create(&user2)
	if result.Error != nil {
		t.Fatalf("failed to create user2: %v", result.Error)
	}

	messages := []models.Message{
		{
			User:    user1,
			Channel: "channel1",
			Text:    "something",
			Time:    time.Now().Add(time.Duration(-1) * time.Minute),
		},
		{
			User:    user2,
			Channel: "channel1",
			Text:    "something else",
			Time:    time.Now().Add(time.Duration(-50) * time.Minute),
		},
	}
	for i, m := range messages {
		result := db.Create(&m)
		if result.Error != nil {
			t.Fatalf("failed to create message %d: %v", i, result.Error)
		}
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTesting(server.URL, db),
	}

	grantPoints(ps, db)

	var transactions []models.GambaTransaction
	result = db.Find(&transactions)
	if result.Error != nil {
		panic(result.Error)
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
	db := databasetest.NewFakeDB()
	server := newTestServer()
	setFakes(server.URL)
	defer resetFakes()

	db.Where("1 = 1").Delete(&models.User{})

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	result := db.Create(&user1)
	if result.Error != nil {
		t.Fatalf("failed to create user1: %v", result.Error)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	result = db.Create(&user2)
	if result.Error != nil {
		t.Fatalf("failed to create user2: %v", result.Error)
	}

	ps := map[string]base.Platform{
		"FakeTwitch": twitch.NewForTesting(server.URL, db),
	}

	got := getInactiveUsers(ps, db)
	want := []models.User{user2}

	if len(got) != len(want) {
		t.Fatalf("getInactiveUsers() got %d users, want %d: got: %v, want: %v", len(got), len(want), got, want)
	}

	if got[0].ID != want[0].ID {
		t.Errorf("getInactiveUsers()[0].ID = %d want %d", got[0].ID, want[0].ID)
	}
}

func TestGetActiveUsers(t *testing.T) {
	db := databasetest.NewFakeDB()

	user1 := models.User{TwitchID: "user1", TwitchName: "user1"}
	result := db.Create(&user1)
	if result.Error != nil {
		t.Fatalf("failed to create user1: %v", result.Error)
	}
	user2 := models.User{TwitchID: "user2", TwitchName: "user2"}
	result = db.Create(&user2)
	if result.Error != nil {
		t.Fatalf("failed to create user2: %v", result.Error)
	}

	messages := []models.Message{
		{
			User:    user1,
			Channel: "channel1",
			Text:    "something",
			Time:    time.Now().Add(time.Duration(-1) * time.Minute),
		},
		{
			User:    user2,
			Channel: "channel1",
			Text:    "something else",
			Time:    time.Now().Add(time.Duration(-50) * time.Minute),
		},
	}
	for i, m := range messages {
		result := db.Create(&m)
		if result.Error != nil {
			t.Fatalf("failed to create message %d: %v", i, result.Error)
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
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: true,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
					},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User: models.User{
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: true,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
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
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: false,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
					},
					IsActive: false,
				},
			},
			want: []grant{
				{
					User: models.User{
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: false,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
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
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: false,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
					},
					IsActive: true,
				},
			},
			want: []grant{
				{
					User: models.User{
						Model: gorm.Model{
							ID: 2,
						},
					},
					IsActive: true,
				},
				{
					User: models.User{
						Model: gorm.Model{
							ID: 1,
						},
					},
					IsActive: false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := deduplicateByUser(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("deduplicateByUser() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func startDuel() error {
	db := databasetest.NewFakeDBConn()
	var user1, user2 models.User
	result := db.First(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	})
	if result.Error != nil {
		return fmt.Errorf("failed to find user1: %v", result.Error)
	}
	result = db.First(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	})
	if result.Error != nil {
		return fmt.Errorf("failed to find user2: %v", result.Error)
	}
	result = db.Create(&models.Duel{
		UserID:   user1.ID,
		User:     user1,
		TargetID: user2.ID,
		Target:   user2,
		Amount:   25,
		Pending:  true,
		Accepted: false,
	})
	return result.Error
}

func add50PointsToUser1() error {
	db := databasetest.NewFakeDBConn()
	var user models.User
	result := db.First(&user, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	})
	if result.Error != nil {
		return fmt.Errorf("failed to find user1: %v", result.Error)
	}
	return add50PointsToUser(user, db)
}

func add50PointsToUser2() error {
	db := databasetest.NewFakeDBConn()
	var user models.User
	result := db.First(&user, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	})
	if result.Error != nil {
		return fmt.Errorf("failed to find/create user2: %v", result.Error)
	}
	return add50PointsToUser(user, db)
}

func add50PointsToUser(user models.User, db *gorm.DB) error {
	txn := models.GambaTransaction{
		Game:  "FAKE - TEST",
		User:  user,
		Delta: 50,
	}
	result := db.Create(&txn)
	if result.Error != nil {
		return fmt.Errorf("failed to insert gamba transaction: %v", result.Error)
	}
	return nil
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/group/user/user1/chatters") {
			fmt.Fprint(w, twitchtmitest.FetchChattersSingleChatterResp)
		} else if strings.Contains(r.URL.Path, "/group/user/user2/chatters") {
			fmt.Fprint(w, twitchtmitest.FetchChattersSingleChatterResp)
		} else if strings.Contains(r.URL.Path, "/users") {
			fmt.Fprint(w, twitchtest.GetUsersResp)
		} else {
			log.Printf("Unknown URL sent to test server: %s", r.URL.Path)
		}
	}))
}

var savedTwitchTmiBaseURL = twitchtmi.BaseURL

func setFakes(url string) {
	twitchtmi.BaseURL = url
}

func resetFakes() {
	twitchtmi.BaseURL = savedTwitchTmiBaseURL
}
