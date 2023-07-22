default_server_conn_opts : Modifier = RPCServer(framework="grpc")
web_opts : Modifier = WebServer(framework="default")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
cpool_opts : Modifier = ClientPool(max_clients=100)
client_modifiers : List[Modifier] = [cpool_opts]

jaegerTracer : Tracer = JaegerTracer().WithServer(default_deployer)

jaegerTracerModifier : Callable[str, Modifier] = lambda x : TracerModifier(tracer=jaegerTracer, service_name= x, sampling_rate= 1)

server_modifiers : Callable[str, List[Modifier]] = lambda x : [default_server_conn_opts, default_deployer, jaegerTracerModifier(x)]

web_modifiers : Callable[str, List[Modifier]] = lambda x : [web_opts, default_deployer, jaegerTracerModifier(x)]

leafService : Service = LeafServiceImpl().WithServer(server_modifiers("LeafService")).WithClient(client_modifiers)

nonleafService : Service = NonLeafServiceImpl(leafService=leafService).WithServer(server_modifiers("NonLeafService")).WithClient(client_modifiers)

webService : Service  = WebServiceImpl(leafService=leafService).WithServer(web_modifiers("WebService"))
