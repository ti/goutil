package ip

import (
	"net"
	"net/http"
	"strings"
)

var noProxy bool

// GetIP returns IP address from request.
// Only when it used use proxy
func GetIP(r *http.Request) net.IP {
	if !noProxy {
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			parts := strings.Split(ip, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}
			return net.ParseIP(parts[0])
		}
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return net.ParseIP(ip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP(r.RemoteAddr)
	}
	return net.ParseIP(host)
}

func NoProxyMode()  {
	noProxy = true
}