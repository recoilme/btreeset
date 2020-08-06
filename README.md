# `btreeset`

[![GoDoc](https://godoc.org/github.com/recoilme/btreeset?status.svg)](https://godoc.org/github.com/recoilme/btreeset)

Just an itsy bitsy b-tree set. Based on github.com/tidwall/tinybtree

## Usage

Put keys in and you are done.

### Functions

```

```

### Example

```go

```

### Benchmark

```
go test -benchmem -run=^$ github.com/recoilme/btreeset -bench BenchmarkTidwall
seed: 1596710262060548000
goos: darwin
goarch: amd64
pkg: github.com/recoilme/btreeset
BenchmarkTidwallSequentialSet-8          4176504               468 ns/op              72 B/op          1 allocs/op
BenchmarkTidwallSequentialGet-8          2481506               534 ns/op               0 B/op          0 allocs/op
BenchmarkTidwallRandomSet-8              1000000              1206 ns/op              54 B/op          1 allocs/op
BenchmarkTidwallRandomGet-8              1000000              1085 ns/op               0 B/op          0 allocs/op
```
## Contact

Vadim Kulibaba [@recoilme](http://t.me/recoilme)

## License

`btreeset` source code is available under the MIT [License](/LICENSE).
