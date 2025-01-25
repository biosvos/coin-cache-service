package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/biosvos/coin-cache-service/internal/app/flow"
	"github.com/biosvos/coin-cache-service/internal/app/miner"
	"github.com/biosvos/coin-cache-service/internal/app/trader"
	"github.com/biosvos/coin-cache-service/internal/pkg/buses/local"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/real"
	"github.com/biosvos/coin-cache-service/internal/pkg/upbit"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Options struct {
	Port int `default:"8888" help:"Port to listen on" short:"p"`
}

type ListCoinsBody struct {
	Coins []string
}

type ListCoinsRequest struct {
}

type ListCoinsResponse struct {
	Body *ListCoinsBody `doc:"Body" json:"body"`
}

type TradeBody struct {
	Date  time.Time
	Price string
}

type ListTradesBody struct {
	Trades []*TradeBody
}

type ListTradesRequest struct {
	CoinID string `path:"coinID"`
}

type ListTradesResponse struct {
	Body *ListTradesBody `doc:"Body" json:"body"`
}

func AddRoutes(api huma.API, service *flow.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "list.coins",
		Summary:     "List coins",
		Method:      http.MethodGet,
		Path:        "/coins",
	}, func(ctx context.Context, input *ListCoinsRequest) (*ListCoinsResponse, error) {
		ret, err := service.ListCoins(ctx)
		if err != nil {
			return nil, err
		}
		resp := &ListCoinsResponse{
			Body: &ListCoinsBody{
				Coins: ret,
			},
		}
		return resp, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "list.trades",
		Summary:     "List trades",
		Method:      http.MethodGet,
		Path:        "/trades/{coinID}",
	}, func(ctx context.Context, input *ListTradesRequest) (*ListTradesResponse, error) {
		ret, err := service.ListTrades(ctx, domain.CoinID(input.CoinID))
		if err != nil {
			return nil, err
		}
		var trades []*TradeBody
		for _, trade := range ret.Trades() {
			trades = append(trades, &TradeBody{
				Date:  trade.Date(),
				Price: string(trade.LastPrice()),
			})
		}
		resp := &ListTradesResponse{
			Body: &ListTradesBody{
				Trades: trades,
			},
		}
		return resp, nil
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	service := upbit.NewService()
	repo := real.NewRepository("/tmp/coins")
	bus := local.NewBus(logger)
	mine := miner.NewMiner(logger, service, repo, bus)

	trader := trader.NewTrader(logger, bus, service, repo)
	trader.Start(context.Background())

	err := mine.Mine(context.Background())
	if err != nil {
		panic(err)
	}

	flowService := flow.NewService(repo)

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		router := chi.NewMux()
		api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

		AddRoutes(api, flowService)

		const (
			readTimeout       = 5 * time.Second
			writeTimeout      = 5 * time.Second
			idleTimeout       = 30 * time.Second
			readHeaderTimeout = 2 * time.Second
			giveUpTimeout     = 5 * time.Second
		)
		server := &http.Server{ //nolint:exhaustruct
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ReadHeaderTimeout: readHeaderTimeout,

			Addr:    fmt.Sprintf(":%v", options.Port),
			Handler: router,
		}

		hooks.OnStart(func() {
			err := server.ListenAndServe()
			if err != nil {
				log.Printf("failed to start server: %v", err)
			}
		})

		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), giveUpTimeout)
			defer cancel()
			err := server.Shutdown(ctx)
			if err != nil {
				log.Printf("failed to shutdown server: %v", err)
			}
		})
	})

	cli.Run()
}
