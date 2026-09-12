High-performance concurrent map for Go using lock striping (256 shards). Up to 8× faster than a single `sync.Mutex` under contention.

Based on the design from [Shard your locks: benchmarking 6 Go cache designs](https://strebkov.dev/posts/shard-your-locks/).
