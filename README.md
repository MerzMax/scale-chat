# scale-chat

The Idea is to create a simple chat application and a correspoinding client that grow over time. This means that the 
server might run in several problems regarding scale that need to be fixed. Which problems will occur we are going to 
see if we go along.

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
    * Start several processes in a docker contianer? %rarr; would be a break in the concept of docker
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
* Number of concurrent client connections 
* Number of chat messages per second 
* Hardware stats (cpu/ram)

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