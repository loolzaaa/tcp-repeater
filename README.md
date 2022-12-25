# tcp-repeater

```shell
> cd path/to/project/cmd/repeater
> go build repeater.go
> ./repeater.exe --test-port 4000 -d localhost -p 3000 -r localhost:6000-3000
```

### Repeater pattern

`<host>:<port>-<original-port>`  
host - repeater host  
port - repeater port  
original-port - from which port forward to repeater port

## Control service

Control server runs on 6400 port by default.

Available commands:
- shutdown - stop whole repeater service
- refresh - close all repeater connections and try to reconnect each