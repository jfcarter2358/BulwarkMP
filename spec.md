# BulwarkMP

## Frames

### ACCEPT

- Server -> Client0
    - Server consents to data send/receive action
    
### ACKNOWLEDGE

- Client -> Server
    - Client acknowledges server connection
    - Client acknowledges data receipt
- Server -> Client
    - Server acknowledges data receipt

### CLOSE

- Client -> Server
    - Client wants to close connection
- Server -> Client
    - Server accepts closing the connection

### CONNECT

- Client -> Server
    - Client requests to connect
- Server -> Client
    - Server confirms connection

### DATA

- Server -> Client
    - Server is sending back data to client

### ERROR

- Client -> Server
    - Client ran into error reading frame from server
- Server -> Client
    - Server ran into error reading frame from client

### PULL

- Client -> Server
    - Client requests data

### PUSH

- Client -> Server
    - Client requests to send data

### REJECT

- Server -> Client
    - Server is unable to accept request at this time

### RETRY

- Client -> Server
    - Client asks server to send data again
- Server -> Client
    - Server asks Client to send data again

## Frame Format

```
ENDPOINT: queue|buffer
VERSION: v1/secure|v1/plain
KIND: ACCEPT|CLOSE|CONFIRM|CONNECT|PULL|PUSH|REJECT|RETRY
CONTENT-TYPE: JSON|YAML|XML|TEXT|BINARY
CONTENT-LENGTH: int
...
```
