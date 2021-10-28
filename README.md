# scale-chat

The Idea is to create a simple chat application and a correspoinding client that grow over time. This means that the 
server might run in several problems regarding scale that need to be fixed. Which problems will occur we are going to 
see if we go along.

#### First steps:
* Create a simple chat server that broadcastst received messages to all connected clients &rarr; will be implemented in 
GO
* Create a corresponding client that publishes messages &rarr; will be implemented in GO
* Simulate a lot of clients with the help of a load test and see how the server behaves
 
#### Questions that have to be answered: 
* Communication protocol of clients and server 
    * Is there some sort of handshake?
    * What is the content of a chat message?
* How to create an reproducable environment for the server?
    * An idea could be a docker container with restricted power to run into problems quiet fast
* Private chats as well?
* How to scale the clients? 
    * An idea could be docker containers; BUT this might lead into problems on a local mashine (starting 1000 dc 
    could result in performance bottlenecks)
    * Start several processes in a docker contianer? %rarr; would be a break in the concept of docker
* How to measure the performance? What could be significant metrics? Is this even necessary?