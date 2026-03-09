# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

Сравнение профилей:

```
go tool pprof -sample_index=alloc_objects -top -diff_base=profiles/base_heap_active.pprof profiles/result_heap_active.pprof
File: main
Type: alloc_objects
Time: 2026-03-08 16:43:38 MSK
Duration: 60.01s, Total samples = 947336 
Showing nodes accounting for 25675, 2.71% of 947336 total
Dropped 22 nodes (cum <= 4736)
      flat  flat%   sum%        cum   cum%
     79738  8.42%  8.42%      79738  8.42%  strings.(*Builder).grow
    -65537  6.92%  1.50%     -98305 10.38%  path.Join
    -33861  3.57%  2.08%     -46621  4.92%  go-musthave-shortener/internal/usecase/createshortlinkbatchusecase.(*Usecase).Execute
     32769  3.46%  1.38%      61507  6.49%  context.WithCancel
    -32769  3.46%  2.08%     -32769  3.46%  net/textproto.MIMEHeader.Set (inline)
     32768  3.46%  1.38%      32768  3.46%  github.com/gin-gonic/gin.(*node).getValue
     32768  3.46%  4.84%      32768  3.46%  github.com/google/uuid.NewRandomFromReader
    -32768  3.46%  1.38%     -32768  3.46%  internal/profile.decodeString (inline)
    -32768  3.46%  2.08%     -32768  3.46%  path.(*lazybuf).string (inline)
     26217  2.77%  0.69%      28738  3.03%  context.withCancel (inline)
     24476  2.58%  3.28%      24476  2.58%  net/textproto.readMIMEHeader
    -21845  2.31%  0.97%     -21845  2.31%  io.LimitReader (inline)
     19358  2.04%  3.01%      19358  2.04%  net/textproto.MIMEHeader.Add (inline)
    -16385  1.73%  1.28%     -16385  1.73%  reflect.packEface
    -16384  1.73%  0.45%     -16384  1.73%  compress/flate.(*huffmanDecoder).init
    -16383  1.73%  2.18%     -16383  1.73%  github.com/gin-gonic/gin.(*Context).Set
     15294  1.61%  0.56%     109104 11.52%  net/http.(*conn).readRequest
     13654  1.44%  0.88%      13654  1.44%  encoding/json.appendString[go.shape.string]
     12743  1.35%  2.23%     -85562  9.03%  net/url.(*URL).JoinPath
     12137  1.28%  3.51%      12137  1.28%  bytes.growSlice
     11471  1.21%  4.72%     -93935  9.92%  go-musthave-shortener/internal/di/app.(*DI).initMux.LoggingMiddleware.func3
    -10923  1.15%  3.56%     -10923  1.15%  encoding/json.(*decodeState).literalStore
    -10923  1.15%  2.41%     -21846  2.31%  encoding/json.(*decodeState).object
     10923  1.15%  3.56%      10923  1.15%  fmt.Errorf
     10923  1.15%  4.72%      19116  2.02%  github.com/gin-gonic/gin.(*Context).String (inline)
    -10744  1.13%  3.58%     -17570  1.85%  compress/flate.newHuffmanBitWriter (inline)
    -10179  1.07%  2.51%     -10179  1.07%  net/http.Header.Clone (inline)
      8193  0.86%  3.37%    -108567 11.46%  go-musthave-shortener/internal/di/app.(*DI).initMux.AuthMiddleware.func1
      8193  0.86%  4.24%     -13652  1.44%  net/http.readTransfer
     -8192  0.86%  3.37%      -7936  0.84%  encoding/json.newEncodeState
     -8192  0.86%  2.51%      -8192  0.86%  internal/profile.(*Location).key
      7943  0.84%  3.35%       7943  0.84%  internal/profile.(*Profile).postDecode
     -6826  0.72%  2.63%      -6826  0.72%  compress/flate.newHuffmanEncoder (inline)
      6554  0.69%  3.32%       6554  0.69%  go-musthave-shortener/internal/repository/shorterrepository.(*Repo).AddBatch
     -6554  0.69%  2.63%      -6554  0.69%  net.(*conn).Read
      5960  0.63%  3.26%      -2231  0.24%  internal/profile.(*profileMerger).mapSample
     -5474  0.58%  2.68%      -5420  0.57%  runtime/pprof.(*profileBuilder).emitLocation
     -5461  0.58%  2.10%      -5461  0.58%  internal/profile.init.func5
     -4370  0.46%  1.64%      -4370  0.46%  runtime/pprof.allFrames
     -4097  0.43%  1.21%       1863   0.2%  go-musthave-shortener/internal/usecase/getuserurlsusecase.(*Usecase).Execute
      4096  0.43%  1.64%      23212  2.45%  go-musthave-shortener/internal/usecase/redirectfromshortlinkusecase.(*Usecase).Execute
      3278  0.35%  1.99%      32304  3.41%  net/http.readRequest
      3277  0.35%  2.33%       3277  0.35%  compress/gzip.NewWriterLevel
      2521  0.27%  2.60%       2521  0.27%  context.(*cancelCtx).propagateCancel
     -2050  0.22%  2.38%      -2050  0.22%  encoding/json.(*Decoder).refill
      2050  0.22%  2.60%       2050  0.22%  reflect.growslice
     -1821  0.19%  2.41%      -1821  0.19%  github.com/go-playground/validator/v10.New.func1
      1597  0.17%  2.58%       1597  0.17%  compress/flate.(*huffmanEncoder).generate
      1490  0.16%  2.73%       5160  0.54%  encoding/json.Marshal
       994   0.1%  2.84%     -15390  1.62%  io.ReadAll
      -771 0.081%  2.76%       -771 0.081%  sync.(*Pool).pinSlow
      -195 0.021%  2.74%       -195 0.021%  compress/flate.NewReader
      -173 0.018%  2.72%       -173 0.018%  compress/flate.(*compressor).initDeflate (inline)
       -63 0.0067%  2.71%     -17805  1.88%  compress/flate.NewWriter (inline)
       -54 0.0057%  2.70%      -2285  0.24%  internal/profile.Merge
        53 0.0056%  2.71%     -32715  3.45%  internal/profile.decodeStrings (inline)
        -2 0.00021%  2.71%      21843  2.31%  net/url.parse
         1 0.00011%  2.71%     -17742  1.87%  compress/flate.(*compressor).init
         0     0%  2.71%       -961   0.1%  bufio.(*Writer).Flush
         0     0%  2.71%      20330  2.15%  bytes.(*Buffer).Write
         0     0%  2.71%      -8193  0.86%  bytes.(*Buffer).WriteByte
         0     0%  2.71%      12137  1.28%  bytes.(*Buffer).grow
         0     0%  2.71%       1629  0.17%  compress/flate.(*Writer).Close (inline)
         0     0%  2.71%       1629  0.17%  compress/flate.(*compressor).close
         0     0%  2.71%       1597  0.17%  compress/flate.(*compressor).deflate
         0     0%  2.71%       1597  0.17%  compress/flate.(*compressor).writeBlock
         0     0%  2.71%     -16384  1.73%  compress/flate.(*decompressor).Read
         0     0%  2.71%     -16384  1.73%  compress/flate.(*decompressor).nextBlock
         0     0%  2.71%     -16384  1.73%  compress/flate.(*decompressor).readHuffman
         0     0%  2.71%       3650  0.39%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0%  2.71%       1597  0.17%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0%  2.71%     -16384  1.73%  compress/gzip.(*Reader).Read
         0     0%  2.71%       -195 0.021%  compress/gzip.(*Reader).Reset
         0     0%  2.71%       -195 0.021%  compress/gzip.(*Reader).readHeader
         0     0%  2.71%       1629  0.17%  compress/gzip.(*Writer).Close
         0     0%  2.71%     -47349  5.00%  compress/gzip.(*Writer).Write
         0     0%  2.71%       -195 0.021%  compress/gzip.NewReader (inline)
         0     0%  2.71%       3277  0.35%  compress/gzip.NewWriter
         0     0%  2.71%     -21846  2.31%  encoding/json.(*Decoder).Decode
         0     0%  2.71%      -2050  0.22%  encoding/json.(*Decoder).readValue
         0     0%  2.71%     -19796  2.09%  encoding/json.(*decodeState).array
         0     0%  2.71%     -19796  2.09%  encoding/json.(*decodeState).unmarshal
         0     0%  2.71%     -19796  2.09%  encoding/json.(*decodeState).value
         0     0%  2.71%      11606  1.23%  encoding/json.(*encodeState).marshal
         0     0%  2.71%      11606  1.23%  encoding/json.(*encodeState).reflectValue
         0     0%  2.71%      19799  2.09%  encoding/json.arrayEncoder.encode
         0     0%  2.71%      19799  2.09%  encoding/json.sliceEncoder.encode
         0     0%  2.71%      19799  2.09%  encoding/json.stringEncoder
         0     0%  2.71%      11606  1.23%  encoding/json.structEncoder.encode
         0     0%  2.71%       5960  0.63%  github.com/gin-gonic/gin.(*Context).AbortWithStatus
         0     0%  2.71%     -32769  3.46%  github.com/gin-gonic/gin.(*Context).Header
         0     0%  2.71%      11120  1.17%  github.com/gin-gonic/gin.(*Context).JSON (inline)
         0     0%  2.71%    -108567 11.46%  github.com/gin-gonic/gin.(*Context).Next
         0     0%  2.71%      19313  2.04%  github.com/gin-gonic/gin.(*Context).Render
         0     0%  2.71%      22635  2.39%  github.com/gin-gonic/gin.(*Context).SetCookie
         0     0%  2.71%     -40053  4.23%  github.com/gin-gonic/gin.(*Context).ShouldBindJSON (inline)
         0     0%  2.71%     -40053  4.23%  github.com/gin-gonic/gin.(*Context).ShouldBindWith (inline)
         0     0%  2.71%     -75798  8.00%  github.com/gin-gonic/gin.(*Engine).ServeHTTP
         0     0%  2.71%     -75799  8.00%  github.com/gin-gonic/gin.(*Engine).handleHTTPRequest
         0     0%  2.71%     -16139  1.70%  github.com/gin-gonic/gin.(*responseWriter).Write
         0     0%  2.71%     -10179  1.07%  github.com/gin-gonic/gin.(*responseWriter).WriteHeaderNow (partial-inline)
         0     0%  2.71%    -108567 11.46%  github.com/gin-gonic/gin.CustomRecoveryWithWriter.func1
         0     0%  2.71%     -18207  1.92%  github.com/gin-gonic/gin/binding.(*defaultValidator).ValidateStruct
         0     0%  2.71%      -1822  0.19%  github.com/gin-gonic/gin/binding.(*defaultValidator).validateStruct
         0     0%  2.71%     -40053  4.23%  github.com/gin-gonic/gin/binding.decodeJSON
         0     0%  2.71%     -40053  4.23%  github.com/gin-gonic/gin/binding.jsonBinding.Bind
         0     0%  2.71%     -18207  1.92%  github.com/gin-gonic/gin/binding.validate (inline)
         0     0%  2.71%       5160  0.54%  github.com/gin-gonic/gin/codec/json.jsonApi.Marshal
         0     0%  2.71%      11120  1.17%  github.com/gin-gonic/gin/render.JSON.Render
         0     0%  2.71%       8193  0.86%  github.com/gin-gonic/gin/render.String.Render
         0     0%  2.71%      11120  1.17%  github.com/gin-gonic/gin/render.WriteJSON
         0     0%  2.71%       8193  0.86%  github.com/gin-gonic/gin/render.WriteString
         0     0%  2.71%      -1822  0.19%  github.com/go-playground/validator/v10.(*Validate).Struct (inline)
         0     0%  2.71%      -1822  0.19%  github.com/go-playground/validator/v10.(*Validate).StructCtx
         0     0%  2.71%      32768  3.46%  github.com/google/uuid.New
         0     0%  2.71%      32768  3.46%  github.com/google/uuid.NewRandom
         0     0%  2.71%    -155780 16.44%  go-musthave-shortener/internal/di/app.(*DI).initMux.GzipMiddleware.func2
         0     0%  2.71%     -58882  6.22%  go-musthave-shortener/internal/di/app.(*DI).initMux.WrapF.func13
         0     0%  2.71%      14153  1.49%  go-musthave-shortener/internal/middleware.(*bodyWriter).Write
         0     0%  2.71%      10923  1.15%  go-musthave-shortener/internal/repository/shorterrepository.(*Repo).Get
         0     0%  2.71%     -36925  3.90%  go-musthave-shortener/internal/usecase/createshortlinkjsonusecase.(*Usecase).Execute
         0     0%  2.71%      11947  1.26%  go-musthave-shortener/internal/usecase/createshortlinkusecase.(*Usecase).Execute
         0     0%  2.71%      -8192  0.86%  internal/profile.(*profileMerger).mapLocation
         0     0%  2.71%     -46842  4.94%  internal/profile.Parse
         0     0%  2.71%     -38176  4.03%  internal/profile.decodeMessage
         0     0%  2.71%     -32715  3.45%  internal/profile.init.func6
         0     0%  2.71%     -30233  3.19%  internal/profile.parseUncompressed
         0     0%  2.71%     -38176  4.03%  internal/profile.unmarshal (inline)
         0     0%  2.71%       -192  0.02%  io.Copy (inline)
         0     0%  2.71%       -192  0.02%  io.CopyN
         0     0%  2.71%       -192  0.02%  io.copyBuffer
         0     0%  2.71%       -192  0.02%  io.discard.ReadFrom
         0     0%  2.71%       3277  0.35%  net/http.(*Cookie).String
         0     0%  2.71%       -961   0.1%  net/http.(*chunkWriter).Write
         0     0%  2.71%       -961   0.1%  net/http.(*chunkWriter).writeHeader
         0     0%  2.71%      32475  3.43%  net/http.(*conn).serve
         0     0%  2.71%      -6554  0.69%  net/http.(*connReader).backgroundRead
         0     0%  2.71%     -10179  1.07%  net/http.(*response).WriteHeader
         0     0%  2.71%       -959   0.1%  net/http.(*response).finishRequest
         0     0%  2.71%      19358  2.04%  net/http.Header.Add (inline)
         0     0%  2.71%     -32769  3.46%  net/http.Header.Set (inline)
         0     0%  2.71%       -769 0.081%  net/http.Header.WriteSubset (inline)
         0     0%  2.71%       -769 0.081%  net/http.Header.sortedKeyValues
         0     0%  2.71%       -769 0.081%  net/http.Header.writeSubset
         0     0%  2.71%      22635  2.39%  net/http.SetCookie
         0     0%  2.71%       -257 0.027%  net/http.newTextprotoReader
         0     0%  2.71%        256 0.027%  net/http.putTextprotoReader
         0     0%  2.71%     -75798  8.00%  net/http.serverHandler.ServeHTTP
         0     0%  2.71%     -56604  5.98%  net/http/pprof.collectProfile
         0     0%  2.71%     -58882  6.22%  net/http/pprof.handler.ServeHTTP
         0     0%  2.71%     -58882  6.22%  net/http/pprof.handler.serveDeltaProfile
         0     0%  2.71%      24476  2.58%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0%  2.71%      54616  5.77%  net/url.(*URL).String
         0     0%  2.71%      21845  2.31%  net/url.(*URL).setPath
         0     0%  2.71%     -27306  2.88%  net/url.JoinPath
         0     0%  2.71%       3640  0.38%  net/url.Parse
         0     0%  2.71%      18203  1.92%  net/url.ParseRequestURI
         0     0%  2.71%      21845  2.31%  net/url.unescape
         0     0%  2.71%     -32768  3.46%  path.Clean
         0     0%  2.71%       2050  0.22%  reflect.Value.Grow
         0     0%  2.71%     -16385  1.73%  reflect.Value.Interface (inline)
         0     0%  2.71%       2050  0.22%  reflect.Value.grow
         0     0%  2.71%     -16385  1.73%  reflect.valueInterface
         0     0%  2.71%      -9762  1.03%  runtime/pprof.(*Profile).WriteTo
         0     0%  2.71%      -9790  1.03%  runtime/pprof.(*profileBuilder).appendLocsForStack
         0     0%  2.71%      -9762  1.03%  runtime/pprof.writeHeap
         0     0%  2.71%      -9762  1.03%  runtime/pprof.writeHeapInternal
         0     0%  2.71%      -9762  1.03%  runtime/pprof.writeHeapProto
         0     0%  2.71%      79738  8.42%  strings.(*Builder).Grow
         0     0%  2.71%      -2784  0.29%  sync.(*Pool).Get
         0     0%  2.71%        258 0.027%  sync.(*Pool).Put
         0     0%  2.71%       -771 0.081%  sync.(*Pool).pin
```
