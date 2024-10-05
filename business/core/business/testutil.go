package business

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/google/uuid"
)

func TestGenerateNewBusinesses(n int, userID uuid.UUID) []NewBusiness {
	newBsns := make([]NewBusiness, n)

	const lorem = `Lorem ipsum dolor sit amet, consectetur adipiscing elit,
	sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
	`
	const loremLen = len(lorem) - 1

	for i := 0; i < n; i++ {
		idx := rand.IntN(10000)

		nb := NewBusiness{
			Name:    fmt.Sprintf("Name%d", idx),
			OwnerID: userID,
			Desc:    lorem[:rand.IntN(loremLen)],
		}

		newBsns[i] = nb
	}

	return newBsns
}

func TestGenerateSeedBusinesses(n int, api *Core, userID uuid.UUID) ([]Business, error) {
	newBsns := TestGenerateNewBusinesses(n, userID)

	bsns := make([]Business, len(newBsns))
	for i, nb := range newBsns {
		b, err := api.Create(context.Background(), nb)
		if err != nil {
			return nil, fmt.Errorf("seeding business: idx: %d: %w", i, err)
		}

		bsns[i] = b
	}

	return bsns, nil
}
