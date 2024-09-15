package sync

import (
	"bufio"
	"bytes"
	"net"
)

type peer struct {
	conn     net.Conn
	commands chan<- command
}

func (p *peer) ReadInput() {
	for {
		msg, err := bufio.NewReader(p.conn).ReadBytes('\n')
		if err != nil {
			return
		}
		msg = msg[:len(msg)-1] //removes \n

		args := bytes.Split(msg, []byte(" "))
		cmd := string(args[0])

		switch cmd {
		case "BLOOM_FILTER":
			p.commands <- command{
				id:   BLOOM_FILTER,
				peer: p,
				args: args[1:],
			}
		case "BUCKET_DIGESTS":
			p.commands <- command{
				id:   BUCKET_DIGESTS,
				peer: p,
				args: args[1:],
			}
		case "NON_MATCHING_BUCKETS":
			p.commands <- command{
				id:   NON_MATCHING_BUCKETS,
				peer: p,
				args: args[1:],
			}
		case "BUCKET_DIFFS":
			p.commands <- command{
				id:   BUCKET_DIFFS,
				peer: p,
				args: args[1:],
			}
		}
	}
}
