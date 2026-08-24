package download

// ResumePolicy encodes F-B-11: range fragments of the same nonce share one use.
type ResumePolicy struct {
	CompleteRatio float64
}

func DefaultResume() ResumePolicy {
	return ResumePolicy{CompleteRatio: 0.95}
}

func (p ResumePolicy) Completed(bytesDone, size int64) bool {
	if size <= 0 {
		return false
	}
	ratio := p.CompleteRatio
	if ratio <= 0 || ratio > 1 {
		ratio = 0.95
	}
	return bytesDone >= int64(float64(size)*ratio)
}

func (p ResumePolicy) AllowAnotherHit(uses, maxUses int, bytesDone, size int64) bool {
	if uses < maxUses {
		return true
	}
	return !p.Completed(bytesDone, size)
}

func MergeBytes(prev, add, size int64) int64 {
	n := prev + add
	if size > 0 && n > size {
		return size
	}
	if n < 0 {
		return 0
	}
	return n
}
