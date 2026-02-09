# Быстрый старт

Минимальный рабочий пример за 5 минут.

## 1. Создайте YAML-конфигурацию

```yaml
# configs/onlineconf.yml
/app/name: "myservice"
/app/port: "8080"
/app/debug: "1"
/app/timeout: "5s"
/app/hosts: '["host1", "host2"]'
```

## 2. Сгенерируйте CDB

```bash
mkdir -p /tmp/onlineconf

go run github.com/Educentr/go-onlineconf/cmd/cli \
    -command generate \
    -yaml ./configs/onlineconf.yml \
    -dir /tmp/onlineconf \
    -module TREE
```

## 3. Используйте в коде

### Context API

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Educentr/go-onlineconf/pkg/onlineconf"
)

func main() {
    ctx, err := onlineconf.Initialize(
        context.Background(),
        onlineconf.WithConfigDir("/tmp/onlineconf"),
    )
    if err != nil {
        panic(err)
    }

    name, _ := onlineconf.GetString(ctx, "/app/name")
    port, _ := onlineconf.GetInt(ctx, "/app/port", 8080)
    debug, _ := onlineconf.GetBool(ctx, "/app/debug", false)
    timeout, _ := onlineconf.GetDuration(ctx, "/app/timeout", 5*time.Second)

    fmt.Printf("name=%s port=%d debug=%v timeout=%s\n", name, port, debug, timeout)
}
```

### Instance API

```go
package main

import (
    "fmt"

    "github.com/Educentr/go-onlineconf/pkg/onlineconf"
)

func main() {
    oc := onlineconf.Create(
        onlineconf.WithConfigDir("/tmp/onlineconf"),
    )

    name, _ := oc.GetString("/app/name")
    port, _ := oc.GetInt("/app/port", 8080)

    fmt.Printf("name=%s port=%d\n", name, port)
}
```

## 4. Инициализация через переменные окружения

Для тестирования без CDB-файлов:

```bash
export ONLINECONFIG_FROM_ENV=1
export OC_app__name=myservice
export OC_app__port=8080

go run main.go
```

Переменные с префиксом `OC_` преобразуются в пути: `OC_app__name` → `/app/name`.
