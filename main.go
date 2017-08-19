package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"os/signal"
	"syscall"
	"context"
	"time"
	"strconv"
)

func envDefault(variable string, defaultVal string) int {
	delayVal := os.Getenv(variable)
	if (delayVal == "") {
		delayVal = defaultVal
	}

	delay, _ := strconv.Atoi(delayVal)
	return delay
}

type Server struct {
	logger *log.Logger
	mux    *http.ServeMux
	healthy bool
}

func NewServer(options ...func(*Server)) *Server {
	s := &Server{
		logger: log.New(os.Stdout, "", 0),
		mux:    http.NewServeMux(),
		healthy: true,
	}

	for _, f := range options {
		f(s)
	}

	s.mux.HandleFunc("/", s.index)
	s.mux.HandleFunc("/env", s.env)
	s.mux.HandleFunc("/fail", s.fail)
	s.mux.HandleFunc("/health", s.health)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("GET /")

	w.Write([]byte("<p style=\"font-size:96px\">Hello, World!</p>"))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("GET /")
	if (!s.healthy) {
		http.Error(w, "Not healthy", 500)
		return
	}
	w.Write([]byte("Healthy"))
}

func (s *Server) fail(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("GET /")
	s.healthy = false;
	w.Write([]byte("oh no!"))
}

func (s *Server) env(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("GET /env")

	w.Write([]byte(strings.Join(os.Environ(), "\n")))
}

func main() {
	stop := make(chan os.Signal, 1)


	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":80"
	}

	delay := envDefault("DELAY", "0")
	startupDelay := envDefault("STARTUP_DELAY", "0")

	time.Sleep(time.Duration(startupDelay)*time.Second)

	logger := log.New(os.Stdout, "", 0)
	s := NewServer(func(s *Server) { s.logger = logger })

	h := &http.Server{Addr: addr, Handler: s}

	go func() {
		logger.Printf("Listening on http://0.0.0.0%s\n", addr)

		if err := h.ListenAndServe(); err != nil {
			logger.Println(err)
		}
	}()

	sig := <-stop


	logger.Println("\nShutting down the server...")
	logger.Println(sig)

	h.Shutdown(context.Background())

	time.Sleep(time.Duration(delay)*time.Second)
	logger.Println("Server gracefully stopped")
}

//type server struct{}
//
//func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte(strings.Join(os.Environ(), "\n")))
//}
