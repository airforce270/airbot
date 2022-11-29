package gamba

import (
	"testing"

	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/database/databasetest"
	"github.com/airforce270/airbot/database/models"
)

func TestFetchUserPoints(t *testing.T) {
	db := databasetest.NewFakeDB()
	var user1 models.User
	result := db.FirstOrCreate(&user1, models.User{
		TwitchID:   "user1",
		TwitchName: "user1",
	})
	if result.Error != nil {
		t.Fatalf("failed to find/create user1: %v", result.Error)
	}
	var user2 models.User
	result = db.FirstOrCreate(&user2, models.User{
		TwitchID:   "user2",
		TwitchName: "user2",
	})
	if result.Error != nil {
		t.Fatalf("failed to find/create user2: %v", result.Error)
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
			db.Migrator().DropTable(&models.GambaTransaction{})
			database.Migrate(db)

			for _, txn := range tc.transactions {
				result = db.Create(&txn)
				if result.Error != nil {
					t.Fatalf("failed to insert gamba transaction: %v", result.Error)
				}
			}

			if got := fetchUserPoints(db, user1); got != tc.want {
				t.Errorf("fetchUserPoints() = %d, want %d", got, tc.want)
			}
		})
	}
}
