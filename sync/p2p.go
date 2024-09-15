package sync

import (
	"log"
	"net"
	"time"
)

const PORT = "8080"

type p2p struct {
	peers    []*peer
	commands chan command
}

func (p *p2p) NewPeer(conn net.Conn) *peer {
	return &peer{
		conn:     conn,
		commands: p.commands,
	}
}

func (p *p2p) Sync() {
	syncTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-syncTicker.C:
			//TODO
		}
	}
}

func (p *p2p) ListenForConnections() {
	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}

	defer listener.Close()
	log.Printf("TCP Server started on :%s", PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err.Error())
			continue
		}

		peer := p.NewPeer(conn)
		go peer.ReadInput()
	}
}

func (p *p2p) HandleCommands() {
	for cmd := range p.commands {
		switch cmd.id {
		case BLOOM_FILTER:
			//TODO
		case BUCKET_DIGESTS:
			//TODO
		case NON_MATCHING_BUCKETS:
			//TODO
		case BUCKET_DIFFS:
			//TODO
		}
	}
}

func (p *p2p) Run() {
	go p.ListenForConnections()
	go p.Sync()
	p.HandleCommands()
}
