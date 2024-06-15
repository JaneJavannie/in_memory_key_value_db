package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/google/uuid"
)

type TcpServer struct {
	address  string
	sem      chan struct{}
	listener net.Listener
	db       *Database
	log      *slog.Logger

	wg sync.WaitGroup
}

func NewTcpServer(maxConnections int, address string, db *Database, logger *slog.Logger) *TcpServer {
	// limit the number of connections
	connectionsCount := make(chan struct{}, maxConnections)
	for i := 0; i < maxConnections; i++ {
		connectionsCount <- struct{}{}
	}

	return &TcpServer{
		address: address,
		sem:     connectionsCount,
		db:      db,
		log:     logger,
	}
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
			go s.handleClient(ctx, conn)
		}
	}()

	return nil
}

func (s *TcpServer) Stop() error {
	err := s.listener.Close()
	s.wg.Wait()

	return err
}

func (s *TcpServer) handleClient(ctx context.Context, conn net.Conn) {
	defer func() {
		s.log.Info("handle client", "msg", "connection close")

		conn.Close()
		s.sem <- struct{}{} // release connection
	}()

	<-s.sem

	slog.Info("accepted new connection")

	for {
		requestCtx := context.WithValue(ctx, consts.RequestID, uuid.New().String())
		l := s.log.With(consts.RequestID, requestCtx.Value(consts.RequestID))

		bytes, err := ReadBytesWithContext(ctx, conn)
		if err != nil {
			return
		}

		l.Info("handle client: incoming request", "data", string(bytes))

		// Process and use the data
		result, err := s.db.HandleRequest(requestCtx, string(bytes))
		if err != nil {
			l.Error("handle client: db: handle request", "error", err)
		}

		resp := fmt.Sprintf("query result: [ %s ] error: [ %v ] \n", result, err)

		_, err = conn.Write([]byte(resp))
		if err != nil {
			l.Error("handle client: write response result", "error", err)
			return
		}
	}
}

func ReadBytesWithContext(ctx context.Context, conn net.Conn) ([]byte, error) {
	type result struct {
		bytes []byte
		err   error
	}
	done := make(chan result, 1)

	go func() {
		connBuf := bufio.NewReader(conn)
		bytes, err := connBuf.ReadBytes('\n')
		done <- result{
			bytes: bytes,
			err:   err,
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		return r.bytes, r.err
	}
}
