package math

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// Port returns the address the HTTP server should listen on, in the form ":PORT".
// When the PORT env var is set, it takes precedence; otherwise a random high
// port is picked.
func Port() string {
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		return p
	}
	porta := rand.Intn(16383) + 49152
	return fmt.Sprintf(":%d", porta)
}
