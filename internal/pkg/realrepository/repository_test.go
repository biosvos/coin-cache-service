package realrepository_test

import (
	"context"
	"testing"
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/realrepository"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateCoin(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo := realrepository.NewRepository(dir)
	now := time.Now()
	ctx := context.Background()
	domainCoin := domain.NewCoin("A", false, now)

	ret, err := repo.CreateCoin(ctx, domainCoin)

	require.NoError(t, err)
	require.Equal(t, domain.CoinID("A"), ret.ID())
	require.False(t, ret.IsDanger())
	require.Equal(t, now, ret.ModifiedAt())
}

func TestRepository_ListCoins(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo := realrepository.NewRepository(dir)
	now := time.Now()
	ctx := context.Background()
	domainCoin := domain.NewCoin("A", false, now)
	_, _ = repo.CreateCoin(ctx, domainCoin)

	coins, err := repo.ListCoins(ctx)

	require.NoError(t, err)
	require.Len(t, coins, 1)
	require.Equal(t, domain.CoinID("A"), coins[0].ID())
	require.False(t, coins[0].IsDanger())
	require.Equal(t, now.Unix(), coins[0].ModifiedAt().Unix())
}
