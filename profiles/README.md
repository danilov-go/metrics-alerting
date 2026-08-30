# Оптимизация производительности

## Анализ производительности

При анализе базового профиля памяти (pprof) выявлено две ключевые проблемы:
1. GzipMiddleware: на каждый ответ сервер заново создавал объекты сжатия.
2. Постоянное создание buf в HTTP-хендлерах:на каждый запрос рантайм Go заново выделял буферы под JSON.

### Профиль до оптимизации

flat  flat%   sum%        cum   cum%
 1805.17kB 54.01% 54.01%  1805.17kB 54.01%  compress/flate.NewWriter (inline)
  512.88kB 15.35% 69.36%   512.88kB 15.35%  reflect.growslice
  512.05kB 15.32% 84.68%   512.05kB 15.32%  reflect.unsafe_New
 -512.05kB 15.32% 69.36%  -512.05kB 15.32%  sync.runtime_notifyListWait
         0     0% 69.36%  1805.17kB 54.01%  compress/gzip.(*Writer).Write
         0     0% 69.36%  1024.93kB 30.67%  encoding/json.(*decodeState).array
         0     0% 69.36%   512.05kB 15.32%  encoding/json.(*decodeState).literalStore
         0     0% 69.36%   512.05kB 15.32%  encoding/json.(*decodeState).object
         0     0% 69.36%  1024.93kB 30.67%  encoding/json.(*decodeState).unmarshal
         0     0% 69.36%  1024.93kB 30.67%  encoding/json.(*decodeState).valueещз

### Результаты бенчмарков до оптимизации

BenchmarkApiUpdateHandler-8       507559              2599 ns/op            2773 B/op         18 allocs/op
BenchmarkApiValueHandler-8        481041              2297 ns/op            2772 B/op         18 allocs/op
BenchmarkApiUpdatesHandler-8      119572              9892 ns/op            7747 B/op         52 allocs/op

## Оптимизация

Для устранения пиков аллокаций и оптимизации работы с памятью были реализованы следующие решения:
1. Оптимизация GzipMiddleware
 1.1 Внедрен sync.Pool для повторного использования тяжелых объектов сжатия. Теперь вместо создания нового писателя на каждый HTTP-ответ рантайм Go переиспользует существующий объект. При этом писатели инициализируются с уровнем gzip.BestSpeed, что обеспечивает максимальную скорость обработки данных и минимальную нагрузку на CPU.
2. Оптимизация HTTP-хендлеров
 2.1 Внедрен sync.Pool для переиспользования буферов `*bytes.Buffer` при чтении запросов и парсинге через json.Unmarshal. Буферы больше не выделяются в памяти с нуля под каждый запрос — они берутся из пула.

### Результаты бенчмарков после оптимизации

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

## Выводы

1. Благодаря оптимизации HTTP-хендлеров удалось добиться следующих результатов:
   * 1.1. Для одиночных запросов (ApiUpdateHandler / ApiValueHandler) потребление памяти снизилось с 2773 B/op до 1291 B/op (аллокации упали с 18 до 16).
   * 1.2. Для пакетных запросов (ApiUpdatesHandler) память уменьшилась с 7747 B/op до 3695 B/op (аллокации упали с 52 до 49).
   * 1.3. Внедрение пула буферов исключило создание новой памяти под каждый запрос (bytes.growSlice: -514kB). Благодаря этому encoding/json.Unmarshal автоматически перестал тратить лишнюю память на десериализацию метрик, сэкономив еще -1025.11kB.

2. Благодаря оптимизации GzipMiddleware удалось добиться следующих результатов:
   * 2.1. Дифференциальный профиль зафиксировал экономию памяти: -5415.52kB в compress/flate.NewWriter и -1127.67kB в newDeflateFast.
