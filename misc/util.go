package misc

import (
	"crypto/rand"
	"net"
	"net/http"
	"strings"

	"github.com/srhnsn/go-utils/log"
)

const RandomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GetProxiedIpAddress(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		log.Warning.Printf("GetProxiedIpAddress(): No X-Forwarded-For set, using real IP address (%s)", ip)
	}

	parts := strings.Split(ip, ",")
	ip = strings.TrimSpace(parts[0])

	return ip
}

func GetRandomString(length uint8) string {
	var bytes = make([]byte, length)
	alphabetLength := float64(len(RandomStringAlphabet) - 1)

	rand.Read(bytes)

	for i, b := range bytes {
		value := uint8((float64(b) / 255 * alphabetLength) + 0.5) // Poor man's Round()
		bytes[i] = RandomStringAlphabet[value]
	}

	return string(bytes)
}
