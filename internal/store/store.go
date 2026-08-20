package store

// NumShards is number of partitions in store
// Power of 2 used for bitwise AND, faster than division modulo. To be tuned
const NumShards = 64

// Store is a shareded KV store
type Store struct {
	shards [NumShards]*shard
}

// New creates a Store with all shards initialised with a loop
func New() *Store {
	s := Store{}
	for i := 0; i < NumShards; i++ {
		s.shards[i] = newShard()
	}
	return &s
}

// getShard returns the shard responsible for the given key
// Sharding happens with FNV-1a; fast, good avalanche and distribution, deterministics
func (s *Store) getShard(key string) *shard {
	hash := fnv1a(key)
	index := hash & (NumShards - 1)
	return s.shards[index]
}

// fnv1a computer FNV-1a hash
// Algorithm: start with offset, and then for each byte, XOR with byte and multiply by FNV prime
func fnv1a(key string) uint64 {
	const offset = 14695981039346656037
	const prime = 1099511628211

	hash := uint64(offset)

	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime
	}
	return hash
}

// Get retrieves value from store
func (s *Store) Get(key string) ([]byte, bool) {
	shard := s.getShard(key)
	return shard.Get(key)
}

// Set stores a key-value pair.
func (s *Store) Set(key string, value []byte) {
	shard := s.getShard(key)
	shard.Set(key, value)
}

// Delete removes a key from the store.
func (s *Store) Delete(key string) bool {
	shard := s.getShard(key)
	return shard.Delete(key)
}
