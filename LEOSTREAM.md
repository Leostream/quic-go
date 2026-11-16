# Leostream fork of quic-go

Branch `leostream` is upstream quic-go plus one patch: pluggable congestion control.
Nothing else differs, so it can be rebased onto each upstream release.

- `congestion/`: the public `CongestionControl` interface (from apernet/quic-go, MIT, as Hysteria uses it).
- `monotime/`: the exported monotonic clock the interface uses.
- `Conn.SetCongestionControl(newCC)`: installs a controller from a constructor. Unlike
  apernet's version, the constructor is called again after a path migration, so the
  custom controller isn't silently replaced by quic-go's default (Reno).
- A custom controller keeps `ConnectionStats` (lost packets and bytes, congestion window)
  as quic-go's own sender does, and lost MTU probes don't reach it as congestion.

Consumers keep importing `github.com/quic-go/quic-go` and add:

```
replace github.com/quic-go/quic-go => github.com/Leostream/quic-go v0.65.1-leostream.2
```

## Updating to a new upstream release

```bash
git fetch upstream --tags
git rebase --onto vX.Y.Z vPREV leostream
go generate ./internal/mocks/... && go vet ./... && go test -race -short ./...
git tag vX.Y.(Z+1)-leostream.1 && git push origin leostream --tags
```
