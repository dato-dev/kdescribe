# kdescribe

Terminal highlighter for `kubectl describe` output.

`kdescribe` reads describe output from stdin, finds common Kubernetes failure
signals, and prints the most important findings first.

## Usage

```sh
kubectl describe pod <pod-name> | kdescribe
```

As a kubectl plugin:

```sh
kubectl describe pod <pod-name> | kubectl kdescribe
```

For local development:

```sh
kubectl describe pod <pod-name> | go run .
```

## Output formats

```sh
kubectl describe pod <pod-name> | kdescribe --output human
kubectl describe pod <pod-name> | kdescribe --output json
kubectl describe pod <pod-name> | kdescribe --output markdown
```

Useful flags:

- `--no-color` disables ANSI colors
- `--min-score 80` only shows high-signal findings
- `--exit-code` exits with status 1 when findings are present
- `--show-all` prints every finding instead of the top results

## Current MVP

- regex-based detection for common workload issues
- severity and score for every finding
- line number and original describe line for context
- simple cluster health risk: `LOW`, `MEDIUM`, or `HIGH`
- lightweight parsing for `Containers`, `Conditions`, and `Events`

Example signals:

- `OOMKilled`
- `CrashLoopBackOff`
- `ImagePullBackOff`
- `ErrImagePull`
- `FailedScheduling`
- `FailedMount`
- readiness/liveness probe failures

## Releasing

The release binary is named `kubectl-kdescribe` so krew can expose it as:

```sh
kubectl kdescribe
```

Build release artifacts with GoReleaser:

```sh
goreleaser release --snapshot --clean
```

The draft krew manifest lives in `krew/kdescribe.yaml`. Replace `sha256: TODO`
values with checksums from the release before submitting to the krew index.
