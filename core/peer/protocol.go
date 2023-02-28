package peer

import (
	"bufio"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
)

const senderProtocol = "/libp2p/peer/send/1.0.0"

func streamHandler(s network.Stream) {
	log.Println("Got a new stream!")
	if err := doEcho(s); err != nil {
		log.Println(err)
		_ = s.Reset()
	} else {
		_ = s.Close()
	}
}

// doEcho reads a line of data from a stream and writes it back
func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}
