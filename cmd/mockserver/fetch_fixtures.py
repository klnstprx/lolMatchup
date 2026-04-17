# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""
Fetch match fixtures from the Riot API for a given player.

Usage:
  uv run cmd/mockserver/fetch_fixtures.py <API_KEY> <gameName#tagLine> [--count N] [--region REGION]

Example:
  uv run cmd/mockserver/fetch_fixtures.py RGAPI-xxxx "eepy edward#1337" --count 25
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

FIXTURES = Path(__file__).parent / "fixtures"

REGION_TO_CLUSTER = {
    "na1": "americas", "br1": "americas", "la1": "americas", "la2": "americas",
    "euw1": "europe", "eun1": "europe", "ru": "europe", "tr1": "europe",
    "kr": "asia", "jp1": "asia",
    "oc1": "sea", "sg2": "sea", "tw2": "sea", "vn2": "sea",
}


def api_get(url: str, api_key: str):
    req = urllib.request.Request(url, headers={
        "X-Riot-Token": api_key,
        "User-Agent": "lolMatchup-fixture-fetcher/1.0",
    })
    try:
        with urllib.request.urlopen(req) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        print(f"HTTP {e.code} for {url}: {e.read().decode()}", file=sys.stderr)
        sys.exit(1)


def load_json(name: str) -> dict:
    path = FIXTURES / name
    if path.exists():
        with open(path, encoding="utf-8") as f:
            return json.load(f)
    return {}


def save_json(name: str, data):
    with open(FIXTURES / name, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)
        f.write("\n")


def main():
    parser = argparse.ArgumentParser(description="Fetch Riot API fixtures for a player")
    parser.add_argument("api_key", help="Riot API key")
    parser.add_argument("riot_id", help="Player Riot ID (gameName#tagLine)")
    parser.add_argument("--count", type=int, default=25, help="Number of matches to fetch (default: 25)")
    parser.add_argument("--region", default="euw1", help="Riot region (default: euw1)")
    args = parser.parse_args()

    parts = args.riot_id.split("#", 1)
    if len(parts) != 2:
        print("Error: riot_id must be in format gameName#tagLine", file=sys.stderr)
        sys.exit(1)
    game_name, tag_line = parts
    cluster = REGION_TO_CLUSTER.get(args.region, args.region)
    riot_base = f"https://{cluster}.api.riotgames.com"
    region_base = f"https://{args.region}.api.riotgames.com"

    # 1. Account
    acct = api_get(
        f"{riot_base}/riot/account/v1/accounts/by-riot-id/{urllib.request.quote(game_name)}/{urllib.request.quote(tag_line)}",
        args.api_key,
    )
    puuid = acct["puuid"]
    print(f"Account: {acct['gameName']}#{acct['tagLine']}  puuid={puuid[:20]}...")

    # 2. Summoner
    summoner = api_get(f"{region_base}/lol/summoner/v4/summoners/by-puuid/{puuid}", args.api_key)
    print(f"Summoner level: {summoner['summonerLevel']}")

    # 3. League entries
    league = api_get(f"{region_base}/lol/league/v4/entries/by-puuid/{puuid}", args.api_key)
    print(f"League entries: {len(league)}")

    # 4. Match IDs
    match_ids = api_get(
        f"{riot_base}/lol/match/v5/matches/by-puuid/{puuid}/ids?count={args.count}",
        args.api_key,
    )
    print(f"Match IDs fetched: {len(match_ids)}")

    # 5. Save account
    accounts = load_json("accounts.json")
    accounts[f"{acct['gameName']}#{acct['tagLine']}".lower()] = acct
    save_json("accounts.json", accounts)
    print("Updated accounts.json")

    # 6. Save summoner
    summoners = load_json("summoners.json")
    summoners[puuid] = summoner
    save_json("summoners.json", summoners)
    print("Updated summoners.json")

    # 7. Save league entries
    league_data = load_json("league_entries.json")
    league_data[puuid] = league
    save_json("league_entries.json", league_data)
    print("Updated league_entries.json")

    # 8. Save match IDs
    mid_data = load_json("match_ids.json")
    mid_data[puuid] = match_ids
    save_json("match_ids.json", mid_data)
    print(f"Updated match_ids.json with {len(match_ids)} IDs")

    # 9. Fetch each match
    matches_dir = FIXTURES / "matches"
    matches_dir.mkdir(exist_ok=True)
    for i, mid in enumerate(match_ids):
        fpath = matches_dir / f"{mid}.json"
        if fpath.exists():
            print(f"  [{i+1}/{len(match_ids)}] {mid} — already exists, skipping")
            continue
        print(f"  [{i+1}/{len(match_ids)}] Fetching {mid}...")
        match = api_get(f"{riot_base}/lol/match/v5/matches/{mid}", args.api_key)
        with open(fpath, "w", encoding="utf-8") as f:
            json.dump(match, f, indent=2)
            f.write("\n")
        time.sleep(1.3)  # respect rate limits

    print(f"\nDone! {len(match_ids)} matches available in fixtures.")


if __name__ == "__main__":
    main()
