# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!--
## [Unreleased]

### Added

- ...

### Changed

- ...
-->

## [1.1.0] - 2026-07-22

### Added

- Support FIP.16 vote-power unification from reward epoch 417 on Flare and Songbird, including signing-weight-based FTSO medians, stake-aware reward splitting, random-only reveals, and the FDC fee share routed to the FIRE pool.
- Support the upgraded VoterRegistry and FlareSystemsCalculator contracts on Flare, Songbird, and Coston.

### Changed

- Rename the bridged TON/USD feed to GRAM/USD.

### Fixed

- Match the reference calculator's weighted median, upper-quartile, and deterministic FDC consensus-bit-vote selection.
- Correct signing reward deadline and eligibility-window handling.
- Scan the complete secure-random lookahead window and fail explicitly when voter registration data is missing.
- Skip zero-weight FDC signing voters.
- Correct community reward-offer accumulation and ordering.

## [1.0.1] - 2026-03-13

### Fixed

- Gracefully handle delisted feeds for Fast Updates reward calculation.

## [1.0.0] - 2026-02-24

### Added

- Initial release
