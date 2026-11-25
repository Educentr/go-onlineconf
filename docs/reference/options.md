# Опции инициализации

Опции передаются в `Initialize()` или `Create()` для настройки экземпляра.

## WithConfigDir

Задаёт директорию с CDB-файлами. По умолчанию — `/usr/local/etc/onlineconf`.

```go
onlineconf.Initialize(ctx,
    onlineconf.WithConfigDir("/opt/myapp/etc"),
)
```

## WithLogger

Подключает пользовательский логгер. Должен реализовывать интерфейс `onlineconfInterface.Logger`:

```go
type Logger interface {
    Warn(ctx context.Context, msg string, args ...any)
    Error(ctx context.Context, msg string, args ...any)
    Fatal(ctx context.Context, msg string, args ...any)
}
```

Пример:

```go
type MyLogger struct{}

func (l *MyLogger) Warn(ctx context.Context, msg string, args ...any) {
    log.Printf("[WARN] %s %v", msg, args)
}
func (l *MyLogger) Error(ctx context.Context, msg string, args ...any) {
    log.Printf("[ERROR] %s %v", msg, args)
}
func (l *MyLogger) Fatal(ctx context.Context, msg string, args ...any) {
    log.Fatalf("[FATAL] %s %v", msg, args)
}

onlineconf.Initialize(ctx, onlineconf.WithLogger(&MyLogger{}))
```

## WithModules

Предзагружает модули при инициализации:

```go
onlineconf.Initialize(ctx,
    onlineconf.WithModules([]string{"TREE", "module1", "module2"}, true),
)
```

Второй аргумент (`required`) определяет, что произойдёт, если модуль открыть не удалось —
файла `<имя>.cdb` нет в каталоге конфигурации или файл непригоден как CDB-база
(`ErrInvalidCDB`: не CDB или оборванная сборка, см. [Модули и CDB](../architecture/modules.md)):

- `required = true` — **panic** при создании экземпляра, то есть внутри `Create()` или
  `Initialize()`. Это не ошибка, которую возвращает `Initialize()`: вызов не возвращается
  вовсе, поэтому перехватить это можно только через `recover()`. Сообщение panic содержит имя
  модуля и текст ошибки открытия.
- `required = false` — ошибка пишется в логгер экземпляра (`Error`) вместе с именем модуля,
  модуль пропускается, инициализация продолжается. Остальные модули из списка открываются как
  обычно.

Пропущенный модуль не зарегистрирован в экземпляре:

- `GetOrAddModule()` (и все функции чтения, которые идут через него) при следующем обращении
  **пробует открыть его снова** — если к этому моменту файл появился или был перезаписан
  целиком, модуль откроется;
- `GetModule()` для такого модуля возвращает `nil` и повторную попытку открытия **не делает**;
- `GetModuleNames()` его не содержит.

!!! note "Указывайте `WithModules` последним"
    Опции применяются в том порядке, в котором переданы, а `WithModules` сразу читает каталог
    конфигурации и логгер экземпляра. Если указать его **до** `WithConfigDir()` или
    `WithLogger()`, модули будут открываться из каталога по умолчанию
    (`/usr/local/etc/onlineconf`), а ошибки уйдут в `DefaultLogger`.

!!! warning "Режим `ONLINECONFIG_FROM_ENV`"
    `Initialize()` генерирует `TREE.cdb` из переменных `OC_*` **после** применения опций,
    поэтому `WithModules([]string{"TREE"}, true)` в этом режиме видит каталог ещё до генерации
    файла — и паникует, если `TREE.cdb` там ещё нет. В этом режиме не перечисляйте `TREE` в
    `WithModules`.
