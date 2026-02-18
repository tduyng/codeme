package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(b))
}
