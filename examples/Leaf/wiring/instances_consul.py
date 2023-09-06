default_server_conn_opts : Modifier = RPCServer(framework="grpc", resolver="consul")
web_opts : Modifier = WebServer(framework="default")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)

jaegerTracer : Tracer = JaegerTracer().WithServer(default_deployer)

consul : Registry = ConsulRegistry().WithServer(default_deployer)

jaegerTracerModifier : Callable[str, Modifier] = lambda x : TracerModifier(tracer=jaegerTracer, service_name= x, sampling_rate= 1)

consulModifier : Callable[str, Modifier] = lambda x : ConsulModifier(reg=consul, service_name=x, service_id=x)

server_modifiers : Callable[str, List[Modifier]] = lambda x : [default_server_conn_opts, default_deployer, jaegerTracerModifier(x), consulModifier(x)]

web_modifiers : Callable[str, List[Modifier]] = lambda x : [web_opts, default_deployer, jaegerTracerModifier(x), consulModifier(x)]

leafService : Service = LeafServiceImpl().WithServer(server_modifiers("leafService"))

nonleafService : Service = NonLeafServiceImpl(leafService=leafService).WithServer(server_modifiers("nonleafService"))

webService : Service  = WebServiceImpl(leafService=leafService).WithServer(web_modifiers("WebService"))
