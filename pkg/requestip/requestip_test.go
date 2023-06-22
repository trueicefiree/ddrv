package requestip

import (
	"net/http"
	"testing"
)

func TestWithXRealIPHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.0.100")
	ip := Extract(req)
	if ip != "192.168.0.100" {
		t.Errorf("Expected IP: 192.168.0.100, Got: %s", ip)
	}
}

func TestXForwardedForHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.0.200, 10.0.0.1")
	ip := Extract(req)
	if ip != "192.168.0.200" {
		t.Errorf("Expected IP: 192.168.0.200, Got: %s", ip)
	}
}

func TestRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.0.300:12345"
	ip := Extract(req)
	if ip != "192.168.0.300" {
		t.Errorf("Expected IP: 192.168.0.300, Got: %s", ip)
	}
}

func TestHeadersOrRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ip := Extract(req)
	if ip != "" {
		t.Errorf("Expected empty IP, Got: %s", ip)
	}
}
