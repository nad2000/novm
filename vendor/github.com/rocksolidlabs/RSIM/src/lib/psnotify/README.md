# Process notifications for Go

## Overview

The psnotify package captures process events from the kernel via
kqueue on Darwin/BSD and the netlink connector on Linux.

Example:
```go
    watcher, err := psnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    // Process events
    go func() {
        for {
            select {
            case ev := <-watcher.Fork:
                log.Println("fork event:", ev)
            case ev := <-watcher.Exec:
                log.Println("exec event:", ev)
            case ev := <-watcher.Exit:
                log.Println("exit event:", ev)
            case err := <-watcher.Error:
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Watch(os.Getpid(), psnotify.PROC_EVENT_ALL)
    if err != nil {
        log.Fatal(err)
    }

    /* ... do stuff ... */
    watcher.Close()
```
