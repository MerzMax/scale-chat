# scale-chat

The Idea is to create a simple chat application and a correspoinding client that grow over time. This means that the 
server might run in several problems regarding scale that need to be fixed. Which problems will occur we are going to 
see if we go along.

There is a blog article available that describes what the idea was and how we went along: 
[Blog Entry](https://blog.mi.hdm-stuttgart.de)

### How to run the applications: 

#### Server and Monitoring
There is a docker-compose file available that can be used to deploy the setup in a Docker Swarm cluster. Documentation 
on how to deploy a docker-compose file can be found 
[here](https://docs.docker.com/engine/swarm/stack-deploy/#deploy-the-stack-to-the-swarm).

When deployed the chat server will run on port 80. The Grafana dashboards will be accessible at port 3000. There are 
two dashboards available: 
* cAdvisor Exporter - contains hardware statistics
* Go Processes - contains the services custom metrics & metrics provided by the prometheus GO client

#### Clients

There is a demo client available if you just requests port 80 with a browser. Apart from this demo client you can start 
the load-test-client (`/src/load-test-client`) without any arguments. This will run a simple terminal client that can be 
used to chat. if you navigate to `/src/load-test-client` you can start the client with the following command: 

```bash
go run . 
```

If you want to make a load test you can start the load-test-client with the `--load-test` flag. For a list of all 
configurations execute the following command in the `/src/load-test-client` folder:
```bash
go run . --help 
```

##### Client monitoring 
If the client is getting started in the load-test mode a csv file containing message events will be created. In order 
to plot the results the events have to be processed. Therefore, a `csv-processor` was implemented. You can run the 
processor if you are in the `/src/load-test-client/csv-processor` directory with the following command: 
```bash
go run . 
```
In order to display the configuration possibilities execute the following command: 
```bash
go run . --help 
```

In the end the results can be plotted with Matplotlib, a Python library. 

TODO

