# ChatKJB tailnet transport

ChatKJB's embedded Herdr client connects only to
`wss://neam-macmini.taild81d38.ts.net:8443`. Tailscale Serve terminates TLS
inside the tailnet and proxies to the relay bound to `127.0.0.1:8375`.
Funnel, Cloudflare, gateway, WebRTC, STUN, PCP, NAT-PMP, and UPnP are not part
of this deployment path.

The pairing setup fragment contains `relay=` and must never contain
`gateways=`. It is a credential: do not print, commit, or capture it.

## Rollback

The commands used here exist at `/bin/cp`, `/bin/launchctl`, and
`/usr/local/bin/tailscale`. `tailscale serve --help` was checked before this
procedure was recorded. The successful 8443 creation command also printed the
matching targeted removal command shown below.

1. Restore the protected pre-cutover relay environment:

   ```sh
   /bin/cp -p \
     /Users/neam/.config/herdr/plugins/config/herdr-mobile-relay.events/relay.env.pre-tailnet-20260903T0416.bak \
     /Users/neam/.config/herdr/plugins/config/herdr-mobile-relay.events/relay.env
   chmod 600 /Users/neam/.config/herdr/plugins/config/herdr-mobile-relay.events/relay.env
   ```

2. Restart only the ChatKJB relay LaunchAgent:

   ```sh
   /bin/launchctl bootout gui/$(id -u)/com.neamkim.chatkjb.herdr-relay
   /bin/launchctl bootstrap gui/$(id -u) \
     /Users/neam/Library/LaunchAgents/com.neamkim.chatkjb.herdr-relay.plist
   ```

3. Remove only the 8443 HTTPS handler:

   ```sh
   /usr/local/bin/tailscale serve --https=8443 off
   ```

4. Confirm in the machine-readable status that TCP 10110 still forwards to
   `127.0.0.1:10109`, TCP 8787 still forwards to `127.0.0.1:8788`, and no 8443
   web handler remains:

   ```sh
   /usr/local/bin/tailscale serve status --json
   ```

Never use `tailscale serve reset` or `tailscale serve clear` for this rollback.
They are whole-configuration operations and can remove the unrelated 10110 and
8787 mappings.

## Acceptance tests

- Case A: Wi-Fi off, mobile data on, Tailscale on — the embedded agent list and
  command path must continue to work.
- Case B: on any network, Tailscale off — the relay connection must fail.
- With the phone connected, the relay process may expose only loopback UDP
  8376 and loopback TCP 8375. It must have no connection to the community
  gateway IP addresses and no UDP 3478 STUN socket.

Cutover receipts are stored outside the repository at
`/Volumes/NEAM_SSD/.artifacts/2026-09-03/chatkjb-tailscale-cutover/`.
