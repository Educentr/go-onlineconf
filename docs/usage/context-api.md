# Context API

Основной рекомендуемый способ работы — через `context.Context`.

## Инициализация

```go
ctx, err := onlineconf.Initialize(context.Background())
if err != nil {
    log.Fatal(err)
}
```

С опциями:

```go
ctx, err := onlineconf.Initialize(
    context.Background(),
    onlineconf.WithConfigDir("/opt/myapp/etc"),
    onlineconf.WithLogger(&MyLogger{}),
    onlineconf.WithModules([]string{"TREE", "module1"}, true),
)
```

## Чтение значений

Все функции извлекают экземпляр из контекста и читают из модуля `TREE`:

```go
// С обязательным значением
name, err := onlineconf.GetString(ctx, "/app/name")

// С значением по умолчанию
port, err := onlineconf.GetInt(ctx, "/app/port", 8080)

// Проверка существования
val, exists, err := onlineconf.GetStringIfExists(ctx, "/app/optional")
```

## Работа с модулями

```go
module, err := onlineconf.GetOrAddModule(ctx, "module1")
if err != nil {
    log.Fatal(err)
}

val, err := module.GetString("/param")
```

## Watcher

```go
// Запуск
err := onlineconf.StartWatcher(ctx)

// Подписка на изменения конкретного параметра.
// Callback вызывается только если значение /app/flag реально изменилось.
onlineconf.RegisterSubscription(ctx, "TREE", []string{"/app/flag"}, func() error {
    // Можно безопасно читать конфигурацию внутри callback
    val, _ := onlineconf.GetString(ctx, "/app/flag")
    log.Println("flag changed to", val)
    return nil
})

// Подписка на несколько путей — callback вызывается максимум один раз,
// если хотя бы один из путей изменился.
onlineconf.RegisterSubscription(ctx, "TREE", []string{"/app/rate", "/app/limit"}, func() error {
    log.Println("rate or limit changed")
    return nil
})

// Остановка
onlineconf.StopWatcher(ctx)
```

## Clone/Release

```go
reqCtx, err := onlineconf.Clone(globalCtx, requestCtx)
defer onlineconf.Release(globalCtx, reqCtx)

val, _ := onlineconf.GetString(reqCtx, "/app/name")
```
