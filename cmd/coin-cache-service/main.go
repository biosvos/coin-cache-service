package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/biosvos/coin-cache-service/internal/app/flow"
	"github.com/biosvos/coin-cache-service/internal/app/miner"
	"github.com/biosvos/coin-cache-service/internal/app/prohibitor"
	"github.com/biosvos/coin-cache-service/internal/app/trader"
	"github.com/biosvos/coin-cache-service/internal/pkg/buses/local"
	"github.com/biosvos/coin-cache-service/internal/pkg/domain"
	"github.com/biosvos/coin-cache-service/internal/pkg/realrepository"
	"github.com/biosvos/coin-cache-service/internal/pkg/upbit"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
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
	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "list.coins",
		Summary:     "List coins",
		Method:      http.MethodGet,
		Path:        "/coins",
	}, func(ctx context.Context, _ *ListCoinsRequest) (*ListCoinsResponse, error) {
		ret, err := service.ListCoins(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		resp := &ListCoinsResponse{
			Body: &ListCoinsBody{
				Coins: ret,
			},
		}
		return resp, nil
	})
	huma.Register(api, huma.Operation{ //nolint:exhaustruct
		OperationID: "list.trades",
		Summary:     "List trades",
		Method:      http.MethodGet,
		Path:        "/trades/{coinID}",
	}, func(ctx context.Context, input *ListTradesRequest) (*ListTradesResponse, error) {
		ret, err := service.ListTrades(ctx, domain.CoinID(input.CoinID))
		if err != nil {
			return nil, errors.WithStack(err)
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
	logger, _ := zap.NewProduction()
	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()
	ctx := context.Background()

	service := upbit.NewService()
	repo := realrepository.NewRepository("/tmp/coins")
	bus := local.NewBus(logger)
	mine := miner.NewMiner(logger, service, repo, bus)
	prohibitor := prohibitor.NewProhibitor(logger, bus, repo)
	err := prohibitor.Start(ctx)
	if err != nil {
		panic(err)
	}
	trader := trader.NewTrader(logger, bus, service, repo)
	trader.Start(ctx)

	err = mine.Start()
	if err != nil {
		panic(err)
	}
	defer mine.Stop()

	flowService := flow.NewService(repo)

	cli := newClient(ctx, flowService)

	cli.Run()
}

func newClient(ctx context.Context, flowService *flow.Service) humacli.CLI {
	return humacli.New(func(hooks humacli.Hooks, options *Options) {
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
			ctx, cancel := context.WithTimeout(ctx, giveUpTimeout)
			defer cancel()
			err := server.Shutdown(ctx)
			if err != nil {
				log.Printf("failed to shutdown server: %v", err)
			}
		})
	})
}
