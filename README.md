# gbans

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Test, Build & Publish](https://github.com/leighmacdonald/gbans/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/leighmacdonald/gbans/actions/workflows/build.yml)
[![release](https://github.com/leighmacdonald/gbans/actions/workflows/release.yml/badge.svg?event=release)](https://github.com/leighmacdonald/gbans/actions/workflows/release.yml)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/f06234b0551a49cc8ac111d7b77827b2)](https://www.codacy.com/manual/leighmacdonald/gbans?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=leighmacdonald/gbans&amp;utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/leighmacdonald/gbans)](https://goreportcard.com/report/github.com/leighmacdonald/gbans)
[![GoDoc](https://godoc.org/github.com/leighmacdonald/gbans?status.svg)](https://pkg.go.dev/github.com/leighmacdonald/gbans)
![Lines of Code](https://tokei.rs/b1/github/leighmacdonald/gbans)
[![Discord chat](https://img.shields.io/discord/704508824320475218)](https://discord.gg/YEWed3wY3F)

gbans is a centralized game administration system for Team Fortress 2 that incorporates a global ban system similar to
[sourcebans++](https://sbpp.dev) as well as some additional basic community tools like news & wikis.
The project was initially only intended to be a more modern & secure replacement for sourcebans, but has evolved to
handle broader community aspects.

## Stability / Usage Notice

While we currently are [dogfooding](https://en.wikipedia.org/wiki/Eating_your_own_dog_food) the project on a
community with around 50 servers, I would not recommend non-developers use the project yet. It's still in fairly
major development mode and large sections are still incomplete or function but very rough. This is
very notable for the web frontend which we don't really use yet. Sticking with the discord command interface is the
best current way to interact with the system.

Before we tag a 1.0.0 release, we will write some proper user-facing documentation.

## Features

- [ ] General
    - [x] Multi server support
    - [x] Global bans & mutes
    - [x] [Docker support](https://hub.docker.com/repository/docker/leighmacdonald/gbans)
    - [ ] Multi-tenant support
- [ ] Import/Export of gbans databases
    - [ ] Backend linking of gbans services to enable use of other operators lists in real-time.
    - [x] [tf2_bot_detector](https://github.com/PazerOP/tf2_bot_detector) compatible player list
- [x] Game support
    - [x] Team Fortress 2
- [ ] Blocking lists & types
    - [x] Valves source server banip config
    - [ ] Existing sourcebans database
    - [x] [CIDR/IP](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) bans
    - [x] [tf2_bot_detector](https://github.com/PazerOP/tf2_bot_detector/blob/master/staging/cfg/playerlist.official.json)
    - [ ] Known VPN Networks
    - [ ] Known non-residential addresses
    - [ ] Known proxies
    - [ ] [FireHOL](https://github.com/firehol/blocklist-ipsets) databases
- [x] Database support
    - [x] Postgresql w/PostGIS
- [x] Relay game logs to central service
- [ ] SourceMod Plugin
    - [x] Game server authentication
    - [x] Restrict banned players from connecting
    - [x] Restrict muted/gagged players on join
    - [x] Dynamically load admins/mods/reserved slots from gbans
    - [ ] `/gb_ban <player_id|steam_id> duration Reason` Ban a user
    - [ ] `/gb_unban` Unban a previously banned user
    - [ ] `/gb_kick` Kick a user
    - [x] `/gb_mod or /mod <message>` Call for a mod on discord
- [ ] User Interfaces
    - [ ] CLI
    - [x] Discord
    - [x] Web
    - [ ] Web
        - [x] Server Browser
        - [x] News
        - [x] Wiki
        - [x] User reports
        - [x] Ban Appeals
        - [ ] Quick Play
        - [ ] Pug

## Docker

Docker is recommended to run gbans. You can find the official docker images on
[ghcr.io](https://github.com/leighmacdonald/gbans/pkgs/container/gbans).

Assuming you have created your config file and have a database setup you can run it using something
like:

    docker run -it --rm -v `$(pwd)`/gbans.yml:/app/gbans.yml:ro ghcr.io/leighmacdonald/gbans:latest

There is also a docker-compose config you can use which provides the database as well.

    docker-compose -f docker/docker-compose.yml up --build --remove-orphans --abort-on-container-exit --exit-code-from gbans

## Documentation

For installation, configuration and usage instruction, developer notes, please see the [docs](docs) directory.
