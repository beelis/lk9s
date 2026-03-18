# lk9s

A k9s-style terminal UI for [LiveKit](https://livekit.io).

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
