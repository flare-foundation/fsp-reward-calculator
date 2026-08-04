//go:generate  abigen --abi=tee.abi --pkg=tee --type=TeeManager --out=autogen.go

// Package tee holds the binding for the FlareTeeManager contract, which emits the TeeInstructionsSent
// event carrying the TEE instruction dispatch fees credited to the RewardManager.
//
// tee.abi holds only the TeeInstructionsSent event, taken from the compiled InstructionsFacet artifact.
// FlareTeeManager is an EIP-2535 diamond and the event is declared in IInstructions.sol and emitted from
// a library inlined into InstructionsFacet, so it is absent from the FlareTeeManager.sol artifact, which
// carries only DiamondCut - do not "correct" the source to the diamond artifact, as the resulting ABI
// contains no matching event. The events are queried at the diamond address
// (params.Net.Contracts.FlareTeeManager). Add further entries from the artifact if more are ever read.
package tee
