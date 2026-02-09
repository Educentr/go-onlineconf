# Установка

## Требования

- Go 1.21.5+

## Подключение

```bash
go get github.com/Educentr/go-onlineconf
```

## Конфигурационные файлы

По умолчанию `go-onlineconf` ищет CDB-файлы в директории `/usr/local/etc/onlineconf`. Путь можно изменить через опцию `WithConfigDir`.

CDB-файлы генерируются [onlineconf-updater](https://github.com/onlineconf/onlineconf#onlineconf-updater) или CLI-утилитой из YAML.

## CLI-утилита

Для генерации CDB из YAML (удобно для разработки и тестов):

```bash
go run github.com/Educentr/go-onlineconf/cmd/cli \
    -command generate \
    -yaml ./configs/onlineconf.yml \
    -dir /usr/local/etc/onlineconf \
    -module TREE
```

## Проверка

```go
package main

import (
    "context"
    "fmt"

    "github.com/Educentr/go-onlineconf/pkg/onlineconf"
)

func main() {
    ctx, err := onlineconf.Initialize(context.Background())
    if err != nil {
        panic(err)
    }

    val, ex, err := onlineconf.GetStringIfExists(ctx, "/test/param")
    fmt.Printf("value=%s exists=%v err=%v\n", val, ex, err)
}
```
