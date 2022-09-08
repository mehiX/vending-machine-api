package app

import (
	"testing"
	"time"
)

func TestAppHttpServer(t *testing.T) {

	addr := "127.0.0.1:8989"

	srvr := NewApp(addr, nil).HttpServer()

	if srvr.Addr != addr {
		t.Errorf("wrong address for http. Expected: %s, got: %s", addr, srvr.Addr)
	}

	if srvr.Handler == nil {
		t.Error("routes not setup correctly")
	}

	if srvr.ReadHeaderTimeout != 3*time.Second {
		t.Errorf("wrong value for ReadHeaderTimeout. Expected: %s, got: %s", 3*time.Second, srvr.ReadHeaderTimeout)
	}

	if srvr.ReadTimeout != 30*time.Second {
		t.Errorf("wrong value for ReadTimeout. Expected: %s, got: %s", 30*time.Second, srvr.ReadTimeout)
	}

	if srvr.WriteTimeout != 30*time.Second {
		t.Errorf("wrong value for WriteTimeout. Expected: %s, got: %s", 30*time.Second, srvr.WriteTimeout)
	}

	if srvr.IdleTimeout != 60*time.Second {
		t.Errorf("wrong value for IdleTimeout. Expected: %s, got: %s", 60*time.Second, srvr.IdleTimeout)
	}
}
