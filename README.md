# Hand-in 5: Replication

### Welcome to Stanniol's implementation

**Setup instructions**

Once the repository has been cloned and have opened the program, you should be at the root of this repository.
The instructions assume you start at the root.

1. Step is to set up all the nodes the clients are familiar with:
- open a terminal
- `cd server`
- `go run server.go -port <port address>`
- The clients know the following ports: 50051, 50052 and 50053

2. Step is to set up a client
- open a different terminal
- `cd client`
- `go run client.go`
- repeat the amount of clients you want

**Crash instructions**

You can use `control + c` on mac and  `Ctrl + c` on Windows, in a terminal running a server to crash the server

