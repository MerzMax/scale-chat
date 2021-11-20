# Scale-Chat

The Idea is to create a simple chat application and a correspoinding client that grow over time. This means that there 
will be more and more clients so that the server might run in several problems regarding scale that need to be fixed. 
Which exact problems will occur, we are going to see if we go along.

## Features

TODO

## Repository structure

### Documentation
TODO


### Implementation 
The projects implementation lives in the [./src](./src) folder and contains several directories: 

* `chat` : Contains models that will be shared with the server and the client.

* `client` : Contains the implementation of the client

* `loadtest-client` : Contains the implementation of the loadtest-client that is capable of starting several client 
instances and running in a manual mode. Look [here](#client) for more information. 

* `server` : Contains the implementation of the chat server.

The server and the client are implemented in GO with the help of 
[gorilla/websockets](https://github.com/gorilla/websocket).


## Startup

### Server

The server can be stared by executing the following command in the `./src/server` directory: 

```bash
$ go run .
```

Besides this there is a Dockerfile at `./src/server` that can be used to start the server as well. Be aware of that the 
context of the docker build should be the project's root directory. 

In order to start the server more easily and with a more consistent configuration use the `docker-compose.yaml` 
configuration that is located in the project's root directory. Start the server by executing: 

```bash
$ docker-compose up
```

### Client

#### Web / demo client

There is a webbased client available on [http://localhost:8080](http://localhost:8080) that can be used for demo 
purposes.

#### Loadtest client

To execute the loadtests we implemented a go loadtest client that is capable starting several clients that send 
messages to the server automatically. To configure the loadtest client, just use the following flags on startup:

```bash
$ ./loadtest-client --help
Usage of ./loadtest-client:
  -clients int
        Number of clients that will be started (just for loadtest mode) (default 1)
  -loadtest
        Flag indicates weather the client should start in the loadtest mode
  -msg-frequency int
        The frequency of the messages in ms (just for loadtest mode) (default 1000)
  -msg-size int
        The size of the messages in bytes (just for loadtest mode) (default 256)
  -server-url string
        The url of the server to connect to (default "ws://localhost:8080/ws")
```

The loadtest client can be stared by executing the following command in the `./src/loadtest-client` directory: 

```bash
$ go run . --loadtest-client --clients 5 --msg-frequency 1000 --msg-size 512
```

If you just want to start a simple commandline client that can send user input, start the loadtest-client without any 
flags:

```bash
$ go run .
```


## Running Loadtests

TODO



----

---



## First steps:
* Create a simple chat server that broadcastst received messages to all connected clients &rarr; will be implemented in 
GO
* Create a corresponding client that publishes messages &rarr; will be implemented in GO
* Simulate a lot of clients with the help of a load test and see how the server behaves

## Questions that have to be answered: 
* Communication protocol of clients and server 
    * Is there some sort of handshake?
    * What is the content of a chat message?
    * Plobal chat or private chats?
* How to create an reproducable environment for the server?
    * An idea could be a docker container with restricted power to run into problems quiet fast
* Private chats as well?
* How to scale the clients? 
    * An idea could be docker containers; BUT this might lead into problems on a local mashine (starting 1000 dc 
    could result in performance bottlenecks)
    * Start several processes in a docker contianer? &rarr; would be a break in the concept of docker
* How to measure the performance? What could be significant metrics? Is this even necessary?

### Communication Protocol
> What are the messages that will be send by the server and the client?

New chat message (global chat):  
```JSON
{
    "clientId": "string",
    "Message": "string"
}
```

### Loadtests
> How to simulate the chat clients and how to measure the server?

Metrics that could be usefull: 
* Number of concurrent client connections &rarr; be aware of the OS limit for a process / mashine
* Number of chat messages per second 
* Hardware stats (cpu/ram)
* Messagesize (variation of message size could lead to different problems)

How to measure metrics: 
* Prometheus for docker hardware stats (there is an adapter available)
* Grafana for displaying the results
* Log messages of the server that log messages like the number of connections etc.

Server Setup:
* The server should be contained in a docker-container 
* The containers resouces should be limited (this makes it easier to run into scaling issues)

Client Setup: 
* To simplify the start of the clients in a reproducable environment the client could run in a docker container as well.
  
    There are several posibilities, how the clients clould be started with this approach:
    1. Start several chat clients within one process
        * 1 process = 1 docker contianer
    2. Start several chat client processes
        * x client processes = 1 docker container %rarr; break in docker concept! 
* The clients could be started with the help of metadata that will be configured with the help of 
environment variables 