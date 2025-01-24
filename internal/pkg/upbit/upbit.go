package upbit

import (
	"context"
	"encoding/json"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/http"
	"github.com/pkg/errors"
)

type Upbit struct {
	client *http.Client
}

func NewUpbit() *Upbit {
	client := http.NewClient()
	return &Upbit{client: client}
}

// ListCoins implements coinservice.CoinService.
func (u *Upbit) ListCoins(ctx context.Context) ([]*Coin, error) {
	resp, err := u.client.Get(ctx, "https://api.upbit.com/v1/market/all?is_details=true")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(resp.Body))
	}
	ret, err := UnmarshalSlice[Coin](resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}

// ListTrades implements coinservice.CoinService.
func (s *Service) ListTrades(ctx context.Context, coinID domain.CoinID) ([]*domain.Trade, error) {
	panic("unimplemented")
}

func Unmarshal[T any](body []byte) (*T, error) {
	var ret T
	err := json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ret, nil
}

func UnmarshalSlice[T any](body []byte) ([]*T, error) {
	var ret []*T
	err := json.Unmarshal(body, &ret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, nil
}
