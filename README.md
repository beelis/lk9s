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
  - name: prod
    url: https://prod.livekit.example.com
    api-key: prodkey
    api-secret: prodsecret
```

## Usage

```
lk9s                   # interactive context selection
lk9s -context prod     # connect directly
```

## Planned features

- [ ] Search/filter rows by typing
- [ ] Status bar with counts and last-refresh time
- [ ] Switch context without restarting (`c` to reopen picker)
