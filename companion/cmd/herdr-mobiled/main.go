package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/engine"
)

func defaultSocket() string {
	if v := os.Getenv("HERDR_SOCKET_PATH"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

func defaultPushEndpointPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "mobile-push-endpoint")
}

func defaultPushRegistrationTokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "herdr", "mobile-push-registration-token")
}

// isNonLoopbackBind reports whether addr binds anything other than loopback,
// so we can warn (the v1 API is unauthenticated). Empty host and 0.0.0.0/::
// (all interfaces) count as non-loopback.
func isNonLoopbackBind(addr string) bool {
	host := addr
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		host = addr[:i]
	}
	host = strings.Trim(host, "[]")
	switch host {
	case "127.0.0.1", "::1", "localhost":
		return false
	default:
		return true
	}
}

// parseSpaces splits the --notify-spaces list, dropping empties so that an
// explicitly empty value means "every space" rather than "no space".
func parseSpaces(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	socket := flag.String("socket", defaultSocket(), "path to herdr.sock")
	// Default to loopback: v1 has NO API auth and the API can inject terminal
	// input, so it must not be world-reachable out of the box. Bind it to your
	// Tailscale IP explicitly (the systemd unit does this) to reach it from the
	// phone over the tailnet.
	listen := flag.String("listen", "127.0.0.1:8787", "WS listen address (bind to your tailnet IP, e.g. `tailscale ip -4`, to reach it from the phone)")
	poll := flag.Duration("poll", 1500*time.Millisecond, "pane.list poll interval")
	// Default to General only: OCA manager/worker spaces finish constantly and
	// their notifications drown the ones the owner actually acts on.
	notifySpaces := flag.String("notify-spaces", "General",
		"comma-separated Herdr space labels allowed to raise notifications; empty means all spaces")
	pushEndpointPath := flag.String("push-endpoint-file", defaultPushEndpointPath(),
		"private file used to preserve the registered UnifiedPush endpoint across restarts; empty disables persistence")
	pushRegistrationTokenPath := flag.String("push-registration-token-file", defaultPushRegistrationTokenPath(),
		"private bearer-token file for the current ChatKJB push registration endpoint; empty disables registration")
	pushEndpointOrigin := flag.String("push-endpoint-origin", "",
		"optional trusted UnifiedPush endpoint origin, for example http://100.64.0.1:2586")
	flag.Parse()

	e := engine.New(engine.Config{
		SocketPath:                *socket,
		ListenAddr:                *listen,
		PollInterval:              *poll,
		NotifySpaces:              parseSpaces(*notifySpaces),
		PushEndpointPath:          *pushEndpointPath,
		PushRegistrationTokenPath: *pushRegistrationTokenPath,
		PushEndpointOrigin:        *pushEndpointOrigin,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if isNonLoopbackBind(*listen) {
		log.Printf("WARNING: listening on %s — the v1 API has no authentication and can send input to your terminals. Only bind to a private (e.g. Tailscale) address, never a public one.", *listen)
	}
	log.Printf("herdr-mobiled: socket=%s listen=%s", *socket, *listen)
	if err := e.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
