default_server_conn_opts : Modifier = WebServer(framework="default")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
server_modifiers : List[Modifier] = [default_server_conn_opts, default_deployer]
leafService : LeafService = LeafServiceImpl()

nonleafService : NonLeafService = NonLeafServiceImpl(leafService=leafService).WithServer(server_modifiers)

process : Process = Process(services=[leafService, nonleafService])
