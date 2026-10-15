package pg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5/pgconn"
)

type Pool struct {
	// Connection pool to the postgres database
	connn map[int]*Connection
}

type Bouncer struct {
	DB_URL          string // Database URL
	MAX_CONNECTIONS int    // Maximum number of connections to the postgres database
	MAX_REQUESTS    int    // Maximum number of requests ingested by the bouncer
	pool            *Pool  // Connection pool to the postgres database
	lastUsed        int    // Last used connection id
}

func NewPostgresBouncer(dsn string, connections int, processes int) (*Bouncer, error) {
	return &Bouncer{
		DB_URL:          dsn,
		MAX_CONNECTIONS: processes,
		MAX_REQUESTS:    connections,
		lastUsed:        0,
		pool:            &Pool{connn: make(map[int]*Connection)},
	}, nil
}

func (b *Bouncer) AddConnection(ctx context.Context, id int, dsn string) error {
	if len(b.pool.connn) >= b.MAX_CONNECTIONS {
		return fmt.Errorf("maximum number of connections reached")
	}
	if _, exists := b.pool.connn[id]; exists {
		return fmt.Errorf("connection with id already exists %d", id)
	}
	conn, err := NewConnection(ctx, dsn)
	if err != nil {
		return err
	}
	b.pool.connn[id] = conn
	slog.Info("Connection Added", "id", id)
	return nil
}
func (b *Bouncer) Serve(context context.Context) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		return
	}
	// Ensure the listener closes when the application exits
	defer listener.Close()

	slog.Info("TCP Server is running on port :8080...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue // Keep running to accept other connections
		}

		log.Printf("New client connected: %s", conn.RemoteAddr().String())
		go b.handleConnection(context, conn)
	}
}
func (b *Bouncer) routeQueryRR(context context.Context, query string) (pgconn.CommandTag, error) {
	slog.Info("Executing Query on Connection", "id", b.lastUsed+1%b.MAX_CONNECTIONS)
	conn := b.pool.connn[b.lastUsed%b.MAX_CONNECTIONS]

	tag, err := conn.pgxConnection.Exec(context, query)
	b.lastUsed = b.lastUsed + 1
	if err != nil {
		slog.Error("Failed to execute query: ", "error", err)
		return pgconn.CommandTag{}, err
	}
	slog.Info("ADDing to the lastUsed")

	return tag, nil
}

func (b *Bouncer) handleConnection(context context.Context, conn net.Conn) {
	defer conn.Close()
	log.Printf("Handling connection from %s", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnected gracefully: %s", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading data: %v", err)
			}
			return // Exit the loop and trigger deferred close
		}
		log.Printf("Received message: %s", message)

		// routeing to the connection pool happens below
		tag, err := b.routeQueryRR(context, message)
		if err != nil {
			_, writeErr := conn.Write([]byte(err.Error()))
			if writeErr != nil {
				log.Printf("Failed to write error to client: %v", writeErr)
			}
			return
		}

		response := fmt.Sprintf("Success: %s\n", tag)
		conn.Write([]byte(response))
	}
}
