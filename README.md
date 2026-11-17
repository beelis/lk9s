# lk9s

A k9s-style terminal UI for [LiveKit](https://livekit.io).

> This is a fork of [beelis/lk9s](https://github.com/beelis/lk9s) with its own releases. See
> [Differences from upstream](#differences-from-upstream).

## Installation

```bash
go install github.com/mariusgassen/lk9s/cmd/lk9s@latest
```

Or build from source:

```bash
git clone https://github.com/mariusgassen/lk9s.git
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
RoomRecord, IngressAdmin). Because the minted token can itself mutate the
server, grants for RoomAdmin, RoomCreate, RoomRecord and IngressAdmin are
refused on read-only contexts; join, publish and RoomList grants are not.

`~/.lk9s.yaml` holds API secrets, so keep it private (`chmod 600 ~/.lk9s.yaml`);
lk9s warns on startup if other users can read it. Context names must be unique.

## Usage

```
lk9s                              # interactive context selection
lk9s -context prod                # connect directly
lk9s -debug-log /tmp/lk9s.log     # also log room names and API errors to a file
```

### Commands

Press `:` anywhere to open the command bar (Tab/arrows autocomplete):

| Command                | Action                                    |
| ---------------------- | ----------------------------------------- |
| `:projects` / `:ctx`   | switch context without restarting         |
| `:rooms`               | back to the room list                     |
| `:create-room`         | create a room (write contexts only)       |
| `:token`               | generate an access token                  |
| `:sip`                 | list SIP trunks and dispatch rules        |
| `:quit` / `:q`         | exit                                      |

### Keys

Each view shows its keys in the legend at the bottom. Common ones: `Enter`
open, `Esc` back, `/` filter, `Shift+<column letter>` sort, `Ctrl+D` delete
room or kick participant (type-to-confirm), `Ctrl+E` full error text.

## Differences from upstream

Compared with [beelis/lk9s](https://github.com/beelis/lk9s), this fork adds:

- **Read-only by default.** Mutating actions need `write: true` on the context; a `[WRITE]` tag marks it in the header.
  Token grants that can mutate the server (RoomAdmin, RoomCreate, RoomRecord, IngressAdmin) are refused on read-only
  contexts too.
- **Type-to-confirm** for deleting a room and kicking a participant; Esc aborts a request that is already in flight.
- **More actions:** create room, mute/unmute a track, and local access-token minting with configurable grants and expiry
  (`T` or `:token`).
- **More views:** published track details (codecs, simulcast layers), room info, agent dispatches, SIP trunks and
  dispatch rules, a synthesized participant activity log (`v`), and full error text (`Ctrl+E`).
- **Navigation:** a k9s-style `:` command bar (`:projects`, `:ctx`, `:rooms`, `:create-room`, `:token`, `:sip`, `:q`),
  project switching without a restart, and `/` filtering on every table.
- **Config checks:** duplicate context names are rejected, and a config file readable by other users triggers a warning.
- **CI:** lint, `go vet` and `go test -race` on every PR, and Dependabot for Go modules and Actions.

## Releases

Releases are cut with [release-please](https://github.com/googleapis/release-please), independently of upstream.
Commits on `main` follow [Conventional Commits](https://www.conventionalcommits.org) (`feat:`, `fix:`, `docs:`, ...);
release-please keeps a release PR open that updates `CHANGELOG.md` and the version, and merging it creates the tag and
GitHub release. PRs are squash-merged, so the **PR title** becomes the commit message and must follow the same
format (checked by CI). Versions stay below 1.0.0 for now, so `feat:` bumps the minor and `fix:` the patch.
