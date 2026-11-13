# Trie, radix trie and ART

Learning how tries search effectively and why we need ART, since I've been working on NATS infra..
Go implementations with unit tests and a small Python script for benchmarking.
Compare lookup, insert, delete, walking all keys and memory usage.

- Keys have a fixed length of 22 bytes, with shared or random prefixes.
- Run tests `go test ./...`.
- Install matplotlib `python3 -m pip install matplotlib`.
- Run `python3 bench.py --sizes 100000`. Output charts go into `bench-results/`.
