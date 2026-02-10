package rand

import (
	"crypto/rand"
	"math/big"
)

func NewReqID() string {
	// 64 bits aleatórios => ~13 chars em base36
	max := new(big.Int).Lsh(big.NewInt(1), 64)
	n, _ := rand.Int(rand.Reader, max)
	return n.Text(36)
}
