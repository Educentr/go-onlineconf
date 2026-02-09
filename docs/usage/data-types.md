# Типы данных

## Поддерживаемые типы

| Метод | Go-тип | CDB-формат | Описание |
|---|---|---|---|
| `GetString` | `string` | `s` | UTF-8 строка |
| `GetInt` | `int64` | `s` | Строка, парсится через `strconv.ParseInt` |
| `GetBool` | `bool` | `s` | `"0"` или пустая строка = `false`, остальное = `true` |
| `GetDuration` | `time.Duration` | `s` | Строка, парсится через `time.ParseDuration` (`"5s"`, `"100ms"`) |
| `GetStrings` | `[]string` | `s` или `j` | Строка через запятую или JSON-массив |
| `GetStruct` | `interface{}` | `j` | JSON, десериализуется через `json.Unmarshal` |

## Два варианта методов

### Get* — с значением по умолчанию

Возвращает значение или default. Ошибка если ключ не найден и default не задан.

```go
// Без default — ошибка если ключ не существует
name, err := module.GetString("/app/name")

// С default — вернёт 8080 если ключ не существует
port, err := module.GetInt("/app/port", 8080)
```

### Get*IfExists — с проверкой существования

Возвращает значение, флаг существования и ошибку.

```go
val, exists, err := module.GetStringIfExists("/app/optional")
if err != nil {
    // Ошибка чтения
}
if !exists {
    // Параметр не существует
}
```

## GetStrings

Поддерживает два формата:

```yaml
# Формат s: строка через запятую
/app/hosts: "host1, host2, host3"

# Формат j: JSON-массив
/app/hosts: '["host1", "host2", "host3"]'
```

```go
hosts, err := module.GetStrings("/app/hosts", []string{"localhost"})
```

## GetStruct

Десериализует JSON в произвольную структуру. Результат кэшируется.

```yaml
# В CDB (формат j)
/app/config: '{"workers": 4, "buffer_size": 1024}'
```

```go
type AppConfig struct {
    Workers    int `json:"workers"`
    BufferSize int `json:"buffer_size"`
}

var cfg AppConfig
exists, err := module.GetStruct("/app/config", &cfg)
```

!!! warning "Кэширование"
    `GetStruct` и `GetStrings` кэшируют десериализованные значения. Не модифицируйте возвращённые по ссылке значения — это изменит кэш.
