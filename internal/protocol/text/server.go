package text

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/JaneJavannie/in_memory_key_value_db/utils"
)

type TcpServer struct {
	address  string
	sem      chan struct{}
	listener net.Listener

	log       *slog.Logger
	onReceive func(ctx context.Context, request string) string

	wg sync.WaitGroup
}

func NewTcpServer(maxConnections int, address string, logger *slog.Logger) *TcpServer {
	// limit the number of connections
	connectionsCount := make(chan struct{}, maxConnections)
	for i := 0; i < maxConnections; i++ {
		connectionsCount <- struct{}{}
	}

	return &TcpServer{
		address: address,
		sem:     connectionsCount,
		log:     logger,
	}
}

func (s *TcpServer) SetOnReceive(onReceive func(ctx context.Context, request string) string) {
	s.onReceive = onReceive
}

func (s *TcpServer) Start() error {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.listener = listener
	s.log.Info("server started", "address", s.address)

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for {
			// Accept incoming connections
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				s.log.Warn("failed to accept connection", "error", err)
				continue
			}

			// Handle client connection in a goroutine
			go func() {
				err := s.handleClient(ctx, conn)
				if err != nil {
					s.log.Error("failed to handle client connection", "error", err)
				}
			}()
		}
	}()

	return nil
}

func (s *TcpServer) Stop() error {
	if s.listener == nil {
		return nil
	}

	err := s.listener.Close()
	s.wg.Wait()

	return err
}

func (s *TcpServer) handleClient(ctx context.Context, conn net.Conn) error {
	defer func() {
		s.log.Info("handle client", "msg", "connection close")

		conn.Close()
		s.sem <- struct{}{} // release connection
	}()

	<-s.sem

	slog.Info("accepted new connection")

	for {
		requestCtx := context.WithValue(ctx, consts.RequestID, utils.GetRequestUUID())
		l := s.log.With(consts.RequestID, requestCtx.Value(consts.RequestID))

		request, err := readWithContext(requestCtx, conn)
		if err != nil {
			return fmt.Errorf("failed to read data from client: %w", err)
		}

		l.Info("handle client: incoming request", "data", request)

		response := s.onReceive(requestCtx, request)

		err = writeWithContext(ctx, conn, response)
		if err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}
}
