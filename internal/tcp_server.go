package internal

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"github.com/google/uuid"
)

type TcpServer struct {
	//maxConnections int
	address string

	sem      chan struct{}
	db       *Database
	log      *slog.Logger
	listener net.Listener
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

func (s *TcpServer) Start(ctx context.Context) error {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.listener = listener

	s.log.Info("server started", "address", s.address)

	go func() {
		for {
			if ctx.Err() != nil {
				return
			}

			// Accept incoming connections
			conn, err := listener.Accept()
			if err != nil {
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
	return s.listener.Close()
}

func (s *TcpServer) handleClient(ctx context.Context, conn net.Conn) {
	defer func() {
		s.log.Info("handle client", "connection close")

		conn.Close()
		s.sem <- struct{}{} // release connection
	}()

	<-s.sem

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
		<-ctx.Done()
		done <- result{
			bytes: nil,
			err:   ctx.Err(),
		}

	}()

	go func() {
		connBuf := bufio.NewReader(conn)
		bytes, err := connBuf.ReadBytes('\n')
		done <- result{
			bytes: bytes,
			err:   err,
		}
	}()

	r := <-done

	return r.bytes, r.err
}
