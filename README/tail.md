## Benchmarks

See the original article for detailed results across read-only, read-heavy, balanced, and write-heavy workloads. The sharded implementation consistently ranks at or near the top.

## Why This Design?

- A single mutex becomes a major bottleneck under concurrency.
- `sync.RWMutex` often performs worse than expected, especially with writes.
- `sync.Map` has higher overhead and shines only in specific scenarios.
- Sharded maps with per-shard locks provide the best balance for general use.
