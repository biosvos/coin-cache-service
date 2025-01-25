package upbit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

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
func (u *Upbit) ListDayCandles(ctx context.Context, market string, count int) ([]*Candle, error) {
	if count < 1 || 200 < count {
		return nil, errors.New("count must be between 1 and 200")
	}
	values := url.Values{}
	values.Add("market", market)
	countString := strconv.FormatInt(int64(count), 10)
	values.Add("count", countString)
	encodedValues := values.Encode()
	url := fmt.Sprintf("https://api.upbit.com/v1/candles/days?%v", encodedValues)
	resp, err := u.client.Get(ctx, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, http.ErrTooManyRequests
		}
		return nil, errors.New(string(resp.Body))
	}
	candles, err := UnmarshalSlice[Candle](resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return candles, nil
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
