# LoLMatchup

A web application built with Go for retrieving and displaying detailed League of Legends champion information, player lookups, and live game spectator data.

![GitHub last commit](https://img.shields.io/github/last-commit/klnstprx/lolMatchup)
![GitHub issues](https://img.shields.io/github/issues/klnstprx/lolMatchup)
![GitHub license](https://img.shields.io/github/license/klnstprx/lolMatchup)

## Features

- **Champion Lookup** with fuzzy search and autocomplete (Meraki Analytics API)
- **Player Lookup** by Riot ID via Summoner API
- **Live Game Spectator** view via Spectator v5 API
- **Server-Side Rendering** with [templ](https://templ.guide/) + [htmx](https://htmx.org/)
- **Persistent Cache** with automatic patch-version invalidation

## Project Structure

```
lolMatchup/
├── main.go                  # Entrypoint, server setup, graceful shutdown
├── config/                  # TOML-based configuration
├── router/                  # Gin router setup
├── handlers/                # HTTP request handlers
│   ├── champion.go          # Champion search
│   ├── player.go            # Player lookup
│   ├── livegame.go          # Live game spectator
│   ├── autocomplete.go      # Fuzzy search suggestions
│   └── page_handlers.go     # Page rendering
├── components/              # Templ templates (*.templ)
├── client/                  # Riot & Meraki API client
├── cache/                   # In-memory + persistent cache with fuzzy search
├── models/                  # Domain models
├── data/                    # Data initialization & patch checking
├── middleware/               # Request logging & panic recovery
├── renderer/                # Custom Gin renderer for templ
├── static/                  # Embedded static assets (htmx)
└── lolmatchup_testing/      # Bruno REST client collection
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

## Testing

```bash
make test         # Run tests with coverage
make lint         # Format and vet
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
