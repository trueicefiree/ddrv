package requestip

import (
    "net/http"
    "strings"
)

// Extract extracts the client IP address from the request
func Extract(r *http.Request) string {
    // Check if X-Real-IP header exists
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }

    // Check if X-Forwarded-For header exists
    if ips := r.Header.Get("X-Forwarded-For"); ips != "" {
        // Get the first IP from the comma-separated list
        ipsList := strings.Split(ips, ",")
        if len(ipsList) > 0 {
            return strings.TrimSpace(ipsList[0])
        }
    }

    // Fallback to RemoteAddr
    return strings.Split(r.RemoteAddr, ":")[0]
}
