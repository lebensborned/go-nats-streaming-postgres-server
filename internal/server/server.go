package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/lebensborned/go-nats-streaming-postgres-server/internal/db"
)

type Server struct {
	router *mux.Router
	srv    http.Server
	repo   db.Repository
}

// NewServer returns a object of server
func NewServer( /* config */ repo db.Repository) *Server {
	server := &Server{
		router: mux.NewRouter(),
		repo:   repo,
	}
	return server
}

// Setup only one route in server
func (s *Server) configureRouter() {
	http.Handle("/", s.router)
	s.router.HandleFunc("/order/{id}", s.GetOrderById).Methods(http.MethodGet)
}

// Some stuff (router, init a cache)
func (s *Server) Start() error {
	s.configureRouter()
	s.srv.Handler = s.router
	s.srv.Addr = os.Getenv("SERVER_PORT")
	s.repo.Fill()
	go s.serve()
	log.Println("Server started at port", s.srv.Addr)
	return nil
}

// Listening http
func (s *Server) serve() {
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print("Listen error: ", err)
	}
}

// Graceful shutdown
func (s *Server) Close() error {
	err := s.repo.Close()
	if err != nil {
		log.Fatalf("Database Shutdown Failed: %+v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}
	log.Println("HTTP listener is closed")
	return nil
}
