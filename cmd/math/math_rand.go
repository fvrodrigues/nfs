package math

import (
	"fmt"
	"math/rand"
)

func Port() string {
	porta := rand.Intn(16383) + 49152
	return fmt.Sprintf(":%d", porta)
}
