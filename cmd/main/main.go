package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tender_service/internal/handlers/bids/bid_feedback"
	"tender_service/internal/handlers/bids/bid_submit_decision"
	"tender_service/internal/handlers/bids/bids_rollback"
	"tender_service/internal/handlers/bids/get_bid_status"
	"tender_service/internal/handlers/bids/get_bids"
	"tender_service/internal/handlers/bids/get_my_bids"
	"tender_service/internal/handlers/bids/get_reviews"
	"tender_service/internal/handlers/bids/new"
	"tender_service/internal/handlers/bids/patch_bid"
	"tender_service/internal/handlers/bids/put_bid_status"
	"tender_service/internal/handlers/ping"
	"tender_service/internal/handlers/tenders/get_my_tenders"
	"tender_service/internal/handlers/tenders/get_tender_status"
	"tender_service/internal/handlers/tenders/get_tenders"
	"tender_service/internal/handlers/tenders/new_tender"
	"tender_service/internal/handlers/tenders/patch_tender_status"
	"tender_service/internal/handlers/tenders/put_tender_status"
	"tender_service/internal/handlers/tenders/tenders_rollback"
	psq "tender_service/internal/storage"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"tender_service/internal/config"
)

func main() {
	cfg := config.Load()
	storage := &psq.Storage{}

	ctx, cancel := context.WithCancel(context.Background())
	log := setuplogger()
	go func() {
		err := psq.New(cancel, storage, cfg)
		if err != nil {
			log.Error("failed to init storage", slog.String("error", err.Error()))
		}
	}()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/api", func(r chi.Router) {
		r.Route("/tenders", func(r chi.Router) {
			r.Post("/new", new_tender.New(storage))
			r.Get("/{tenderId}/status", gettenderstatus.New(storage))
			r.Put("/{tenderId}/status", puttenderstatus.New(storage))
			r.Patch("/{tenderId}/edit", patchtenderstatus.New(storage))
			r.Get("/", gettenders.New(storage))
			r.Get("/my", getmytenders.New(storage))
			r.Put("/{tenderId}/rollback/{version}", tendersrollback.New(storage))

		})

		r.Route("/bids", func(r chi.Router) {
			r.Post("/new", newbid.New(storage))
			r.Get("/{bidId}/status", getbidstatus.New(storage))
			r.Put("/{bidId}/status", putbidstatus.New(storage))
			r.Patch("/{bidId}/edit", patchbid.New(storage))
			r.Get("/my", getmybids.New(storage))
			r.Get("/{tenderId}/list", getbids.New(storage))
			r.Put("/{bidId}/submit_decision", bidsubmitdecision.New(storage))
			r.Put("/{bidId}/feedback", bidfeedback.New(storage))
			r.Get("/{tenderId}/reviews", getreviews.New(storage))
			r.Put("/{bidId}/rollback/{version}", bidsrollback.New(storage))

		})

		r.Get("/ping", ping.New(ctx))
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  time.Minute,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancelCtx := context.WithTimeout(serverCtx, 15*time.Second)
		defer cancelCtx()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				panic("graceful shutdown timed out.. forcing exit.")
			}
		}()

		log.Info("server shut down")
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			panic(err.Error())
		}
		serverStopCtx()
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err.Error())
	}

	<-serverCtx.Done()

	log.Info("server stopped")

}

func setuplogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
