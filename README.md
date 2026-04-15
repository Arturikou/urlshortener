# Сервис сокращения URL


## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
git fetch template && git checkout template/v2 .github
```

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента.

---

## Профилирование и оптимизация памяти


### 1. Устранение Heap Escape в методах GetURL и AddURL

**До:**

| Бенчмарк | allocs/op | B/op |
|---|---|---|
| `BenchmarkGetURL` | 1 | 16 |
| `BenchmarkAddURL` | 5 | 160 |

**`pprof -top`:**
```
Showing nodes accounting for 7.90GB, 100% of 7.90GB total
   3.90GB 49.33%  (*Shortener).GetURL
   0.70GB  8.92%  encoding/base64.(*Encoding).EncodeToString
   0.33GB  4.19%  (*Shortener).AddURL
```
**Проблема:** Использование указателя *uuid.UUID приводило к Heap Escape  
**Решение:** Передача переменной по значению

**После:**

| Бенчмарк | allocs/op | B/op |
|---|-----------|------|
| `BenchmarkGetURL` | 0         | 0    |
| `BenchmarkAddURL` | 4         | 144  |

**`pprof -top -diff_base`:**
```
Showing nodes accounting for -3.49GB, 44.17% of 7.90GB total
   -3.90GB 49.33%  (*Shortener).GetURL
    0.70GB  8.92%  encoding/base64.(*Encoding).EncodeToString
   -0.33GB  4.19%  (*Shortener).AddURL
```

**Итог: −3.49GB (−44%)**

---

### 2. Анализ функции `generateAlias`

**До:**

| Бенчмарк | allocs/op | B/op |
|---|---|---|
| `BenchmarkGenerateAlias` | 1 | 24 |

**Вывод**: оптимизация не требуется `make([]byte, 16)` остаётся на стеке. Единственная аллокация — `base64.EncodeToString`.

---


### 3. Оптимизация метода `AddURLs` — антипаттерн `[]*ShortenBatch`

**Инструмент:** бенчмарк (`-memprofile profiles/base_urls.pprof`)

**До:**

| Бенчмарк           | allocs/op | B/op |
|--------------------|-----------|------|
| `BenchmarkAddURLs` | 101       | 7264 |

**`pprof -top`:**
```
Showing nodes accounting for 1466.05MB, 99.80% of 1469.05MB total
  980.53MB 66.75%  (*Shortener).AddURLs
  485.51MB 33.05%  encoding/base64.(*Encoding).EncodeToString
```

**Проблема:** Использование среза указателей []*ShortenBatch приводило к аллокациям при разыменовании объектов  
**Решение:** Использование среза значений


**После:**


| Бенчмарк           | allocs/op | B/op |
|--------------------|-----------|------|
| `BenchmarkAddURLs` | 100       | 2400 |

**`pprof -top -diff_base`:**
```
Showing nodes accounting for -0.97GB, 67.73% of 1.43GB total
   -0.96GB 66.75%  (*Shortener).AddURLs
   -0.01GB  0.99%  encoding/base64.(*Encoding).EncodeToString
```

**Итог: −0.97GB (−68%)**

---

### 4. Оптимизация Gzip middleware

**Инструмент:** `hey` (80k запросов, 100 goroutines) + `pprof` через HTTP

**До:**

| Метрика      | Значение |
|--------------|----------|
| Requests/sec | 15 561   |

**`pprof -top`:**
```
Showing nodes accounting for 61977.93MB, 97.75% of 63404.04MB total
50696.43MB 79.96%  compress/flate.NewWriter
10722.76MB 16.91%  compress/flate.(*compressor).initDeflate
```
**Проблема:** `gzip.NewWriter` вызывался на каждый запрос с `application/json` — внутренние буферы `flate.Writer` (~750KB) выделялись заново при каждом вызове  
**Решение:** `sync.Pool` для переиспользования `gzip.Writer` между запросами

**После:**

| Метрика      | Значение      |
|--------------|---------------|
| Requests/sec | **84 564**    |

**`pprof -top -diff_base`:**
```
Showing nodes accounting for -61684.88MB, 97.29% of 63404.04MB total
-50463.73MB 79.59%  compress/flate.NewWriter
-10669.92MB 16.83%  compress/flate.(*compressor).initDeflate
  -547.20MB  0.86%  compress/flate.(*huffmanEncoder).generate
```

**Итог: −61.6GB (−97%), пропускная способность ×5.4**

---
