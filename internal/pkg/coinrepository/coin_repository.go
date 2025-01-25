package coinrepository

type CoinRepository interface {
	CoinCommand
	CoinQuery

	BannedCoinCommand
	BannedCoinQuery

	TradeCommand
	ListTradesQuery
}
