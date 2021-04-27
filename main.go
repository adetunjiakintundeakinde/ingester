package main

import (
	"context"
	"github.com/demoware/ingester/dispatcher"
	"github.com/demoware/ingester/metrics"
	"github.com/demoware/ingester/payload_handler/cpu_handler"
	"github.com/demoware/ingester/payload_handler/last_kernel_update_handler"
	"github.com/demoware/ingester/payload_handler/load_average_handler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	options := dispatcher.DispatcherOptions{
		BatchLimit: 10,
		MetricsUri: "http://localhost:8080/metrics",
		AuthToken:  "deadbeef",
	}
	dispatcher := options.NewDispatcher()
	cpuHandler := cpu_handler.NewCpuHandler()
	lastKernelUpdateHandler := last_kernel_update_handler.NewLastKernelUpgradeHandler()
	loadAverageHandler := load_average_handler.NewLoadAverageHandler()

	dispatcher.NewHandler(cpuHandler)
	dispatcher.NewHandler(lastKernelUpdateHandler)
	dispatcher.NewHandler(loadAverageHandler)

	go dispatcher.Start(ctx)
	go cpuHandler.Start(ctx)
	go lastKernelUpdateHandler.Start(ctx)
	go loadAverageHandler.Start(ctx)

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Default)
	http.ListenAndServe(":9020", mux)

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan
	cancelFunc()
}
