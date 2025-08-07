--------------------------------------------------------------------------------

# README.md

--------------------------------------------------------------------------------
![Project Status](https://img.shields.io/badge/status-WIP-orange)

> **Note**: This project is currently **not functional** and is a **Work in Progress (WIP)**.

Currently, it is possible to:

- Get raw champion data.

---

# LoLMatchup

LoLMatchup is a web application built with Go that allows users to retrieve and display detailed information about League of Legends champions. It uses Riot Games' Data Dragon API to fetch champion data and presents it through a clean, interactive web interface.

![GitHub last commit](https://img.shields.io/github/last-commit/klnstprx/lolMatchup)
![GitHub issues](https://img.shields.io/github/issues/klnstprx/lolMatchup)
![GitHub license](https://img.shields.io/github/license/klnstprx/lolMatchup)

## Features

- **Champion Lookup**: Enter a champion's name to get detailed stats, abilities, and images.  
- **Dynamic Loading**: Uses htmx for seamless AJAX requests without full page reloads.  
- **Server-Side Rendering**: Utilizes "templ" for efficient server-side HTML rendering.

## Table of Contents

1. [Project Structure](#project-structure)  
2. [Getting Started](#getting-started)  
   1. [Prerequisites](#prerequisites)  
   2. [Installation](#installation)  
3. [Usage](#usage)  
4. [Configuration](#configuration)  
5. [Development](#development)  
6. [Testing](#testing)  
7. [Contributing](#contributing)  
8. [License](#license)

## Project Structure

Below is a high-level overview of the project structure:

```
├── LICENSE
├── README.md
├── .gitignore
├── components
│   ├── champion.templ
│   ├── champion_templ.go
│   ├── home.templ
│   └── home_templ.go
├── config
│   └── config.go
├── go.mod
├── go.sum
├── handlers
│   ├── champion.go
│   └── home.go
├── main.go
├── middleware
│   ├── logger.go
│   └── recovery.go
├── models
│   ├── champion.go
│   └── championList.go
├── renderer
│   └── renderer.go
├── router
│   └── router.go
├── static
│   └── htmx
│       └── htmx.min.js
└── ...
```

## Getting Started

### Prerequisites

- **Go** (1.16 or higher is recommended).  
- **Git** (for cloning the repository).  
- **Internet Connection** (to fetch data from Riot's Data Dragon API).

### Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/lolMatchup.git
   cd lolMatchup
   ```

2. **Install Dependencies**:

   ```bash
   go mod download
   ```

3. **Build the Application**:

   ```bash
   go build
   ```

## Usage

1. **Configure the Application**  
   Make sure the configuration file (config.toml) exists with the desired settings (see [Configuration](#configuration)).

2. **Run the Application**:

   ```bash
   ./lolMatchup
   ```

   The server will start listening on the port specified in your configuration (default is 1337).

3. **Access the Web Interface**  
   Open your web browser and go to:

   ```
   http://localhost:1337
   ```

   You should see the home page with a form to enter a champion's name.

4. **Retrieve Champion Data**  
   - Enter the name of a League of Legends champion.  
   - Click "Submit".  
   - The application will fetch and display champion information including stats, abilities, and images.

## Configuration

The application uses a TOML-based configuration file (config.toml) by default. An example configuration might look like:

```toml
listen_addr           = ""
port                  = 1337
language_code         = "en_US"
ddragon_url           = "https://ddragon.leagueoflegends.com/cdn/"
ddragon_version_url   = "https://ddragon.leagueoflegends.com/api/versions.json"
levenshtein_threshold = 3
debug                 = true
http_client_timeout   = 10
cache_path            = "cache.gob"
riot_api_key          = "YOUR_RIOT_API_KEY_HERE"
riot_region           = "na1"
```

Key fields:  
• listen_addr: The host/IP for the server. Leave empty for localhost.  
• port: The port on which the server listens (default is 1337).  
• language_code: Data localization language (e.g. "en_US").  
• ddragon_url: Riot’s Data Dragon CDN base URL.  
• ddragon_version_url: Endpoint for fetching available patch versions.  
• debug: Enables debug mode for more verbose logs.  
• http_client_timeout: Timeout for HTTP calls, in seconds.  
• cache_path: Where to store champion data cache.  
• riot_api_key: Riot Games API key for accessing live game and summoner endpoints.  
• riot_region: Regional routing value for Riot API (e.g. "na1", "euw1").

### Logging

The application uses charmbracelet/log for logging. If debug = true in config.toml, the log level is set to Debug. Otherwise, it defaults to Info level.

## Development

Highlights:  
• Go modules for dependency management.  
• Gin web framework for routing and middleware.  
• Templ engine for server-side HTML generation.  
• Custom middleware for request logging and panic recovery.  
• A structured, modular design to keep code maintainable and scalable.

## Testing

1. Run any Go-based tests using:  

   ```bash
   go test ./...
   ```  

2. (Optional) You can place integration or system tests in the "lolmatchup_testing" directory or any dedicated testing folder.

## Contributing

1. Fork the repository.  
2. Create a new branch: git checkout -b feature/myFeature.  
3. Commit your changes: git commit -m 'Add some feature'.  
4. Push the branch: git push origin feature/myFeature.  
5. Create a Pull Request on GitHub.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
