# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

This is a Grafana datasource plugin for [EPICS Archiver Appliance](https://epicsarchiver.readthedocs.io). It has two independently built halves that ship together:

- **Backend** (Go, `pkg/`): compiled to a binary (`gpx_archiver-datasource-backend`) that Grafana runs as a subprocess. Handles query execution, data processing functions, and live-streaming proxying.
- **Frontend** (TypeScript/React, `src/`): the query editor, config editor, and (when the backend query path is disabled) a frontend-only query/streaming implementation. Built with the standard `@grafana/create-plugin` scaffolding under `.config/` (webpack, jest, tsconfig, eslint — don't edit configs there directly unless intentionally changing scaffolding).

Both halves implement equivalent query logic in parallel: whether a given install uses the Go backend or the browser-side client is controlled by the `useBackend` datasource setting (see `DataSource.query()` in [src/DataSource.ts](src/DataSource.ts)). Changes to query semantics (operators, functions, alias handling, extrapolation) often need to be made in both `pkg/` and `src/` to stay in sync.

## Common commands

### Frontend (Node — see `.nvmrc` for version, Yarn via Corepack)
```bash
yarn install          # install deps
yarn dev               # webpack watch build (development mode)
yarn build              # production build
yarn typecheck           # tsc --noEmit
yarn lint                # eslint (cached)
yarn lint:fix              # eslint --fix + prettier --write
yarn test                 # jest --watch --onlyChanged
yarn test:ci                # jest, non-interactive, used in CI
```
Run a single frontend test file: `yarn jest src/specs/aafunc.test.ts` (or point `jest` at any path/pattern; the `test`/`test:ci` scripts wrap the same jest config in [.config/jest.config.js](.config/jest.config.js)).

### Backend (Go, uses Mage as the build tool)
```bash
mage -v                       # build backend binaries for Linux/Windows/Darwin (Default target)
mage -l                       # list all available Mage targets
mage vtest                     # go test -v ./pkg/... (verbose, doesn't stop on first failure)
mage atest <TestName>            # go test -v ./pkg/... -run <TestName>  (single test by name)
go test ./pkg/...                 # plain go test across all backend packages
```

### End-to-end tests
```bash
yarn e2e                # cypress install + grafana-e2e run
yarn e2e:update           # same, but updates screenshots
```

### Documentation (Sphinx, in docs/)
```bash
python -m venv env
source env/bin/activate
pip install sphinx myst-parser sphinx_rtd_theme
cd docs && make html
```

### Local dev environment
`docker-compose.yaml` (via `yarn server` / `docker compose up --build`) spins up four containers: `grafana`, `mariadb`, `archappl` (Archiver Appliance), and `ioc` (a softIoc that auto-runs). No PVs are archived by default — register them manually. Add a datasource in the `grafana` container pointing at `http://archappl:8080/retrieval`. A `.devcontainer/` config is also available for VS Code Dev Containers.

CI (`.github/workflows/build-and-test.yml`) runs, in order: `yarn install --immutable`, `yarn typecheck`, `yarn lint`, `yarn test:ci`, `yarn build`, then Go tests via `mage vtest`, then a backend build via `mage -v buildAll` — mirror this sequence locally before pushing.

## Commit message conventions

Prefix every commit's subject line with one of NumPy's standard acronyms (see [NumPy's contributor guide](https://numpy.org/doc/stable/dev/development_workflow.html)), e.g. `ENH: add live-update reconnect backoff` or `BUG: fix alternation isolation for nested groups`:

| Acronym | Description |
| --- | --- |
| `API` | an (incompatible) API change |
| `BENCH` | changes to the benchmark suite |
| `BLD` | change related to building |
| `BUG` | bug fix |
| `DEP` | deprecate something, or remove a deprecated object |
| `DEV` | development tool or utility |
| `DOC` | documentation |
| `ENH` | enhancement |
| `MAINT` | maintenance commit (refactoring, typos, etc.) |
| `REL` | related to releasing |
| `REV` | revert an earlier commit |
| `STY` | style fix (whitespace, formatting, etc., no logic change) |
| `TST` | addition or modification of tests |
| `TYP` | static typing |

## Backend architecture (`pkg/`)

- [pkg/main.go](pkg/main.go) — entrypoint; calls `datasource.Manage` from grafana-plugin-sdk-go.
- [pkg/datasource.go](pkg/datasource.go) — `ArchiverDatasource` implements `QueryData`, `CheckHealth`, and the streaming (`SubscribeStream`/`RunStream`/`PublishStream`) interfaces. `QueryData` fans out per-target queries concurrently via `concurrent.QueryData` (max 10 in flight).
- [pkg/archiverappliance/query.go](pkg/archiverappliance/query.go) — core per-query pipeline: resolve target PV(s) (`makeTargetPVList`, handles regex expansion and `(A|B|C)` alternation isolation), fan out one goroutine per PV against a 30s timeout, apply alias (`applyAlias`), sort by name, apply processing functions (`functions.ApplyFunctions`), apply extrapolation (`dataExtrapol`), then convert each `SingleData` to a `data.Frame`. If live update is enabled, attaches a Grafana Live channel to the frame (`createLiveChannel`).
- [pkg/archiverappliance/aaclient.go](pkg/archiverappliance/aaclient.go), `pbparse.go`, `pvparser.go`, `operators.go` — HTTP client for the Archiver Appliance retrieval API; `pb/` contains generated protobuf types (`EPICSEvent.pb.go`) for parsing the AA binary protocol responses.
- [pkg/functions/](pkg/functions/) — the processing-function pipeline (sort, filter, offset, scale, alias transforms, etc.) applied to query results server-side; `functions_handler.go` dispatches by function name/params.
- [pkg/models/](pkg/models/) — shared types: `ArchiverQueryModel` (parsed query JSON, see [pkg/models/models.go](pkg/models/models.go)), `DatasourceSettings`, and the `SingleData`/`Scalars`/`Arrays`/`Strings` value containers used to represent one PV's timeseries regardless of its underlying EPICS type.
- [pkg/aalive/](pkg/aalive/) — WebSocket proxy to the Archiver Appliance's live-update (`pvws`) endpoint, used for Grafana Live streaming.

Backend tests are colocated as `*_test.go` next to the code they cover; `pkg/testhelper/` and `pkg/models/testhelpers.go` hold shared fixtures/helpers.

## Frontend architecture (`src/`)

- [src/DataSource.ts](src/DataSource.ts) — the `DataSource` class (extends `DataSourceWithBackend`). `query()` branches: if any target has `stream: true` and the time range's `to` is `now`, it runs `StreamQuery`; otherwise if `useBackend` is set it defers to the Go backend via `super.query()`, else it runs the frontend-only path (`doQuery` from [src/query.ts](src/query.ts)) directly against the Archiver Appliance HTTP API. `metricFindQuery` implements Grafana template-variable PV lookups (parses `PV:NAME:.*?limit=N` syntax via [src/pvnameParser.ts](src/pvnameParser.ts)).
- [src/aaclient.ts](src/aaclient.ts) — direct HTTP client to the Archiver Appliance retrieval API (frontend-only query path and PV name search/autocomplete).
- [src/query.ts](src/query.ts) / [src/dataProcessor.ts](src/dataProcessor.ts) / [src/responseParse.ts](src/responseParse.ts) — frontend-side equivalent of the backend's query/processing pipeline, used when `useBackend` is false.
- [src/streamQuery.ts](src/streamQuery.ts) — live-update streaming implementation for panels with `stream: true` and an open (`now`) time range.
- [src/aafunc.ts](src/aafunc.ts) — defines the catalog of processing functions available in the query editor UI (must stay in sync with `pkg/functions` function names/params on the backend).
- [src/components/](src/components/) — React UI: `QueryEditor.tsx` (main query editor), `ConfigEditor.tsx` (datasource settings page), `Function*.tsx` (the function-chain editor: add/edit/remove processing functions on a query).
- [src/types.ts](src/types.ts) — shared query/options type definitions (`AAQuery`, `AADataSourceOptions`, `TargetQuery`, etc.), mirrors the backend's `ArchiverQueryModel` JSON shape.
- [src/module.ts](src/module.ts) — plugin registration entrypoint (wires `DataSource`, `ConfigEditor`, `QueryEditor` into the `DataSourcePlugin`).
- Tests live in [src/specs/](src/specs/) as `*.test.ts`, run via Jest with jsdom.

## Key cross-cutting concepts

- **Query modes**: backend-executed vs. frontend-executed queries are two parallel implementations of the same semantics (target parsing, operators, alias, functions, extrapolation). When changing query behavior, check whether the equivalent Go (`pkg/archiverappliance`, `pkg/functions`) and TypeScript (`src/query.ts`, `src/dataProcessor.ts`, `src/aafunc.ts`) code needs matching updates.
- **PV regex/alternation**: the Archiver Appliance server only supports a limited regex dialect. Alternation groups like `(PV:1|PV:2)` are always isolated/expanded client-side before being sent, even when full regex mode is off — see `isolateBasicQuery`/`makeTargetPVList` in [pkg/archiverappliance/query.go](pkg/archiverappliance/query.go) and its frontend counterpart in [src/pvnameParser.ts](src/pvnameParser.ts).
- **Streaming/live update**: two distinct mechanisms — "Stream" mode (periodic re-query with a circular buffer, driven by `StreamQuery`/`strmInt`/`strmCap`) and "Live" mode (Grafana Live channel backed by a WebSocket proxy to the AA `pvws` endpoint in `pkg/aalive`). Don't conflate the two when reading query options (`stream`/`strmInt`/`strmCap` vs. `live`).
- **Processing functions**: applied server-side (`pkg/functions`) or client-side (`dataProcessor.ts`), configured through the function-chain UI in `src/components/Function*.tsx`. Function definitions/param schemas live in `src/aafunc.ts` (frontend) and must match the dispatch table in `pkg/functions/functions_handler.go` (backend) by name.
