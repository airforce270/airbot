// Package gamba handles gamba-related things.
package gamba

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/airforce270/airbot/base"
	"github.com/airforce270/airbot/database"
	"github.com/airforce270/airbot/utils/ptrs"
)

var (
	grantInterval             = 10 * time.Minute
	activeGrantAmount   int64 = 10
	inactiveGrantAmount int64 = 3
)

// StartGrantingPoints starts a loop to grant points to all chatters on an interval.
// This function blocks and should be run within a goroutine.
func StartGrantingPoints(ctx context.Context, ps map[string]base.Platform, queries *database.Queries) {
	timer := time.NewTicker(grantInterval)
	for {
		select {
		case <-ctx.Done():
			log.Print("Stopping point granting, context cancelled")
			return
		case <-timer.C:
			go grantPoints(ctx, ps, queries)
		}
	}
}

// grant represents a points grant that may be given to a user.
type grant struct {
	// User is the user the grant is for.
	User database.User
	// IsActive is whether the user is currently active.
	IsActive bool
}

// Persist persists the grant in the database. This is not an idempotent operation.
func (g grant) Persist(ctx context.Context, queries *database.Queries) error {
	amount := activeGrantAmount
	if !g.IsActive {
		amount = inactiveGrantAmount
	}
	_, err := queries.CreateGambaTransaction(ctx, database.CreateGambaTransactionParams{
		UserID: &g.User.ID,
		Game:   ptrs.Ptr("AutomaticGrant"),
		Delta:  ptrs.Ptr(int64(amount)),
	})
	if err != nil {
		return fmt.Errorf("failed to create automatic grant (user %d, amount %d): %w", g.User.ID, amount, err)
	}
	return nil
}

// OutboundPendingDuels returns the user's outbound pending duels.
func OutboundPendingDuels(ctx context.Context, user database.User, expire time.Duration, queries *database.Queries) ([]database.Duel, error) {
	duels, err := queries.SelectUserOutboundDuels(ctx, database.SelectUserOutboundDuelsParams{
		UserID:    &user.ID,
		CreatedAt: ptrs.Ptr(time.Now().Add(-expire)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pending outbound duels for user %d: %w", user.ID, err)
	}
	return duels, nil
}

// InboundPendingDuels returns the user's inbound pending duels.
func InboundPendingDuels(ctx context.Context, user database.User, expire time.Duration, queries *database.Queries) ([]database.Duel, error) {
	duels, err := queries.SelectUserInboundDuels(ctx, database.SelectUserInboundDuelsParams{
		TargetID:  &user.ID,
		CreatedAt: ptrs.Ptr(time.Now().Add(-expire)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pending inbound duels for user %d: %w", user.ID, err)
	}
	return duels, nil
}

// grantPoints performs a single point grant to all active and inactive users.
func grantPoints(ctx context.Context, ps map[string]base.Platform, queries *database.Queries) {
	var grants []grant

	for _, inactiveUser := range getInactiveUsers(ctx, ps) {
		grants = append(grants, grant{
			User:     inactiveUser,
			IsActive: false,
		})
	}

	activeUsers, err := getActiveUsers(ctx, queries)
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
	}
	for _, activeUser := range activeUsers {
		grants = append(grants, grant{
			User:     activeUser,
			IsActive: true,
		})
	}

	grants = deduplicateByUser(grants)

	for _, g := range grants {
		err := g.Persist(ctx, queries)
		if err != nil {
			log.Printf("Failed to grant points - failed to persist grant: %v", err)
			return
		}
	}
}

// getInactiveUsers gets all inactive users to grant points to.
// The users returned are not guaranteed to be inactive, the results returned are overinclusive.
func getInactiveUsers(ctx context.Context, ps map[string]base.Platform) []database.User {
	var users []database.User
	for _, p := range ps {
		allUsers, err := p.CurrentUsers()
		if err != nil {
			log.Printf("Failed to retrieve users from %s: %v", p.Name(), err)
			continue
		}

		for _, u := range allUsers {
			user, err := p.User(ctx, u)
			if err != nil {
				if errors.Is(err, base.ErrUserUnknown) {
					// user needs to type something somewhere before they can get points automatically
					continue
				}
				log.Printf("Failed to retrieve user from %s: %v", p.Name(), err)
				continue
			}
			users = append(users, user)
		}
	}
	return users
}

// getActiveUsers gets all active users to grant points to.
// The users returned are guaranteed to be active.
func getActiveUsers(ctx context.Context, queries *database.Queries) ([]database.User, error) {
	activeUsers, err := queries.SelectActiveUsers(ctx, ptrs.Ptr(time.Now().Add(-grantInterval)))
	if err != nil {
		return nil, fmt.Errorf("failed to select recent user: %w", err)
	}
	return activeUsers, nil
}

func deduplicateByUser(grants []grant) []grant {
	sorted := grants
	slices.SortStableFunc(sorted, func(g1, g2 grant) int {
		if g1.IsActive && g2.IsActive {
			return 0
		}
		if g1.IsActive && !g2.IsActive {
			return -1
		}
		return 1
	})
	var deduped []grant
	var grantedIDs []int64
	for _, g := range sorted {
		if slices.Contains(grantedIDs, g.User.ID) {
			continue
		}
		grantedIDs = append(grantedIDs, g.User.ID)
		deduped = append(deduped, g)
	}
	return deduped
}
