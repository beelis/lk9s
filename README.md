# lk9s

A k9s-style terminal UI for [LiveKit](https://livekit.io).

## Installation

```bash
go install github.com/beelis/lk9s/cmd/lk9s@latest
```

Or build from source:

```bash
git clone https://github.com/beelis/lk9s.git
cd lk9s
make build
# binary written to bin/lk9s
```

## Configuration

Create `~/.lk9s.yaml` with one or more contexts:

```yaml
contexts:
  - name: dev
    url: https://dev.livekit.example.com
    api-key: devkey
    api-secret: devsecret
    write: true # allow destructive/mutating actions against this context
  - name: prod
    url: https://prod.livekit.example.com
    api-key: prodkey
    api-secret: prodsecret
    # write defaults to false: prod stays read-only unless set explicitly
```

Contexts are read-only by default. Actions that mutate server state (delete
room, kick participant, mute/unmute a track, edit participant permissions,
create a room) are refused with a status-bar message unless `write: true` is
set for that context. A `[WRITE]` tag next to the context name flags when
it's enabled. Generating an access token (`T` on the rooms/participants
views, or the `:token` command) is unaffected, since it's a local JWT
signature and never mutates the server. The token form covers both
per-room join/publish grants and project-wide ones (RoomCreate, RoomList,
RoomRecord, IngressAdmin).

## Usage

```
lk9s                   # interactive context selection
lk9s -context prod     # connect directly
```

## Planned features

- [x] Search/filter rows by typing (`/` on rooms, participants and tracks)
- [x] Status bar with counts and last-refresh time
- [x] Switch context without restarting (`:projects` command)
