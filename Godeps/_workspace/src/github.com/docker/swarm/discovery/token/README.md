#discovery.hub.docker.com

Docker Swarm comes with a simple discovery service built into the [Docker Hub](http://hub.docker.com)

The discovery service is still in alpha stage and currently hosted at `http://discovery-stage.hub.docker.com`

#####Create a new cluster
`-> POST http://discovery.hub.docker.com/v1/clusters (data="")`

`<- <token>`

#####Add new nodes to a cluster
`-> POST http://discovery.hub.docker.com/v1/clusters/<token> (data="<ip:port1>")`

`<- OK`

`-> POST http://discovery.hub.docker.com/v1/clusters/token (data="<ip:port2>")`

`<- OK`


#####List nodes in a cluster
`-> GET http://discovery.hub.docker.com/v1/clusters/token`

`<- ["<ip:port1>", "<ip:port2>"]`


#####Delete a cluster (all the nodes in a cluster)
`-> DELETE http://discovery.hub.docker.com/v1/clusters/token`

`<- OK`
