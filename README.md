<!-- LOGO -->
<div align="center">
  <a href="https://flare.network/" target="blank">
    <img src="https://content.flare.network/Flare-2.svg" width="300" alt="Flare Logo" />
  </a>
  <br />
  <a href="CONTRIBUTING.md">Contributing</a>
  ·
  <a href="SECURITY.md">Security</a>
  ·
  <a href="CHANGELOG.md">Changelog</a>
</div>


# Flare Systems Protocol Rewards Calculator

## Overview

Reward calculator for Flare Systems Protocols. Produces a reward Merkle root hash and claims for a specified epoch.

## Prerequisites

- Go 1.25.5
- Access to a [FSP C-Chain indexer](https://github.com/flare-foundation/flare-system-c-chain-indexer) instance.

**Note:** Reward calculation performance is primarily bounded by network I/O when retrieving data from the indexer. For
fastest results, run the calculator on the same host as the indexer.

## Configuration

The application uses command-line flags to configure its parameters. The following flags are available:

| Flag | Type   | Description                                                 | Default            |
|------|--------|-------------------------------------------------------------|--------------------|
| `-n` | string | Network (coston, songbird, flare)                           | -                  |
| `-e` | uint64 | Reward epoch id                                             | previous epoch     |
| `-h` | string | Indexer db host                                             | localhost          |
| `-p` | int    | Indexer db port                                             | 3306               |
| `-d` | string | Indexer db name                                             | flare_ftso_indexer |
| `-u` | string | Indexer db user                                             | root               |
| `-w` | string | Indexer db password                                         | root               |
| `-v` | bool   | Verbose output - write detailed per-round result claim data | false              |

## Usage

1. Build the application:
    ```sh
    go build -o fsp-rewards-calculator
    ```

2. Run the application with the required flags:
    ```sh
    ./fsp-rewards-calculator -n <network> -e <epoch> -h <db_host> -p <db_port> -d <db_name> -u <db_user> -w <db_password>
    ```

   Example:
    ```sh
    ./fsp-rewards-calculator -n coston -e 123 -h localhost -p 3306 -d flare_ftso_indexer -u root -w root
    ```

   If using default db connection parameters, and calculating for the previous reward epoch, you can specify only the
   network:
    ```sh
    ./fsp-rewards-calculator -n coston
    ```

3. Results will be produced under `./results/<network>/<epoch>`.

4. Verbose mode (optional).
   When run with the `-v` flag, the calculator will write additional detailed JSON files per round and other
   intermediate result files under `./results/<network>/<epoch>/` (for example: per-round claims, signing/finalization
   details, penalties).
   This is useful for debugging or auditing. Example:

    ```sh
    ./fsp-rewards-calculator -n coston -e 123 -v
    ```