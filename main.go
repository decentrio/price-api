package main

import (
	"context"
	"log"
	"net"
	"net/http"

	_ "github.com/decentrio/price-api/docs/statik"
	"github.com/rakyll/statik/fs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/decentrio/price-api/app"
	"github.com/decentrio/price-api/database"

	"github.com/decentrio/price-api/app/ticker"
	"github.com/decentrio/price-api/app/trade"
	tickertypes "github.com/decentrio/price-api/types/ticker"
	tradetypes "github.com/decentrio/price-api/types/trade"
)

func initModule() []app.AppModule {
	dbHandler := database.NewDBHandler()

	tickerKeeper := ticker.NewKeeper(dbHandler)
	tradeKeeper := trade.NewKeeper(dbHandler)

	modules := []app.AppModule{
		ticker.NewAppModule(*tickerKeeper),
		trade.NewAppModule(*tradeKeeper),
	}

	return modules
}

func runGRPCServer() error {
	lis, err := net.Listen("tcp", ":5060")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	modules := initModule()
	bookApp := app.NewApp(s, modules)
	bookApp.RegisterServices()
	return s.Serve(lis)
}

func runHTTPServer() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := tickertypes.RegisterTickerQueryHandlerFromEndpoint(ctx, mux, ":5060", opts)
	if err != nil {
		return err
	}

	err = tradetypes.RegisterTradeQueryHandlerFromEndpoint(ctx, mux, ":5060", opts)
	if err != nil {
		return err
	}

	http.Handle("/", mux)
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)

	// Serve Swagger UI

	http.Handle("/public/", http.StripPrefix("/public/", staticServer))

	log.Println("HTTP server listening on :5050")
	return http.ListenAndServe(":5050", nil)
}

func main() {
	go func() {
		if err := runGRPCServer(); err != nil {
			log.Fatalf("failed to run gRPC server: %v", err)
		}
	}()
	if err := runHTTPServer(); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}
