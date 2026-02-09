# Переменные окружения

## ONLINECONFIG_FROM_ENV

Когда установлена (любое непустое значение), `Initialize()` генерирует CDB-файл из переменных окружения с префиксом `OC_`.

```bash
export ONLINECONFIG_FROM_ENV=1
export OC_app__name=myservice
export OC_app__port=8080
export OC_app__nested__deep__param=value
```

### Преобразование имён

Двойное подчёркивание (`__`) преобразуется в разделитель пути (`/`), добавляется ведущий `/`:

| Переменная окружения | Путь в конфигурации |
|---|---|
| `OC_app__name` | `/app/name` |
| `OC_app__port` | `/app/port` |
| `OC_app__nested__deep__param` | `/app/nested/deep/param` |

### Применение

Удобно для:

- Локальной разработки без `onlineconf-updater`
- Тестов в CI/CD
- Docker-контейнеров

```bash
# Docker
docker run -e ONLINECONFIG_FROM_ENV=1 -e OC_app__port=9090 myapp
```

!!! note "Модуль"
    Переменные окружения записываются в модуль `TREE` по умолчанию.
