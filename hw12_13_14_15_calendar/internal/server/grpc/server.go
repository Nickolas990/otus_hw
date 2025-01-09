package grpc

import (
	"context"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/api/grpc/pb"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/app"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/config"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/Nickolas990/otus_hw/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
)

// Server реализует gRPC сервис EventService.
type Server struct {
	app app.Application
	log logger.Logger
	pb.UnimplementedEventServiceServer
}

// NewServer возвращает новый экземпляр gRPC сервера.
func NewServer(log logger.Logger, app app.Application) *Server {
	return &Server{
		app: app,
		log: log,
	}
}

// Start RunServer инициализирует и запускает gRPC сервер.
func (s *Server) Start(ctx context.Context, config config.Config) error {
	lis, err := net.Listen("tcp", config.GRPC.Host+":"+config.GRPC.Port)
	if err != nil {
		s.log.Error("failed to listen:", err)
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterEventServiceServer(serv, NewServer(s.log, s.app))

	s.log.Info("server listening at", lis.Addr())
	if err := serv.Serve(lis); err != nil {
		s.log.Error("failed to serve:", err)
		return err
	}

	<-ctx.Done()

	s.log.Info("shutting down gRPC server...")
	serv.GracefulStop()

	s.log.Info("gRPC server stopped")

	return nil
}

func (s *Server) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateUpdateEventResponse, error) {
	grpcEvent := req.GetEvent()
	event, err := s.app.CreateEvent(ctx, storage.Event{
		Title:            grpcEvent.Title,
		Description:      grpcEvent.Description,
		StartTime:        grpcEvent.StartTime.AsTime(),
		EndTime:          grpcEvent.EndTime.AsTime(),
		UserID:           grpcEvent.UserId,
		NotificationTime: grpcEvent.NotificationTime.AsTime(),
	})

	if err != nil {
		// Вместо nil возвращаем статус ошибки через gRPC
		return nil, status.Errorf(codes.Internal, "error creating event: %v", err)
	}

	return &pb.CreateUpdateEventResponse{
		Event: &pb.Event{
			Id:               event.ID,
			Title:            event.Title,
			Description:      event.Description,
			StartTime:        timestamppb.New(event.StartTime),
			EndTime:          timestamppb.New(event.EndTime),
			UserId:           event.UserID,
			NotificationTime: timestamppb.New(event.NotificationTime),
		},
	}, nil
}
