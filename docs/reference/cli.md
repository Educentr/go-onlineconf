# CLI-утилита

Утилита командной строки для генерации CDB-файлов и чтения конфигурации.

## Запуск

```bash
go run github.com/Educentr/go-onlineconf/cmd/cli [флаги]
```

## Флаги

| Флаг | По умолчанию | Описание |
|---|---|---|
| `-command` | `generate` | Команда: `generate` или `get` |
| `-dir` | `/usr/local/etc/onlineconf` | Директория CDB-файлов |
| `-module` | `TREE` | Имя модуля |
| `-yaml` | `./configs/onlineconf.yml` | Путь к YAML-файлу (для `generate`) |
| `-config-name` | — | Имя параметра (для `get`) |
| `-help` | `false` | Показать справку |

## Команды

### generate

Генерирует CDB-файл из YAML:

```bash
go run ./cmd/cli \
    -command generate \
    -yaml ./configs/onlineconf.yml \
    -dir /tmp/onlineconf \
    -module TREE
```

Формат YAML:

```yaml
/app/name: "myservice"
/app/port: "8080"
/app/config: '{"workers": 4}'
```

### get

Читает значение параметра из CDB:

```bash
go run ./cmd/cli \
    -command get \
    -config-name "/app/name" \
    -dir /tmp/onlineconf \
    -module TREE
```

Вывод:

```
Value of config parameter /app/name is myservice
```
