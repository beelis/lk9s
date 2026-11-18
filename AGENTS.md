# lk9s

## What this repository is

lk9s is a k9s-style terminal UI for [LiveKit](https://livekit.io). It talks to the LiveKit server API (Twirp, via
`server-sdk-go/v2`) with the API key/secret of a configured context, and lists rooms, participants, tracks, egresses,
agent dispatches and SIP config, with a few write-gated actions (delete room, kick, mute, edit permissions, create room)
and local access-token minting. Go 1.26, tview/tcell.

## Module Structure

- `cmd/lk9s`: entry point; flags (`-context`, `-debug-log`), context resolution, debug logger
- `internal/config`: loads and validates `~/.lk9s.yaml` (contexts, per-context `write` flag, permission warning)
- `internal/lk`: `Client`, a thin wrapper over the LiveKit SDK that maps protobuf types to plain structs
- `internal/ui`: all tview views; `app.go` wires the page stack, the `:` command bar and context switching

## Architecture & Key Patterns

- The UI depends only on the `ui.RoomLister` interface (implemented by `*lk.Client`)
- `nav` (in `app.go`) is passed by value to every page: app, pages, client, context name, `write`, and a `ctx` that is
  cancelled when the project is switched. The `active` nav in `Run` is kept current by `showRooms`; `:` commands must use
  it, never the startup `n`
- Views refresh on a `refreshInterval` ticker in a goroutine; all widget updates go through `app.QueueUpdateDraw`
- Tables use `tableState[T]` with `column[T]` (sort by Shift+letter, `/` substring filter)
- Contexts are read-only by default. Gate every mutating action with `requireWrite`; token grants that mutate server
  state (`TokenGrant.RequiresWrite`) are also refused on read-only contexts
- Destructive actions use `confirmPage` (type the name back; Esc aborts via context cancellation and reports no error)
- Pages are overlays on `pages`; close them with `RemovePage`, not `SwitchToPage`

## Commands

- `make test` (`go test -race -cover ./...`, includes the TUI e2e tests in `internal/ui/tui_e2e_test.go`: the real app on a tcell simulation screen with a fake client), `make e2e` (`-tags e2e` tests in `internal/lk` against an in-process fake API, or a real LiveKit `--dev` server when `LK9S_E2E_URL` is set, as in the CI `e2e` job), `make lint` (`golangci-lint run`), `make vet`, `make build`
- CI pins golangci-lint v2.15.0; older local versions reject `exhaustruct_v5` in `.golangci.yaml`
- `vendor/` and `.deps/` are git-ignored; with no internet run `go mod vendor` on the host
- Follow the comment density of the surrounding code; lint uses most linters (`funlen` 80 lines)
- Commit and PR titles use Conventional Commits (`feat:`, `fix:`, ...); PRs are squash-merged, so the PR title is what
  release-please reads (enforced by `.github/workflows/pr-title.yml`). Keep the `Assisted-by` trailer
