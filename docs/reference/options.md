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

Второй аргумент (`required`) — если `true`, инициализация завершится ошибкой при отсутствии модуля.
