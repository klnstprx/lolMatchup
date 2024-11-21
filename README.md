![Project Status](https://img.shields.io/badge/status-WIP-orange)

> **Note**: This project is currently **not functional** and is a **Work in Progress (WIP)**.

Currently it is possible to:

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
- **Server-Side Rendering**: Utilizes `templ` for efficient server-side HTML rendering.

## Table of Contents

- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Project Structure

```
├── LICENSE
├── README.md
├── components
│   ├── champion.templ       # Templ file for champion component
│   ├── champion_templ.go    # Generated Go code from champion.templ
│   ├── home.templ           # Templ file for home component
│   └── home_templ.go        # Generated Go code from home.templ
├── config
│   ├── config.go            # Application configuration logic
├── config.toml              # Application configuration file
├── go.mod                   # Go module dependencies
├── go.sum                   # Checksums for module dependencies
├── handlers
│   └── champion.go          # HTTP handlers for champion routes
├── main.go                  # Entry point of the application
├── middleware
│   ├── logger.go            # Middleware for logging requests
│   └── recovery.go          # Middleware for recovering from panics
├── models
│   └── champion.go          # Data models for champions
├── renderer
│   └── renderer.go          # Custom renderer for templ templates
├── router
│   └── router.go            # Application routes setup
└── static
    └── htmx
        └── htmx.min.js      # htmx library for AJAX requests
```

## Getting Started

### Prerequisites

- **Go**: Version 1.16 or higher is recommended.
- **Git**: For cloning the repository.
- **Internet Connection**: Required to fetch data from Riot's Data Dragon API.

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/yourusername/lolMatchup.git
   cd lolMatchup
   ```

2. **Install Dependencies**

   Use `go mod` to download the necessary dependencies:

   ```bash
   go mod download
   ```

3. **Build the Application**

   ```bash
   go build
   ```

## Usage

1. **Configure the Application**

   Ensure the `config.toml` file exists in the root directory with the appropriate settings (see [Configuration](#configuration)).

2. **Run the Application**

   ```bash
   ./lolMatchup
   ```

   The server will start, listening on the port specified in your configuration (default is `1337`).

3. **Access the Web Interface**

   Open your web browser and navigate to:

   ```
   http://localhost:1337
   ```

   You should see the home page with a form to enter a champion's name.

4. **Retrieve Champion Data**

   - Enter the name of a League of Legends champion in the provided form.
   - Click **Submit**.
   - The application will fetch and display detailed information about the champion, including stats, abilities, and images.

## Configuration

The application uses a `config.toml` file for configuration settings. Here's an example of what your `config.toml` might look like:

```toml
listen_addr = ""
port = 1337
patch_number = "14.21.1"
language_code = "en_US"
ddragon_url = "https://ddragon.leagueoflegends.com/cdn/"
debug = true
```

- **listen_addr**: The address the server listens on (leave empty for localhost).
- **port**: The port number for the server (default `1337`).
- **patch_number**: The version of the game data to use.
- **language_code**: The language code for data localization (e.g., `en_US`).
- **ddragon_url**: Base URL for Riot's Data Dragon CDN.
- **debug**: Enables debug mode for more verbose logging.

### Logging Configuration

The application uses `charmbracelet/log` for logging. The logging level is determined by the `debug` setting in your `config.toml`:

- **debug = true**: Sets the logging level to `DebugLevel` for detailed logs.
- **debug = false**: Default logging level.

## Development

### Project Highlights

- **Go Modules**: Dependency management using Go modules (`go.mod`, `go.sum`).
- **Gin Web Framework**: Efficient HTTP routing and middleware support.
- **templ Templating Engine**: Generates efficient Go code from `.templ` files.
- **Custom Middleware**: For request logging and panic recovery.
- **Modular Design**: Organized code structure for scalability and maintainability.

### Key Components

- **main.go**: Initializes the application, loads configuration, and starts the server.
- **handlers/champion.go**: Contains the logic for handling champion data requests.
- **components/**: Holds the templating files (`.templ`) and their generated Go code.
- **models/champion.go**: Defines the data structures for champion data.
- **middleware/**: Custom middleware for logging and error handling.
- **router/router.go**: Sets up the application's HTTP routes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
