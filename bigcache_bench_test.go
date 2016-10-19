package bigcache

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var message = blob('a', 256)

func BenchmarkWriteToCacheWith1Shard(b *testing.B) {
	writeToCache(b, 1, 100*time.Second, b.N)
}

func BenchmarkWriteToLimitedCacheWithSmallInitSizeAnd1Shard(b *testing.B) {
	m := blob('a', 1024)
	cache, _ := NewBigCache(Config{1, 100 * time.Second, 100, 256, false, nil, 1, nil})
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), m)
	}
}

func BenchmarkWriteToUnlimitedCacheWithSmallInitSizeAnd1Shard(b *testing.B) {
	m := blob('a', 1024)
	cache, _ := NewBigCache(Config{1, 100 * time.Second, 100, 256, false, nil, 0, nil})
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), m)
	}
}

func BenchmarkWriteToCache(b *testing.B) {
	for _, shards := range []int{1, 512, 1024, 8192} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			writeToCache(b, shards, 100*time.Second, b.N)
		})
	}
}

func BenchmarkReadFromCache(b *testing.B) {
	for _, shards := range []int{1, 512, 1024, 8192} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			readFromCache(b, 1024)
		})
	}
}

func BenchmarkIterateOverCache(b *testing.B) {
	m := blob('a', 1)

	for _, shards := range []int{512, 1024, 8192} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			cache, _ := NewBigCache(Config{shards, 1000 * time.Second, max(b.N, 100), 500, false, nil, 0, nil})

			for i := 0; i < b.N; i++ {
				cache.Set(fmt.Sprintf("key-%d", i), m)
			}

			b.ResetTimer()
			it := cache.Iterator()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if it.HasNext() {
						it.Value()
					}
				}
			})
		})
	}
}

func BenchmarkWriteToCacheWith1024ShardsAndSmallShardInitSize(b *testing.B) {
	writeToCache(b, 1024, 100*time.Second, 100)
}

func writeToCache(b *testing.B, shards int, lifeWindow time.Duration, requestsInLifeWindow int) {
	cache, _ := NewBigCache(Config{shards, lifeWindow, max(requestsInLifeWindow, 100), 500, false, nil, 0, nil})
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Int()
		counter := 0
		for pb.Next() {
			cache.Set(fmt.Sprintf("key-%d-%d", id, counter), message)
			counter = counter + 1
		}
	})
}

func readFromCache(b *testing.B, shards int) {
	cache, _ := NewBigCache(Config{shards, 1000 * time.Second, max(b.N, 100), 500, false, nil, 0, nil})
	for i := 0; i < b.N; i++ {
		cache.Set(strconv.Itoa(i), message)
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get(strconv.Itoa(rand.Intn(b.N)))
		}
	})
}
