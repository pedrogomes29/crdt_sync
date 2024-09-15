package sync

type commandID int

const (
	BLOOM_FILTER commandID = iota
	BUCKET_DIGESTS
	NON_MATCHING_BUCKETS
	BUCKET_DIFFS
)

type command struct {
	id   commandID
	peer *peer
	args [][]byte
}
