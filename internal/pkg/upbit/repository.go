package upbit

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/biosvos/coin-cache-service/internal/app/coinservice"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/http"
	"github.com/pkg/errors"
)

var _ coinservice.CoinService = (*Service)(nil)

type Service struct {
	upbit *Upbit
}

func NewService() *Service {
	upbit := NewUpbit()
	return &Service{upbit: upbit}
}

func (s *Service) ListCoins(ctx context.Context) ([]*domain.Coin, error) {
	now := time.Now()
	coins, err := s.upbit.ListCoins(ctx)
	if err != nil {
		return nil, err
	}
	var ret []*domain.Coin
	for _, coin := range coins {
		if !strings.HasPrefix(coin.Market, "KRW-") {
			// 한국 화폐만 허용한다.
			continue
		}
		ret = append(ret, coin.ToDomain(now))
	}
	return ret, nil
}

// ListTrades implements coinservice.CoinService.
func (s *Service) ListTrades(ctx context.Context, coinID domain.CoinID) (*domain.Trades, error) {
	var candles []*Candle
	now := time.Now()
	err := retry(func() error {
		var err error
		candles, err = s.upbit.ListDayCandles(ctx, string(coinID), 20)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var ret []*domain.Trade
	for _, candle := range candles {
		dateTime, err := time.ParseInLocation("2006-01-02T15:04:05", candle.CandleDateTimeUtc, time.UTC)
		if err != nil {
			return nil, err
		}

		tradePrice := strconv.FormatFloat(candle.TradePrice, 'f', -1, 64)
		openingPrice := strconv.FormatFloat(candle.OpeningPrice, 'f', -1, 64)
		highPrice := strconv.FormatFloat(candle.HighPrice, 'f', -1, 64)
		lowPrice := strconv.FormatFloat(candle.LowPrice, 'f', -1, 64)

		ret = append(ret, domain.NewTrade(
			dateTime,
			domain.Price(tradePrice),
			domain.Price(openingPrice),
			domain.Price(highPrice),
			domain.Price(lowPrice),
		))
	}
	return domain.NewTrades(coinID, now, ret), nil
}

func retry(fn func() error) error {
	const retryCount = 60
	for range retryCount {
		err := fn()
		if err != nil {
			if errors.Is(err, http.ErrTooManyRequests) {
				time.Sleep(1 * time.Second)
				continue
			}
			return errors.WithStack(err)
		}
		return nil
	}
	return errors.New("too many requests")
}
