## Оптимизация производительности

### до оптимизации

BenchmarkApiUpdateHandler-8       507559              2599 ns/op            2773 B/op         18 allocs/op
BenchmarkApiValueHandler-8        481041              2297 ns/op            2772 B/op         18 allocs/op
BenchmarkApiUpdatesHandler-8      119572              9892 ns/op            7747 B/op         52 allocs/op

### после оптимизации

BenchmarkApiUpdateHandler-8       560126              2058 ns/op            1291 B/op         16 allocs/op
BenchmarkApiValueHandler-8        653864              1827 ns/op            1291 B/op         16 allocs/op
BenchmarkApiUpdatesHandler-8      128104              9274 ns/op            3695 B/op         49 allocs/op

### Результаты профилирования

Duration: 40.01s, Total samples = 3342.16kB 
Showing nodes accounting for -8563.99kB, 256.24% of 3342.16kB total
Dropped 2 nodes (cum <= 16.71kB)
      flat  flat%   sum%        cum   cum%
-5415.52kB 162.04% 162.04% -5998.52kB 179.48%  compress/flate.NewWriter (inline)
-1127.67kB 33.74% 195.78% -1127.67kB 33.74%  compress/flate.newDeflateFast (inline)
-1026.25kB 30.71% 226.48% -1026.25kB 30.71%  compress/flate.(*huffmanEncoder).generate
-1025.06kB 30.67% 257.15% -1025.06kB 30.67%  reflect.growslice
 1024.11kB 30.64% 226.51%  1024.11kB 30.64%  sync.runtime_notifyListWait
  544.67kB 16.30% 210.22%  -583.01kB 17.44%  compress/flate.(*compressor).init
    -514kB 15.38% 225.59%     -514kB 15.38%  bytes.growSlice
 -512.17kB 15.32% 240.92%  -512.17kB 15.32%  net/textproto.readMIMEHeader
 -512.05kB 15.32% 256.24%  -512.05kB 15.32%  github.com/golang-migrate/migrate/v4.(*Migrate).lock.func2
 -512.05kB 15.32% 271.56%  -512.05kB 15.32%  reflect.unsafe_New
  512.02kB 15.32% 256.24%   512.02kB 15.32%  context.WithValue
         0     0% 256.24%     -514kB 15.38%  bytes.(*Buffer).ReadFrom
         0     0% 256.24%     -514kB 15.38%  bytes.(*Buffer).grow
         0     0% 256.24% -1026.25kB 30.71%  compress/flate.(*Writer).Close (inline)
         0     0% 256.24% -1026.25kB 30.71%  compress/flate.(*compressor).close
         0     0% 256.24% -1026.25kB 30.71%  compress/flate.(*compressor).encSpeed
         0     0% 256.24% -1026.25kB 30.71%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 256.24% -1026.25kB 30.71%  compress/flate.(*huffmanBitWriter).writeBlockDynamic
         0     0% 256.24% -1026.25kB 30.71%  compress/gzip.(*Writer).Close
         0     0% 256.24% -5998.52kB 179.48%  compress/gzip.(*Writer).Write
         0     0% 256.24% -1025.11kB 30.67%  encoding/json.(*decodeState).array
         0     0% 256.24% -1025.11kB 30.67%  encoding/json.(*decodeState).unmarshal
         0     0% 256.24% -1025.11kB 30.67%  encoding/json.(*decodeState).value
         0     0% 256.24% -1025.11kB 30.67%  encoding/json.Unmarshal
         0     0% 256.24%  -512.05kB 15.32%  encoding/json.indirect
         0     0% 256.24% -1026.25kB 30.71%  github.com/danilov-go/metrics-alerting.git/internal/handler.(*compressWriter).Close (inline)
         0     0% 256.24% -5998.52kB 179.48%  github.com/danilov-go/metrics-alerting.git/internal/handler.(*compressWriter).Write
         0     0% 256.24% -8563.88kB 256.24%  github.com/danilov-go/metrics-alerting.git/internal/handler.GzipMiddleware.func1
         0     0% 256.24% -1026.25kB 30.71%  github.com/danilov-go/metrics-alerting.git/internal/handler.GzipMiddleware.func1.1
         0     0% 256.24% -7537.63kB 225.53%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 256.24% -8051.86kB 240.92%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 256.24% -7537.63kB 225.53%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 256.24% -7537.63kB 225.53%  main.main.HashMiddleware.func4.1
         0     0% 256.24% -8563.88kB 256.24%  main.main.RequestLogger.func3.1
         0     0% 256.24% -7537.63kB 225.53%  main.main.func1.(*MetricsHandler).ApiUpdatesHandler.3
         0     0% 256.24%  -512.17kB 15.32%  net/http.(*conn).readRequest
         0     0% 256.24% -7539.92kB 225.60%  net/http.(*conn).serve
         0     0% 256.24%  1024.11kB 30.64%  net/http.(*connReader).abortPendingRead
         0     0% 256.24%  1024.11kB 30.64%  net/http.(*response).finishRequest
         0     0% 256.24% -8563.88kB 256.24%  net/http.HandlerFunc.ServeHTTP
         0     0% 256.24%  -512.17kB 15.32%  net/http.readRequest
         0     0% 256.24% -8051.86kB 240.92%  net/http.serverHandler.ServeHTTP
         0     0% 256.24%  -512.17kB 15.32%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 256.24%  -512.05kB 15.32%  reflect.New
         0     0% 256.24% -1025.06kB 30.67%  reflect.Value.Grow
         0     0% 256.24% -1025.06kB 30.67%  reflect.Value.grow
         0     0% 256.24%  1024.11kB 30.64%  sync.(*Cond).Wait