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
* -path - path for requests (default "/api")
* -port - port for requests
* -serverAddr - Server address. Example: 127.0.0.1:8080 (default "127.0.0.1:8080")
* -targetAddr - Target address for requests transmitting. Example: 127.0.0.1:80/callback

Command:
```
release/client-linux -serverAddr 192.168.1.1:8029 -path /myCallback -port 9001 -targetAddr http://localhost/myCallback
```