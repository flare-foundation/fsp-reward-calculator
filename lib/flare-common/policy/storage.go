package policy

import (
	"cmp"
	"fmt"
	"sort"
)

// Does not lock the structure, should be called from a function that does lock.
// We assume that the list is sorted by rewardEpochId and also by startVotingRoundId.
func (s *SigningPolicyStorage) findByVotingRoundId(votingRoundId uint32) *SigningPolicy {
	i, found := sort.Find(len(s.spList), func(i int) int {
		return cmp.Compare(votingRoundId, s.spList[i].StartVotingRoundId)
	})
	if found {
		return s.spList[i]
	}
	if i == 0 {
		return nil
	}
	return s.spList[i-1]
}

func (s *SigningPolicyStorage) Add(sp *SigningPolicy) error {
	s.Lock()
	defer s.Unlock()

	if len(s.spList) > 0 {
		// check consistency, previous epoch should be already added
		if s.spList[len(s.spList)-1].RewardEpochId != sp.RewardEpochId-1 {
			return fmt.Errorf("missing signing policy for reward epoch id %d", sp.RewardEpochId-1)
		}
		// should be sorted by voting round id, should not happen
		if sp.StartVotingRoundId < s.spList[len(s.spList)-1].StartVotingRoundId {
			return fmt.Errorf("signing policy for reward epoch id %d has larger start voting round id than previous policy",
				sp.RewardEpochId)
		}
	}

	s.spList = append(s.spList, sp)
	return nil
}

// Return the signing policy for the voting round, or nil if not found.
// Also returns true if the policy is the last one or false otherwise.
func (s *SigningPolicyStorage) GetForVotingRound(votingRoundId uint32) (*SigningPolicy, bool) {
	s.Lock()
	defer s.Unlock()

	sp := s.findByVotingRoundId(votingRoundId)
	if sp == nil {
		return nil, false
	}
	return sp, sp.RewardEpochId == s.spList[len(s.spList)-1].RewardEpochId
}

// Removes all signing policies with start voting round id <= than the provided one.
// Returns the list of removed reward epoch ids.
func (s *SigningPolicyStorage) RemoveByVotingRound(votingRoundId uint32) []uint32 {
	s.Lock()
	defer s.Unlock()

	var removedRewardEpochIds []uint32
	for len(s.spList) > 0 && s.spList[0].StartVotingRoundId <= votingRoundId {
		removedRewardEpochIds = append(removedRewardEpochIds, uint32(s.spList[0].RewardEpochId))
		s.spList[0] = nil
		s.spList = s.spList[1:]
	}
	return removedRewardEpochIds
}

// RemoveByVotingRoundSafe removes all signing policies that ended strictly before votingRoundId.
// Returns the list of removed reward epoch ids.
func (s *SigningPolicyStorage) RemoveBeforeVotingRound(votingRoundId uint32) []uint32 {
	s.Lock()
	defer s.Unlock()

	var removedRewardEpochIds []uint32
	for len(s.spList) > 1 && s.spList[1].StartVotingRoundId < votingRoundId {
		removedRewardEpochIds = append(removedRewardEpochIds, uint32(s.spList[0].RewardEpochId))
		s.spList[0] = nil
		s.spList = s.spList[1:]
	}
	return removedRewardEpochIds
}
