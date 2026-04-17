# /// script
# requires-python = ">=3.11"
# dependencies = ["flask>=3.0"]
# ///
"""
Riot API Mock Server for lolMatchup development.

Serves fixture data for the Riot API endpoints used by the app:
  - Account v1:   GET /riot/account/v1/accounts/by-riot-id/<gameName>/<tagLine>
  - Summoner v4:  GET /lol/summoner/v4/summoners/by-puuid/<puuid>
  - League v4:    GET /lol/league/v4/entries/by-puuid/<puuid>
  - Spectator v5: GET /lol/spectator/v5/active-games/by-summoner/<puuid>
  - Match v5:     GET /lol/match/v5/matches/by-puuid/<puuid>/ids
  - Match v5:     GET /lol/match/v5/matches/<matchId>

Usage:
  uv run cmd/mockserver/server.py [--port PORT]

Then set in config.toml:
  riot_api_base_url = "http://localhost:9090"
"""

import argparse
import json
import sys
from pathlib import Path

from flask import Flask, jsonify, request

app = Flask(__name__)

FIXTURES_DIR = Path(__file__).parent / "fixtures"


def load_fixture(name: str) -> dict:
    path = FIXTURES_DIR / name
    if not path.exists():
        print(f"ERROR: fixture file not found: {path}", file=sys.stderr)
        return {}
    with open(path, encoding="utf-8") as f:
        return json.load(f)


def load_match_fixtures() -> dict:
    """Load individual match JSON files from fixtures/matches/ directory."""
    matches_dir = FIXTURES_DIR / "matches"
    matches = {}
    if matches_dir.exists():
        for f in matches_dir.glob("*.json"):
            match_id = f.stem
            with open(f, encoding="utf-8") as fh:
                matches[match_id] = json.load(fh)
    return matches


# Load fixtures at startup
ACCOUNTS: dict = load_fixture("accounts.json")
SUMMONERS: dict = load_fixture("summoners.json")
SPECTATOR: dict = load_fixture("spectator.json")
MATCH_IDS: dict = load_fixture("match_ids.json")
MATCHES: dict = load_match_fixtures()
LEAGUE_ENTRIES: dict = load_fixture("league_entries.json")


def riot_404(message: str = "Data not found"):
    return jsonify({
        "status": {
            "status_code": 404,
            "message": message,
        }
    }), 404


@app.route("/riot/account/v1/accounts/by-riot-id/<game_name>/<tag_line>")
def account_by_riot_id(game_name: str, tag_line: str):
    key = f"{game_name}#{tag_line}".lower()
    acct = ACCOUNTS.get(key)
    if acct is None:
        return riot_404(
            f"Data not found - No results found for player with riot id {game_name}#{tag_line}"
        )
    return jsonify(acct)


@app.route("/lol/summoner/v4/summoners/by-puuid/<puuid>")
def summoner_by_puuid(puuid: str):
    summoner = SUMMONERS.get(puuid)
    if summoner is None:
        return riot_404("Data not found - No results found for player")
    return jsonify(summoner)


@app.route("/lol/spectator/v5/active-games/by-summoner/<puuid>")
def spectator_by_puuid(puuid: str):
    game = SPECTATOR.get(puuid)
    if game is None:
        return jsonify({
            "httpStatus": 404,
            "errorCode": "NOT_FOUND",
            "message": "Not Found",
            "implementationDetails": "spectator game info isn't found",
        }), 404
    return jsonify(game)


@app.route("/lol/league/v4/entries/by-puuid/<puuid>")
def league_entries(puuid: str):
    entries = LEAGUE_ENTRIES.get(puuid)
    if entries is None:
        return jsonify([])  # Empty array = unranked
    return jsonify(entries)


@app.route("/lol/match/v5/matches/by-puuid/<puuid>/ids")
def match_ids_by_puuid(puuid: str):
    ids = MATCH_IDS.get(puuid)
    if ids is None:
        return jsonify([])
    count = request.args.get("count", default=20, type=int)
    return jsonify(ids[:count])


@app.route("/lol/match/v5/matches/<match_id>")
def match_by_id(match_id: str):
    match = MATCHES.get(match_id)
    if match is None:
        return riot_404(f"Data not found - match {match_id} not found")
    return jsonify(match)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Riot API Mock Server")
    parser.add_argument("--port", type=int, default=9090, help="Port to listen on")
    args = parser.parse_args()

    print(f"Mock Riot API server listening on http://localhost:{args.port}")
    print(f"Fixtures loaded: {len(ACCOUNTS)} accounts, {len(SUMMONERS)} summoners, {len(SPECTATOR)} spectator games, {len(MATCHES)} matches, {len(LEAGUE_ENTRIES)} league entries")
    print(f"\nSet in config.toml: riot_api_base_url = \"http://localhost:{args.port}\"")
    app.run(host="127.0.0.1", port=args.port, debug=False)
