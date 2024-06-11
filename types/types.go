package types

type EpochId uint64

func (r *EpochId) Add(n uint64) EpochId {
	return EpochId(uint64(*r) + n)
}

type RoundId uint64

func (r *RoundId) Add(n int) RoundId {
	return RoundId(int(*r) + n)
}
