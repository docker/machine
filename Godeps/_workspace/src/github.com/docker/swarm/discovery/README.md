Discovery
=========

`Docker Swarm` comes with multiple Discovery backends

## Examples

##### Using the hosted discovery service

```bash
# create a cluster
$ swarm create
6856663cdefdec325839a4b7e1de38e8 # <- this is your unique <cluster_id>

# on each of your nodes, start the swarm agent
#  <node_ip> doesn't have to be public (eg. 192.168.0.X),
#  as long as the other nodes can reach it, it is fine.
$ swarm join --addr=<node_ip:2375> token://<cluster_id>

# start the manager on any machine or your laptop
$ swarm manage -H tcp://<swarm_ip:swarm_port> token://<cluster_id>

# use the regular docker cli
$ docker -H tcp://<swarm_ip:swarm_port> info
$ docker -H tcp://<swarm_ip:swarm_port> run ...
$ docker -H tcp://<swarm_ip:swarm_port> ps
$ docker -H tcp://<swarm_ip:swarm_port> logs ...
...

# list nodes in your cluster
$ swarm list token://<cluster_id>
<node_ip:2375>
```

###### Using a static file describing the cluster

```bash
# for each of your nodes, add a line to a file
#  <node_ip> doesn't have to be public (eg. 192.168.0.X),
#  as long as the other nodes can reach it, it is fine.
$ echo <node_ip1:2375> >> /tmp/my_cluster
$ echo <node_ip2:2375> >> /tmp/my_cluster
$ echo <node_ip3:2375> >> /tmp/my_cluster

# start the manager on any machine or your laptop
$ swarm manage -H tcp://<swarm_ip:swarm_port> file:///tmp/my_cluster

# use the regular docker cli
$ docker -H tcp://<swarm_ip:swarm_port> info
$ docker -H tcp://<swarm_ip:swarm_port> run ...
$ docker -H tcp://<swarm_ip:swarm_port> ps
$ docker -H tcp://<swarm_ip:swarm_port> logs ...
...

# list nodes in your cluster
$ swarm list file:///tmp/my_cluster
<node_ip1:2375>
<node_ip2:2375>
<node_ip3:2375>
```

###### Using etcd

```bash
# on each of your nodes, start the swarm agent
#  <node_ip> doesn't have to be public (eg. 192.168.0.X),
#  as long as the other nodes can reach it, it is fine.
$ swarm join --addr=<node_ip:2375> etcd://<etcd_ip>/<path>

# start the manager on any machine or your laptop
$ swarm manage -H tcp://<swarm_ip:swarm_port> etcd://<etcd_ip>/<path>

# use the regular docker cli
$ docker -H tcp://<swarm_ip:swarm_port> info
$ docker -H tcp://<swarm_ip:swarm_port> run ...
$ docker -H tcp://<swarm_ip:swarm_port> ps
$ docker -H tcp://<swarm_ip:swarm_port> logs ...
...

# list nodes in your cluster
$ swarm list etcd://<etcd_ip>/<path>
<node_ip:2375>
```

###### Using consul

```bash
# on each of your nodes, start the swarm agent
#  <node_ip> doesn't have to be public (eg. 192.168.0.X),
#  as long as the other nodes can reach it, it is fine.
$ swarm join --addr=<node_ip:2375> consul://<consul_addr>/<path>

# start the manager on any machine or your laptop
$ swarm manage -H tcp://<swarm_ip:swarm_port> consul://<consul_addr>/<path>

# use the regular docker cli
$ docker -H tcp://<swarm_ip:swarm_port> info
$ docker -H tcp://<swarm_ip:swarm_port> run ...
$ docker -H tcp://<swarm_ip:swarm_port> ps
$ docker -H tcp://<swarm_ip:swarm_port> logs ...
...

# list nodes in your cluster
$ swarm list consul://<consul_addr>/<path>
<node_ip:2375>
```

###### Using zookeeper

```bash
# on each of your nodes, start the swarm agent
#  <node_ip> doesn't have to be public (eg. 192.168.0.X),
#  as long as the other nodes can reach it, it is fine.
$ swarm join --addr=<node_ip:2375> zk://<zookeeper_addr1>,<zookeeper_addr2>/<path>

# start the manager on any machine or your laptop
$ swarm manage -H tcp://<swarm_ip:swarm_port> zk://<zookeeper_addr1>,<zookeeper_addr2>/<path>

# use the regular docker cli
$ docker -H tcp://<swarm_ip:swarm_port> info
$ docker -H tcp://<swarm_ip:swarm_port> run ...
$ docker -H tcp://<swarm_ip:swarm_port> ps
$ docker -H tcp://<swarm_ip:swarm_port> logs ...
...

# list nodes in your cluster
$ swarm list zk://<zookeeper_addr1>,<zookeeper_addr2>/<path>
<node_ip:2375>
```

###### Using a static list of ips

```bash
# start the manager on any machine or your laptop
$ swarm manage -H <swarm_ip:swarm_port> nodes://<node_ip1:2375>,<node_ip2:2375>
#or
$ swarm manage -H <swarm_ip:swarm_port> nodes://<node_ip1:2375>,<node_ip2:2375>

# use the regular docker cli
$ docker -H <swarm_ip:swarm_port> info
$ docker -H <swarm_ip:swarm_port> run ...
$ docker -H <swarm_ip:swarm_port> ps
$ docker -H <swarm_ip:swarm_port> logs ...
...
```

## Contributing

Contributing a new discovery backend is easy,
simply implements this interface:

```go
type DiscoveryService interface {
     Initialize(string, int) error
     Fetch() ([]string, error)
     Watch(WatchCallback)
     Register(string) error
}
```

######Initialize
take the `discovery` without the scheme and a heartbeat (in seconds)

######Fetch
returns the list of all the nodes from the discovery

######Watch
triggers an update (`Fetch`),it can happen either via
a timer (like `token`) or use backend specific features (like `etcd`)

######Register
add a new node to the discovery
