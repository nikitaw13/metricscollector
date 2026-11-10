
# Дано

### Требуется оптимизация decompressMiddleware
```bash
🐊 go tool pprof -alloc_space base.pprof 
File: server
Build ID: 919ecce2a43c62c3f0df31246426984ba1f28aaf
Type: alloc_space
Time: 2026-11-10 10:06:07 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top20
Showing nodes accounting for 38785.32kB, 84.40% of 45953.99kB total
Showing top 20 nodes out of 221
      flat  flat%   sum%        cum   cum%
12147.83kB 26.43% 26.43% 12147.83kB 26.43%  compress/flate.(*dictDecoder).init (inline)
    3591kB  7.81% 34.25%     3591kB  7.81%  runtime.mallocgc
 3073.50kB  6.69% 40.94%  3073.50kB  6.69%  crypto/internal/fips140/bigmod.(*Nat).reset (inline)
 2570.83kB  5.59% 46.53% 14718.66kB 32.03%  compress/flate.NewReader
 2570.01kB  5.59% 52.12%  2570.01kB  5.59%  bufio.NewReaderSize (inline)
 2101.71kB  4.57% 56.70%  2101.71kB  4.57%  github.com/jackc/pgx/v5/internal/stmtcache.NewLRUCache
    2050kB  4.46% 61.16%  3074.50kB  6.69%  crypto/internal/fips140/bigmod.(*Nat).montgomeryMul
 1377.12kB  3.00% 64.16%  1377.12kB  3.00%  compress/flate.(*fastGen).addBlock
 1310.59kB  2.85% 67.01%  1855.25kB  4.04%  compress/flate.NewWriter (inline)
 1025.75kB  2.23% 69.24%  1025.75kB  2.23%  sync.(*Pool).pinSlow
 1024.78kB  2.23% 71.47%  1024.78kB  2.23%  vendor/golang.org/x/crypto/cryptobyte.(*Builder).add
 1024.12kB  2.23% 73.70%  1536.23kB  3.34%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
 1024.02kB  2.23% 75.93%  1024.02kB  2.23%  reflect.unsafe_New
  768.26kB  1.67% 77.60%   768.26kB  1.67%  go.uber.org/zap/zapcore.newCounters
  544.67kB  1.19% 78.78%   544.67kB  1.19%  compress/flate.(*compressor).init
  525.43kB  1.14% 79.93%   525.43kB  1.14%  github.com/jackc/pgx/v5/pgtype.(*Map).buildReflectTypeToType
  516.01kB  1.12% 81.05%   516.01kB  1.12%  crypto/internal/fips140/mlkem.GenerateKey768
  515.19kB  1.12% 82.17%  2370.45kB  5.16%  runtime/pprof.(*profileBuilder).emitLocation
  512.34kB  1.11% 83.29% 17801.02kB 38.74%  compress/gzip.NewReader
  512.16kB  1.11% 84.40%   512.16kB  1.11%  encoding/json/jsontext.NewDecoder
```
```text
(pprof) peek decompress
Showing nodes accounting for 45953.99kB, 100% of 45953.99kB total
----------------------------------------------------------+-------------
      flat  flat%   sum%        cum   cum%   calls calls% + context 	 	 
----------------------------------------------------------+-------------
                                        34247.83kB   100% |   net/http.HandlerFunc.ServeHTTP
         0     0%     0% 34247.83kB 74.53%                | github.com/nikitaw13/metricscollector/internal/handler.decompressMiddleware.func1
                                        17801.02kB 51.98% |   compress/gzip.NewReader (inline)
                                        16446.81kB 48.02% |   net/http.HandlerFunc.ServeHTTP
----------------------------------------------------------+-------------
```
```text
Total: 44.88MB
ROUTINE ======================== github.com/nikitaw13/metricscollector/internal/handler.decompressMiddleware.func1 in /home/prometherus/Projects/GoDev/metricscollector/internal/handler/middleware.go
         0    33.45MB (flat, cum) 74.53% of Total
         .          .    204:	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
         .          .    205:		switch r.Header.Get("Content-Encoding") {
         .          .    206:		case "gzip":
         .    17.38MB    207:			gz, err := gzip.NewReader(r.Body)
         .          .    208:			if err != nil {
         .          .    209:				http.Error(w, err.Error(), http.StatusInternalServerError)
         .          .    210:				return
         .          .    211:			}
```

```bash
🐊 go test -bench=Decompress -benchmem -count=5
goos: linux
goarch: amd64
pkg: github.com/nikitaw13/metricscollector/internal/handler
cpu: 12th Gen Intel(R) Core(TM) i5-1235U
BenchmarkDecompressMiddleware-12    	   60234	     19364 ns/op	   51395 B/op	      24 allocs/op
BenchmarkDecompressMiddleware-12    	   58735	     19670 ns/op	   51396 B/op	      24 allocs/op
BenchmarkDecompressMiddleware-12    	   59316	     21326 ns/op	   51402 B/op	      24 allocs/op
BenchmarkDecompressMiddleware-12    	   53893	     19877 ns/op	   51392 B/op	      24 allocs/op
BenchmarkDecompressMiddleware-12    	   58933	     19302 ns/op	   51394 B/op	      24 allocs/op
PASS
ok  	github.com/nikitaw13/metricscollector/internal/handler	5.980s
```
```bash
🐊 go test -bench=Decompress -race   
goos: linux
goarch: amd64
pkg: github.com/nikitaw13/metricscollector/internal/handler
cpu: 12th Gen Intel(R) Core(TM) i5-1235U
BenchmarkDecompressMiddleware-12    	   27128	     42898 ns/op	   53340 B/op	      24 allocs/op
PASS
ok  	github.com/nikitaw13/metricscollector/internal/handler	2.327s
```

## После оптимизации decompressMiddleware
```bash
🐊 go tool pprof -alloc_space -diff_base=base.pprof result.pprof 
File: server
Build ID: f25b34aa35bc5407800235b1fd8892addef8f812
Type: alloc_space
Time: 2026-11-10 10:06:07 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for -18.45MB, 41.12% of 44.88MB total
Dropped 28 nodes (cum <= 0.22MB)
Showing top 10 nodes out of 252
      flat  flat%   sum%        cum   cum%
  -10.32MB 22.99% 22.99%   -10.32MB 22.99%  compress/flate.(*dictDecoder).init (inline)
   -2.01MB  4.48% 27.46%   -12.33MB 27.46%  compress/flate.NewReader
       2MB  4.46% 23.01%     2.50MB  5.57%  net/http.(*conn).readRequest
   -1.50MB  3.34% 26.35%    -1.50MB  3.34%  crypto/internal/fips140/bigmod.(*Nat).reset
   -1.34MB  3.00% 29.35%    -1.34MB  3.00%  compress/flate.(*fastGen).addBlock
   -1.28MB  2.85% 32.20%    -1.81MB  4.04%  compress/flate.NewWriter
      -1MB  2.23% 34.43%       -1MB  2.23%  sync.(*Pool).pinSlow
      -1MB  2.23% 36.66%       -1MB  2.23%  vendor/golang.org/x/crypto/cryptobyte.(*Builder).add
      -1MB  2.23% 38.89%    -0.50MB  1.11%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
      -1MB  2.23% 41.12%       -1MB  2.23%  reflect.unsafe_New
```

На замену `compress/gzip.NewReader` пришел `compress/gzip.(*Reader).Reset`, которого ранее не было. 
```text
(pprof) peek decompress
Showing nodes accounting for -18666.17kB, 40.62% of 45953.99kB total
----------------------------------------------------------+-------------
      flat  flat%   sum%        cum   cum%   calls calls% + context 	 	 
----------------------------------------------------------+-------------
                                       -17240.02kB   100% |   net/http.HandlerFunc.ServeHTTP
         0     0%     0% -17240.02kB 37.52%                | github.com/nikitaw13/metricscollector/internal/handler.decompressMiddleware.func1
                                       -17801.02kB 103.25% |   compress/gzip.NewReader (inline)
                                         4153.82kB 24.09% |   compress/gzip.(*Reader).Reset
                                        -3592.82kB 20.84% |   net/http.HandlerFunc.ServeHTTP
----------------------------------------------------------+-------------
```
```text
(pprof) list decompressMiddleware
Total: 44.88MB
ROUTINE ======================== github.com/nikitaw13/metricscollector/internal/handler.decompressMiddleware.func1 in /home/prometherus/Projects/GoDev/metricscollector/internal/handler/middleware.go
         0   -16.84MB (flat, cum) 37.52% of Total
         .          .    204:	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
         .          .    205:		switch r.Header.Get("Content-Encoding") {
         .          .    206:		case "gzip":
         .   -17.38MB    207:			// gz, err := gzip.NewReader(r.Body)
         .          .    208:			// if err != nil {
         .          .    209:			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
         .          .    210:			// 	return
         .          .    211:			// }
         .          .    212:
         .          .    213:			// defer gz.Close()
         .          .    214:
         .          .    215:			gz := gzReaderPool.Get().(*gzip.Reader)
         .     4.06MB    216:			err := gz.Reset(r.Body)
         .          .    217:			if err != nil {
         .          .    218:				gzReaderPool.Put(gz)
         .          .    219:				http.Error(w, err.Error(), http.StatusInternalServerError)
         .          .    220:				return
         .          .    221:			}
         .          .    222:			defer func() {
         .          .    223:				gz.Close()
         .          .    224:				gzReaderPool.Put(gz)
         .          .    225:			}()
```

#### В ~3 раза быстрее, в пять раз меньше объём аллоцированных байт за операцию
```bash
🐊 go test -bench=Decompress -benchmem -count=5
goos: linux
goarch: amd64
pkg: github.com/nikitaw13/metricscollector/internal/handler
cpu: 12th Gen Intel(R) Core(TM) i5-1235U
BenchmarkDecompressMiddleware-12    	  167869	      6775 ns/op	   10095 B/op	      19 allocs/op
BenchmarkDecompressMiddleware-12    	  167752	      6595 ns/op	   10090 B/op	      19 allocs/op
BenchmarkDecompressMiddleware-12    	  165700	      6341 ns/op	   10094 B/op	      19 allocs/op
BenchmarkDecompressMiddleware-12    	  183093	      6489 ns/op	   10090 B/op	      19 allocs/op
BenchmarkDecompressMiddleware-12    	  185014	      6519 ns/op	   10090 B/op	      19 allocs/op
PASS
ok  	github.com/nikitaw13/metricscollector/internal/handler	5.887s
```
```bash
🐊 go test -bench=Decompress -race   
goos: linux
goarch: amd64
pkg: github.com/nikitaw13/metricscollector/internal/handler
cpu: 12th Gen Intel(R) Core(TM) i5-1235U
BenchmarkDecompressMiddleware-12    	   37214	     33325 ns/op	   22423 B/op	      21 allocs/op
PASS
ok  	github.com/nikitaw13/metricscollector/internal/handler	2.410s
```