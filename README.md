# Marte

Marte is an exploit farm for CTF Attack/Defense focused on performance, reliability, and extensibility. It automates exploit execution, flag management, and competition monitoring through an integrated dashboard.

## Architecture

Marte consists of:
- **Client**: CLI from which exploits are executed
- **Server**:
  - Web dashboard with API for monitoring
  - Optimized TCP server for flag collection
  - Submits flags to the competition server

## Stack

| Component | Technology                               |
|-----------|------------------------------------------|
| Backend   | Go 1.26, SQLite (pure Go), standard library |
| Client    | Go + Cobra CLI                           |
| Frontend  | HTML + CSS + JS, Bootstrap 5.3, Chart.js |
| Database  | SQLite via modernc.org/sqlite            |

## Backend

The backend exposes two services:
- **HTTP** (port 14100): REST API + web dashboard
- **TCP** (port 14101): flag reception from clients

The dashboard shows real-time statistics: accepted/rejected/pending flags, timeline, attacker leaderboard.

## Client

The client is a CLI for managing exploits.

```
marte init                                  Create local environment
marte host <ip> <port>                      Set the backend address
marte login <username> <password>           Authentication
marte pull                                  Download services and configuration
marte exploit create <file>                 Create exploit template
marte exploit run <file> <service>          Run exploit against all teams
marte exploit test <file> <service>         Test on NOP team
marte exploit debug <file> <service> <host> Debug exploit (useful for testing)
```

## Installation

Installation is only supported on macOS and Linux. If you are one Windows, use **WSL2**.

### Client

One liner:

```bash
curl -sSfL https://raw.githubusercontent.com/xEcho1337/Marte/main/scripts/install-client.sh | bash
```

The script will install the Marte's client inside the `~/.local/bin` directory.

### Server

One liner:

```bash
curl -sSfL https://raw.githubusercontent.com/xEcho1337/Marte/main/scripts/install-server.sh | bash
cd Marte
```

After configuring `data/config.yml`, run with:

```bash
docker compose up -d
```

## Build

```bash
scripts/build-client.sh    # client/out/marte
scripts/build-backend.sh   # backend/out/marte
```