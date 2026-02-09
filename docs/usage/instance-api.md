# Instance API

Альтернативный способ работы — напрямую с экземпляром `OnlineconfInstance`.

## Создание экземпляра

```go
oc := onlineconf.Create()
```

С опциями:

```go
oc := onlineconf.Create(
    onlineconf.WithConfigDir("/opt/myapp/etc"),
    onlineconf.WithLogger(&MyLogger{}),
)
```

## Чтение значений

Методы экземпляра работают с модулем `TREE` по умолчанию:

```go
name, err := oc.GetString("/app/name")
port, err := oc.GetInt("/app/port", 8080)
debug, err := oc.GetBool("/app/debug", false)
timeout, err := oc.GetDuration("/app/timeout", 5*time.Second)

val, exists, err := oc.GetStringIfExists("/app/optional")
```

## Работа с модулями

```go
module, err := oc.GetOrAddModule("module1")
val, err := module.GetString("/param")
```

## Watcher

```go
err := oc.StartWatcher(ctx)
defer oc.StopWatcher()

oc.RegisterSubscription("TREE", []string{"/app/flag"}, func() error {
    log.Println("flag changed")
    return nil
})
```

## Clone/Release

```go
cloned, err := oc.Clone()
defer oc.Release(ctx, cloned)

val, _ := cloned.GetString("/app/name")
```
