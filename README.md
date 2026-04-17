# LoLMatchup

A web application built with Go for retrieving and displaying detailed League of Legends champion information, player lookups, and live game spectator data.

![GitHub last commit](https://img.shields.io/github/last-commit/klnstprx/lolMatchup)
![GitHub issues](https://img.shields.io/github/issues/klnstprx/lolMatchup)
![GitHub license](https://img.shields.io/github/license/klnstprx/lolMatchup)

## Features

- **Champion Lookup** with fuzzy search and autocomplete (Meraki Analytics API)
- **Player Lookup** by Riot ID — ranked tier/LP, champion pool summary, win/loss sparkline, match history
- **Live Game Spectator** with opponent enrichment: threat-level scoring, OTP detection, streak tracking, off-role detection
- **Content-Negotiated Routes** — same URL serves HTMX fragments or full pages depending on request type
- **Server-Side Rendering** with [templ](https://templ.guide/) + [htmx](https://htmx.org/) + Tailwind CSS
- **Persistent Cache** with automatic patch-version invalidation

## Project Structure

```
lolMatchup/
├── main.go                  # Entrypoint, server setup, graceful shutdown
├── config/                  # TOML-based configuration
├── router/                  # Gin router setup
├── handlers/                # HTTP request handlers
│   ├── champion.go          # Champion search (fragment + full page)
│   ├── player.go            # Player lookup (fragment + full page)
│   ├── livegame.go          # Live game spectator & opponent enrichment
│   ├── match.go             # Match detail & player stats modal
│   ├── autocomplete.go      # Fuzzy search suggestions
│   └── page_handlers.go     # Home page & unified search routing
├── components/              # Templ templates (*.templ)
├── client/                  # Riot & Meraki API client
├── cache/                   # In-memory + persistent cache with fuzzy search
├── models/                  # Domain models (champion, match, league, spectator)
├── data/                    # Data initialization & patch checking
├── middleware/              # Logging, recovery, rate limiting, cache headers
├── renderer/                # Custom Gin renderer for templ
├── static/                  # Embedded static assets (htmx)
└── cmd/mockserver/          # Flask mock server for local development
```

## Getting Started

### Prerequisites

- **Go** 1.26+
- **templ** CLI (`go install github.com/a-h/templ/cmd/templ@latest`)
- **Git**

### Installation

```bash
git clone https://github.com/klnstprx/lolMatchup.git
cd lolMatchup
go mod download
```

### Build & Run

```bash
make templ        # Generate Go from templ templates
make build        # Build binary (lolmatchup.bin)
./lolmatchup.bin
```

Or for development with live reload:

```bash
air
```

The server starts at `http://localhost:1337` by default.

### Mock Server (Development)

For local development without a Riot API key, use the included mock server:

```bash
uv run cmd/mockserver/server.py
```

Then set in `config.toml`:

```toml
riot_api_base_url = "http://localhost:9090"
```

The mock server serves fixture data for all Riot API endpoints (Account, Summoner, League, Spectator, Match).

## Configuration

Copy the example config and adjust values:

```bash
cp config.toml.example config.toml
```

Key fields:

| Field | Description | Default |
|-------|-------------|---------|
| `listen_addr` | Server host/IP | `127.0.0.1` |
| `port` | Server port | `1337` |
| `meraki_url` | Meraki Analytics CDN base URL | `https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/` |
| `ddragon_version_url` | DDragon versions endpoint (patch detection) | `https://ddragon.leagueoflegends.com/api/versions.json` |
| `debug` | Enable debug logging | `true` |
| `cache_path` | Local cache file path | `cache.json` |
| `riot_api_key` | Riot Games API key (for player/live game features) | — |
| `riot_region` | Regional routing (e.g. `na1`, `euw1`, `kr`) | `na1` |

> **Note**: Champion search works without a Riot API key. Player lookup and live game features require a valid key from the [Riot Developer Portal](https://developer.riotgames.com/).

## Routes

All entity routes are content-negotiated: HTMX requests receive a component fragment, direct browser navigation receives a full page with layout.

| Route | Description |
|-------|-------------|
| `/` | Home page with unified search |
| `/champion?champion=X` | Champion lookup |
| `/player?riotID=X` | Player profile (ranked, champion pool, match history) |
| `/livegame?riotID=X` | Live game spectator with opponent analysis |
| `/search?q=X` | Unified search router (redirects or proxies) |

## Testing

```bash
make test         # Run tests with coverage
make lint         # Format and vet
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
