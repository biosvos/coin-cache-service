package coinservice

import (
	"github.com/biosvos/coin-cache-service/internal/pkg/coinrepository"
)

type CoinService interface {
	coinrepository.ListCoinsQuery
	coinrepository.ListTradesQuery
}
