package peer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func GenerateSwarmKey() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalln("While trying to read random source:", err)
	}

	return fmt.Sprintf("/key/swarm/psk/1.0.0/\n/base16/\n" + hex.EncodeToString(key))
}
