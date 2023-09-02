package gamba

import (
	"testing"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/database/models"
)

func TestFetchUserPoints(t *testing.T) {
	db := databasetest.NewFakeDB(t)
	var user1 models.User
	err := db.FirstOrCreate(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	}).Error
	if err != nil {
		t.Fatalf("failed to find/create user1: %v", err)
	}
	var user2 models.User
	err = db.FirstOrCreate(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	}).Error
	if err != nil {
		t.Fatalf("failed to find/create user2: %v", err)
	}

	tests := []struct {
		desc         string
		transactions []models.GambaTransaction
		want         int64
	}{
		{
			desc:         "no transactions",
			transactions: []models.GambaTransaction{},
			want:         0,
		},
		{
			desc: "single transaction for single user",
			transactions: []models.GambaTransaction{
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 50,
				},
			},
			want: 50,
		},
		{
			desc: "multiple transactions for single user",
			transactions: []models.GambaTransaction{
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 50,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: -20,
				},
			},
			want: 30,
		},
		{
			desc: "multiple transactions +/- for single user",
			transactions: []models.GambaTransaction{
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 50,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: -20,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 5,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 100,
				},
			},
			want: 135,
		},
		{
			desc: "single transaction for other user",
			transactions: []models.GambaTransaction{
				{
					Game:  "FAKE - TEST",
					User:  user2,
					Delta: 50,
				},
			},
			want: 0,
		},
		{
			desc: "multiple transactions +/- for multiple users",
			transactions: []models.GambaTransaction{
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 50,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: -20,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 5,
				},
				{
					Game:  "FAKE - TEST",
					User:  user1,
					Delta: 100,
				},
				{
					Game:  "FAKE - TEST",
					User:  user2,
					Delta: 50,
				},
				{
					Game:  "FAKE - TEST",
					User:  user2,
					Delta: -20,
				},
				{
					Game:  "FAKE - TEST",
					User:  user2,
					Delta: 5,
				},
				{
					Game:  "FAKE - TEST",
					User:  user2,
					Delta: 100,
				},
			},
			want: 135,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if err := db.Migrator().DropTable(&models.GambaTransaction{}); err != nil {
				t.Fatalf("failed to drop GambaTransaction table: %v", err)
			}
			if err := database.Migrate(db); err != nil {
				t.Fatalf("failed to migrate db: %v", err)
			}

			for _, txn := range tc.transactions {
				if err := db.Create(&txn).Error; err != nil {
					t.Fatalf("failed to insert gamba transaction: %v", err)
				}
			}

			got, err := fetchUserPoints(db, user1)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("fetchUserPoints() = %d, want %d", got, tc.want)
			}
		})
	}
}
