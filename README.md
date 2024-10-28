
# Flare Systems Protocol Rewards Calculator

## Overview

Reward calculator for Flare Systems Protocols, currently supports FTSOv2 Scaling with Fast Updates.
Produces the combined Merkle root for epoch and reward distribution data for claimers.

## Prerequisites

- Go 1.23 or later
- Access to a C-chain indexer instance

## Configuration

The application uses command-line flags to configure its parameters. The following flags are available:

- `-n` (string): Network (coston, coston2, songbird, flare)
- `-e` (uint64): Reward epoch number
- `-h` (string): Database host (default: localhost)
- `-p` (int): Database port (default: 3306)
- `-d` (string): Database name (default: flare_ftso_indexer)
- `-u` (string): Database user (default: root)
- `-w` (string): Database password (default: root)

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
3. Results will be produced under `./results/<network>/<epoch>`.