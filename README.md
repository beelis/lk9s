# lk9s

A k9s-style terminal UI for [LiveKit](https://livekit.io).

> lk9s is an independent community project and is not affiliated with, endorsed by, or sponsored by the official
> LiveKit project or LiveKit, Inc. LiveKit is a trademark of its respective owner.

lk9s started as a fork of [beelis/lk9s](https://github.com/beelis/lk9s) and is released independently. See
[Features](#features) and [Credits](#credits).

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

Contexts are read-only by default; set `write: true` to allow actions that change server state (see
[Write actions](#write-actions)). The header shows `[RW]` or `[RO]` next to the context name. The token form covers both
per-room join/publish grants and project-wide ones (RoomCreate, RoomList, RoomRecord, IngressAdmin).

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
| `:ingress`             | list ingresses                            |
| `:sip`                 | list SIP trunks and dispatch rules        |
| `:quit` / `:q`         | exit                                      |

### Keys

The header shows the keys of the current view. Common ones: `Enter` open, `Esc` back, `/` filter,
`Shift+<column letter>` sort, `Ctrl+D` delete or stop (type-to-confirm), `Ctrl+E` full error text.

| View       | Keys                                                                    |
| ---------- | ----------------------------------------------------------------------- |
| Rooms      | `u` update metadata, `d` send data, `e` egresses, `g` agents, `T` token |
| Egresses   | `Ctrl+D` stop egress                                                    |
| Agents     | `c` dispatch an agent, `Ctrl+D` delete a dispatch                       |

## Write actions

Everything that changes server state needs `write: true` on the context. On a read-only context these actions are
refused with a status-bar message and nothing is sent. Browsing, filtering, sorting and token minting always work.

| Action                  | Where                     | Key / command          | Confirmation             |
| ----------------------- | ------------------------- | ---------------------- | ------------------------ |
| Create a room           | any view                  | `:create-room`         | form: name, empty timeout, max participants |
| Delete a room           | rooms                     | `Ctrl+D`               | type the room name       |
| Update room metadata    | rooms                     | `u`                    | form                     |
| Send a data packet      | rooms                     | `d`                    | form: topic, payload, identities (empty = all), reliable or lossy |
| Kick a participant      | participants              | `Ctrl+D`               | type the identity        |
| Edit permissions        | participants              | `p`                    | form with the permission checkboxes; the publish sources are shown read-only |
| Mute / unmute a track   | tracks                    | `m`                    | none                     |
| Stop an egress          | egresses                  | `Ctrl+D`               | type the egress ID       |
| Dispatch an agent       | agents                    | `c`                    | form: agent name, metadata |
| Delete an agent dispatch | agents                   | `Ctrl+D`               | type the dispatch ID     |

Destructive actions are type-to-confirm. Esc aborts, and also cancels a request that is already in flight.

LiveKit servers refuse remote unmuting unless `enable_remote_unmute` is set to `true` in the server config, so unmuting
a track can fail on a default server.

Minting an access token (`T` or `:token`) is a local signature and never calls the server. Because the token itself can
mutate the server, grants for RoomAdmin, RoomCreate, RoomRecord and IngressAdmin are refused on read-only contexts.
Join, publish and RoomList grants are not.

## Features

- **Views:** rooms, participants, published tracks (codecs, simulcast layers), room info, egresses, ingress, agent
  dispatches, SIP trunks and dispatch rules, metadata and attributes, a synthesized participant activity log (`v`),
  and full error text (`Ctrl+E`).
- **Actions** (write contexts only): create and delete rooms, update room metadata, send data packets, kick
  participants, mute/unmute tracks, edit permissions, stop egresses, dispatch and delete agents. Destructive actions are
  type-to-confirm, and Esc aborts a request that is already in flight.
- **Access tokens:** local minting with configurable grants and expiry (`T` or `:token`).
- **Safety:** contexts are read-only by default. Token grants that can mutate the server (RoomAdmin, RoomCreate,
  RoomRecord, IngressAdmin) are refused on read-only contexts too. Duplicate context names are rejected, and a config
  file readable by other users triggers a warning.
- **Navigation:** a k9s-style header with context, URL, lk9s revision, server SDK and `lk` CLI version, the keys of the
  current view and the LiveKit logo; a `:` command bar, project switching without a restart, and `/` filtering and
  sorting on every table.
- **Quality:** lint, `go vet` and `go test -race` on every PR, and Dependabot for Go modules and Actions.

## Credits

lk9s is based on [beelis/lk9s](https://github.com/beelis/lk9s), and the interface follows [k9s](https://k9scli.io).
It uses the [LiveKit server SDK for Go](https://github.com/livekit/server-sdk-go).

## Releases

Releases are cut with [release-please](https://github.com/googleapis/release-please), independently of upstream.
Commits on `main` follow [Conventional Commits](https://www.conventionalcommits.org) (`feat:`, `fix:`, `docs:`, ...);
release-please keeps a release PR open that updates `CHANGELOG.md` and the version, and merging it creates the tag and
GitHub release. PRs are squash-merged, so the **PR title** becomes the commit message and must follow the same
format (checked by CI). Versions stay below 1.0.0 for now, so `feat:` bumps the minor and `fix:` the patch.
