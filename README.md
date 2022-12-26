# tcp-repeater
Service which listens arbitrary ports, forward all data to same ports of certain destination and copies all data to any number of destinations. Also, a control server is available to manage the service and a test mode for debugging on the local machine.

## Build
```shell
$ git clone https://github.com/loolzaaa/tcp-repeater.git
$ cd tcp-repeater
$ go build -o bin/ ./cmd/repeater
$ go build -o bin/ ./tools/...
```

## Usage
Run service command:  
```
./repeater -d destination.com -p 3000 -p 3001 -r repeater.com:8000-3000 -r repeater.com:8001-3001
```
This command runs repeater service which listens `3000` and `3001` ports, forward all data to `destination.com:3000` and `destination.com:3001` and copies it for one repeater by *pattern* (see next).

### Repeater pattern
It is possible to specify any number of repeaters, just follow the following pattern:  
```
<host>:<port>-<original-port>
```
**host** - repeater host  
**port** - repeater port  
**original-port** - from which original port copy data to repeater port

### Additional options
If you need to set timeouts for connecting to a destination or repeaters, then specify the `-td` and `-tr` options at startup, respectively. Available format: 2ms, 4s, etc.  
For help, just run `./repeater --help`

## Control server
A control server is available to manage the service immediately after startup. By default it runs on port `6400`, but this can be configured with the `-cp <port>` option at startup.

Available commands:
- **shutdown** - stop whole repeater service
- **refresh** - close all repeater connections and try to reconnect each

## Test mode
If you start the service with the `--test-port <port>` option, then all data will be redirected to the port specified in the option, regardless of the number of listening ports. This allows you to run the service for debugging on the local machine.
