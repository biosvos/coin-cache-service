package upbit

import (
	"time"

	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
)

type Coin struct {
	Market      string       `json:"market"`
	KoreanName  string       `json:"korean_name"`
	EnglishName string       `json:"english_name"`
	MarketEvent *MarketEvent `json:"market_event"`
}

func (c *Coin) IsHazardous() bool {
	return c.MarketEvent.IsHazardous()
}

type MarketEvent struct {
	Warning bool           `json:"warning"`
	Caution *MarketCaution `json:"caution"`
}

func (m *MarketEvent) IsHazardous() bool {
	if m == nil {
		return false
	}
	return m.Warning || m.Caution.IsHazardous()
}

type MarketCaution struct {
	PRICEFLUCTUATIONS            bool `json:"PRICE_FLUCTUATIONS"`
	TRADINGVOLUMESOARING         bool `json:"TRADING_VOLUME_SOARING"`
	DEPOSITAMOUNTSOARING         bool `json:"DEPOSIT_AMOUNT_SOARING"`
	GLOBALPRICEDIFFERENCES       bool `json:"GLOBAL_PRICE_DIFFERENCES"`
	CONCENTRATIONOFSMALLACCOUNTS bool `json:"CONCENTRATION_OF_SMALL_ACCOUNTS"`
}

func (m *MarketCaution) IsHazardous() bool {
	if m == nil {
		return false
	}
	return m.PRICEFLUCTUATIONS ||
		m.TRADINGVOLUMESOARING ||
		m.DEPOSITAMOUNTSOARING ||
		m.GLOBALPRICEDIFFERENCES ||
		m.CONCENTRATIONOFSMALLACCOUNTS
}

// ToDomain
func (c *Coin) ToDomain(now time.Time) *domain.Coin {
	if c == nil {
		return nil
	}
	return domain.NewCoin(
		domain.CoinID(c.Market),
		c.IsHazardous(),
		now,
	)
}
