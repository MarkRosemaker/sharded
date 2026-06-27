# sharded

High-performance concurrent map for Go using lock striping (256 shards). Up to 8× faster than a single `sync.Mutex` under contention.

Based on the design from [Shard your locks: benchmarking 6 Go cache designs](https://strebkov.dev/posts/shard-your-locks/).

## Features

- Simple, familiar API (`Get`/`Set`/`Delete`/`Len`)
- Excellent scaling with core count
- Minimal overhead and allocations
- Cache-line padded shards to reduce false sharing
- Fast key-to-shard routing via FNV-1a hash and bitmask

## Installation

```bash
go get github.com/MarkRosemaker/sharded
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/MarkRosemaker/sharded"
)

func main() {
	m := sharded.NewStringMap[int]()

	m.Set("key", 3)

	v, ok := m.Get("key")
	fmt.Println(v, ok) // 3 true

	m.Delete("key")
	fmt.Println(m.Len()) // 0
}
```

## Benchmarks

See the original article for detailed results across read-only, read-heavy, balanced, and write-heavy workloads. The sharded implementation consistently ranks at or near the top.

## Why This Design?

- A single mutex becomes a major bottleneck under concurrency.
- `sync.RWMutex` often performs worse than expected, especially with writes.
- `sync.Map` has higher overhead and shines only in specific scenarios.
- Sharded maps with per-shard locks provide the best balance for general use.
