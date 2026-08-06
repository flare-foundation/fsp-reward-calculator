//go:generate  abigen --abi=fdc2.abi --pkg=fdc2 --type=Fdc2 --out=autogen.go

// Package fdc2 holds the binding for the Fdc2Hub contract, which emits the AttestationRequested event
// carrying the FDC2 attestation request fees credited to the RewardManager. fdc2.abi holds only that
// event, taken from the compiled contract artifact; add further entries from it if more are ever read.
//
// Unrelated to the legacy FdcHub (go-flare-common/pkg/contracts/fdchub): different contract, a
// differently named event (AttestationRequested vs AttestationRequest) and a different reward path.
package fdc2
