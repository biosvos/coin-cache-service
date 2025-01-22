package flow_test

import (
	"testing"

	"github.com/biosvos/coin-cache-service/internal/app/flow"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	t.Parallel()
	service := flow.NewService()

	err := service.Usecase()

	require.Error(t, err)
}
