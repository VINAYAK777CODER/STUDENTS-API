package main

import (
	"context"   // Provides cancellation, deadlines → used for graceful shutdown
	"fmt"       // For printing messages to console
	"log/slog"  // Modern structured logger (Go 1.21+)
	"net/http"  // HTTP server, routing, Request/Response
	"os"        // Access OS features (signals, env, process)
	"os/signal" // Used to catch CTRL+C or shutdown signals
	"syscall"   // Provides OS-level signals like SIGTERM, SIGINT
	"time"      // For timeouts: graceful shutdown timeout duration

	"github.com/VINAYAK777CODER/STUDENTS-API/internal/config" // Custom config loader
	"github.com/VINAYAK777CODER/STUDENTS-API/internal/http/handlers/student"
)

func main() {

	//---------------------------------------------------------------------------
	// STEP 1 → Load configuration from config.yaml (using MustLoad)
	// MustLoad() reads YAML file, environment variables & loads settings.
	// It returns a Config struct → cfg
	//---------------------------------------------------------------------------
	cfg := config.MustLoad()



	//---------------------------------------------------------------------------
	// STEP 2 → Setup router (HTTP multiplexer)
	// http.NewServeMux creates a new router which maps routes to handler functions.
	// This router will receive and route all HTTP requests.
	//---------------------------------------------------------------------------
	router := http.NewServeMux()



	//---------------------------------------------------------------------------
	// STEP 3 → Register a route handler
	//
	// HandleFunc pattern: router.HandleFunc("METHOD /PATH", handlerFunc)
	//
	// "GET /" means:
	//   - Method must be GET
	//   - Path must be "/"
	// This requires Go 1.22+ (new HTTP pattern matching)
	//
	// Handler function parameters:
	//   w → ResponseWriter (we write response back to the client)
	//   r → Request (contains request data)
	//---------------------------------------------------------------------------
	router.HandleFunc("POST /api/students", student.New())



	//---------------------------------------------------------------------------
	// STEP 4 → Create HTTP Server instance
	//
	// http.Server struct holds:
	//   Addr    → Address where server listens (like ":8080")
	//   Handler → Router handling all requests
	//
	// cfg.HTTPServer.Addr comes from your YAML config:
	// 
	// http_server:
	//   addr: ":8082"
	//---------------------------------------------------------------------------
	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}



	//---------------------------------------------------------------------------
	// STEP 5 → Create a channel to receive OS shutdown signals
	//
	// make(chan os.Signal, 1)
	//   - Buffer size 1 means channel can hold 1 signal
	//
	// This channel listens for:
	//   - CTRL + C  → os.Interrupt
	//   - SIGINT    → syscall.SIGINT  (interrupt)
	//   - SIGTERM   → syscall.SIGTERM (termination)
	//---------------------------------------------------------------------------
	done := make(chan os.Signal, 1)



	//---------------------------------------------------------------------------
	// STEP 6 → Register signals to be caught by this channel
	//
	// signal.Notify listens for OS signals and sends them into `done` channel.
	//
	// Meaning:
	//   When user presses CTRL+C → a signal is sent to `done`
	//   When OS sends shutdown → signal goes into `done`
	//---------------------------------------------------------------------------
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)



	//---------------------------------------------------------------------------
	// STEP 7 → Run HTTP server in a separate goroutine
	//
	// WHY A GOROUTINE?
	//   Because ListenAndServe is a BLOCKING call.
	//   If we run it in main goroutine, we can never receive shutdown signals.
	//
	// Anonymous goroutine:
	//   go func() { ... }()
	//---------------------------------------------------------------------------
	go func() {

		fmt.Println("server started on", cfg.HTTPServer.Addr)

		// server.ListenAndServe starts serving HTTP requests.
		// It returns an error only when server stops.
		err := server.ListenAndServe()

		// If server stops due to shutdown:
		//   http.ErrServerClosed → normal shutdown
		// If any other error:
		//   real failure
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.String("error", err.Error()))
		}
	}()



	//---------------------------------------------------------------------------
	// STEP 8 → Block main goroutine until shutdown signal received
	//
	// <-done : this waits until something is sent to the channel.
	// Once CTRL+C is pressed, we continue execution (shutdown begins).
	//---------------------------------------------------------------------------
	<-done



	//---------------------------------------------------------------------------
	// STEP 9 → Log shutdown initiation
	//---------------------------------------------------------------------------
	slog.Info("shutting down the server")



	//---------------------------------------------------------------------------
	// STEP 10 → Create context with timeout for graceful shutdown
	//
	// context.WithTimeout:
	//   - Allows ongoing requests to finish within N seconds
	//   - If timeout expires → force shutdown
	//
	// 5 * time.Second:
	//   Maximum wait duration for open connections to close cleanly.
	//---------------------------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()



	//---------------------------------------------------------------------------
	// STEP 11 → Gracefully shut down server
	//
	// server.Shutdown(ctx):
	//   ✔ stops accepting new requests
	//   ✔ finishes ongoing requests
	//   ✔ closes idle connections
	//   ✔ respects timeout
	//---------------------------------------------------------------------------
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", slog.String("error", err.Error()))
	}



	//---------------------------------------------------------------------------
	// STEP 12 → Confirm clean shutdown
	//---------------------------------------------------------------------------
	slog.Info("server shutdown successfully")
}
