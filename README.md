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
    ./fsp-rewards-calculator -n flare -e 123 -h localhost -p 3306 -d flare_ftso_indexer -u root -w root
    ```

   If using default db connection parameters, and calculating for the previous reward epoch, you can specify only the
   network:
    ```sh
    ./fsp-rewards-calculator -n flare
    ```

3. Results will be produced under `./results/<network>/<epoch>`.

4. Verbose mode (optional).
   When run with the `-v` flag, the calculator will write additional detailed JSON files per round and other
   intermediate result files under `./results/<network>/<epoch>/` (for example: per-round claims, signing/finalization
   details, penalties).
   This is useful for debugging or auditing. Example:

    ```sh
    ./fsp-rewards-calculator -n flare -e 123 -v
    ```

## FCC fee accounting

From the reward epoch each network activates it in (Songbird 419, Coston 5877; not yet on Flare), the fees of the
Flare Confidential Compute contracts are accounted for. Both `FlareTeeManager.TeeInstructionsSent` and
`Fdc2Hub.AttestationRequested` credit their fee to the `RewardManager` when it is paid, so every wei of them must be
covered by a claim; until the TEE rewarding logic exists all of it is claimed to the network's FCC fee address as a
single direct claim.

**The indexer must collect the FCC event logs.** A stock FSP-mode indexer has a hardcoded contract list that excludes
both FCC contracts, and missing events are indistinguishable from no activity: the calculation would then produce a
merkle root that leaves the fees unclaimed, with nothing to indicate why. Add two collectors per network:

```toml
[[indexer.collect_logs]]
contract_address = "0x…" # FlareTeeManager (diamond)
topic = "0xf770e69a9fc05b7180797556ec4cedb6108ce2c56ffa76c84aa087efeb5e6963" # TeeInstructionsSent

[[indexer.collect_logs]]
contract_address = "0x…" # Fdc2Hub (proxy)
topic = "0x57c4413905bb1b444f93a5eab5a942fae34c0fcaa1c25cc595ce0b990310f5de" # AttestationRequested
```

The topics are event signature hashes and identical on every network; the addresses are in
`common/params/<network>.go`. They must be collecting before the first epoch that accounts for them: the events
cannot be backfilled from a public node, whose `eth_getLogs` is capped far below a useful range.

To confirm an epoch's fees independently of the indexer, compare the sum of all claims against
`RewardManager.getRewardEpochTotals(<epoch>)` over RPC - the funds credited by `receiveRewards` emit no event and
so cannot be derived from the indexer database.

## Logging

- Default log level is `INFO`.
- To enable debug logs, set `LOG_LEVEL=DEBUG` before running:

```sh
LOG_LEVEL=DEBUG ./fsp-rewards-calculator -n flare -e 123
```

- SQL query logs (including duration) are printed only when `LOG_LEVEL=DEBUG`.
