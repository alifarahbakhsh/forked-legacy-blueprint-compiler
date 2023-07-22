default_server_conn_opts : Modifier = RPCServer(framework="grpc", timeout="1s")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
noop_deployer : Modifier = Deployer(framework="noop")
cpool_opts : Modifier = ClientPool(max_clients=100)
client_modifiers : List[Modifier] = [cpool_opts]

server_modifiers : Callable[str, List[Modifier]] = lambda x : [default_server_conn_opts, default_deployer]
lb_modifiers : List[Modifier] = [noop_deployer]

leafService : Service = LeafServiceImpl().WithServer(server_modifiers("LeafService")).WithClient(client_modifiers)

leafServiceReplica : Service = LeafServiceImpl().WithServer(server_modifiers("LeafServiceReplica")).WithClient(client_modifiers)

loadBalancerLeafService : LoadBalancer = LoadBalancer(clients=[leafService, leafServiceReplica], basetype="LeafService").WithServer(lb_modifiers)

nonleafService : Service = NonLeafServiceImpl(leafService=loadBalancerLeafService).WithServer(server_modifiers("NonLeafService")).WithClient(client_modifiers)