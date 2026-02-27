package server

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Local struct {
	server *server.Server
	nc     *nats.Conn
	url    string
	host   string
	port   int

	tables map[string]*tableRuntime
}

func Open(host string, port int) (*Local, error) {
	natsServer, err := server.NewServer(&server.Options{
		ServerName: "hearts-embedded",
		Host:       host,
		Port:       port,
		NoSigs:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("create embedded nats server: %w", err)
	}

	go natsServer.Start()
	if !natsServer.ReadyForConnections(10 * time.Second) {
		natsServer.Shutdown()
		return nil, fmt.Errorf("embedded nats server did not become ready")
	}

	nc, err := nats.Connect(natsServer.ClientURL())
	if err != nil {
		natsServer.Shutdown()
		return nil, fmt.Errorf("connect embedded nats server: %w", err)
	}

	return &Local{
		server: natsServer,
		nc:     nc,
		url:    natsServer.ClientURL(),
		host:   host,
		port:   port,
		tables: make(map[string]*tableRuntime),
	}, nil
}

func (l *Local) URL() string {
	return l.url
}

func (l *Local) Host() string {
	return l.host
}

func (l *Local) Port() int {
	return l.port
}

func (l *Local) EnsureTable(tableID string) error {
	if tableID == "" {
		return fmt.Errorf("table id is required")
	}
	if _, exists := l.tables[tableID]; exists {
		return nil
	}

	table := newTableRuntime(tableID, l.nc)
	if err := table.register(); err != nil {
		table.shutdown()
		return fmt.Errorf("register table service: %w", err)
	}

	l.tables[tableID] = table
	return nil
}

func (l *Local) Close() {
	for _, table := range l.tables {
		table.shutdown()
	}
	l.nc.Close()
	l.server.Shutdown()
}
