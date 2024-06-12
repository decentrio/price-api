package main

import (
	"context"
	"log"
	"net"
	"net/http"

	_ "github.com/decentrio/price-api/docs/statik"
	"github.com/rakyll/statik/fs"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/decentrio/price-api/app"
	// "github.com/decentrio/price-api/database"
)

func initModule() []app.AppModule {
	// dbHandler := database.NewDBHandler()

	// contractKeeper := contract.NewKeeper(dbHandler)
	// eventKeeper := event.NewKeeper(dbHandler)
	// ledgerKeeper := ledger.NewKeeper(dbHandler)
	// transactionKeeper := transaction.NewKeeper(dbHandler)

	// modules := []app.AppModule{
	// 	contract.NewAppModule(*contractKeeper),
	// 	event.NewAppModule(*eventKeeper),
	// 	ledger.NewAppModule(*ledgerKeeper),
	// 	transaction.NewAppModule(*transactionKeeper),
	// }

	return []app.AppModule{}
}

func runGRPCServer() error {
	lis, err := net.Listen("tcp", ":9090")
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
	// opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// err := contracttypes.RegisterContractQueryHandlerFromEndpoint(ctx, mux, ":9090", opts)
	// if err != nil {
	// 	return err
	// }

	// err = eventtypes.RegisterEventQueryHandlerFromEndpoint(ctx, mux, ":9090", opts)
	// if err != nil {
	// 	return err
	// }

	// err = ledgertypes.RegisterLedgerQueryHandlerFromEndpoint(ctx, mux, ":9090", opts)
	// if err != nil {
	// 	return err
	// }

	// err = txtypes.RegisterTransactionQueryHandlerFromEndpoint(ctx, mux, ":9090", opts)
	// if err != nil {
	// 	return err
	// }

	http.Handle("/", mux)
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)

	// Serve Swagger UI

	http.Handle("/public/", http.StripPrefix("/public/", staticServer))

	log.Println("HTTP server listening on :8080")
	return http.ListenAndServe(":8080", nil)
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
