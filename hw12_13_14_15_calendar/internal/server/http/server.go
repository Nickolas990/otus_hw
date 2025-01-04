package internalhttp

import (
	"context"
	"errors"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/interfaces"
	"log"
	"net/http"
	"time"
)

type Server struct {
	// TODO
	server *http.Server
	log    interfaces.Logger
	app    interfaces.Application
}

func NewServer(logger interfaces.Logger, app interfaces.Application, address string) *Server {
	return &Server{
		log: logger,
		app: app,
		server: &http.Server{
			Addr:    address,
			Handler: nil,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.server.Handler = s.newRouter() // Предполагается, что у вас есть метод newRouter
	log.Printf("Starting HTTP server on %s", s.server.Addr)

	// Запуск сервера в отдельной горутине
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Логирование ошибки, кроме случая закрытия сервера
			log.Printf("Error starting server: %v", err)
		}
	}()

	// Ожидание сигнала на завершение из контекста
	<-ctx.Done()

	// Когда контекст отменяется (ctx.Done() закрывается), выполняется остановка сервера
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		// Логирование ошибки при остановке сервера
		log.Printf("Error shutting down server: %v", err)
		return err
	}

	return nil
}

func (s *Server) newRouter() http.Handler {
	mux := http.NewServeMux()
	// Регистрация обработчиков
	mux.Handle("/hello", loggingMiddleware(http.HandlerFunc(s.handleEvent)))
	return mux
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	// Обработка запроса к эндпоинту /hello
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))

}

func (s *Server) Stop(ctx context.Context) error {

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	log.Println("Server stopped")
	return nil
}

// TODO
