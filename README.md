# Callback retranslator

Listens requests and transmits it to client.
Client responses will be transmitted back.

## Release
Run command:
```
make release
```

## Server

Params:
* -port - port for clients (default: 8029)

Command:
```
release/server-linux
```

## Client
Listens port and responds information about request.

Params:
* -serverAddr - Retranslator server address. Example: 127.0.0.1:8029 (default "127.0.0.1:8029")
* -port - Port for requests. Retranslator will listen this port and forward requests to client. Default *8080*.
* -targetAddr - Target address for requests transmitting. Example: 127.0.0.1:80/callback (default *localhost*)
* -forwardUri - Flag to apply requested uri (path and query string) to targetAddr. Default *true*.

Command:
```
release/client-linux -serverAddr 192.168.1.1:8029 -port 8080 -targetAddr http://localhost/myCallback
```