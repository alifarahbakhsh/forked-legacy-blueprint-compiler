default_server_conn_opts : Modifier = RPCServer(framework="grpc", timeout="1s")
web_modifier : Modifier = WebServer(framework="default")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
cpool_opts : Modifier = ClientPool(max_clients=100)
retry_opts : Modifier = Retry(max_retries=5)
client_modifiers : List[Modifier] = [cpool_opts, retry_opts]
health_check_modifier : Modifier = HealthChecker()

jaegerTracer : Tracer = JaegerTracer().WithServer(default_deployer)

xtracer : XTracer = XTracerImpl().WithServer(default_deployer)

jaegerTraceModifier : Callable[str, Modifier] = lambda x : TracerModifier(tracer=jaegerTracer, service_name= x, sampling_rate= 1)

server_modifiers : Callable[str, List[Modifier]] = lambda x : [jaegerTraceModifier(x), default_server_conn_opts, default_deployer]

web_server_modifiers : Callable[str, List[Modifier]] = lambda x : [jaegerTraceModifier(x), health_check_modifier, web_modifier, default_deployer]

leafService : LeafService = LeafServiceImpl().WithServer(server_modifiers("LeafService")).WithClient(client_modifiers) 

nonleafService : NonLeafService = NonLeafServiceImpl(leafService=leafService).WithServer(server_modifiers("NonLeafService")).WithClient(client_modifiers)

webService : WebService = WebServiceImpl(leafService=leafService).WithServer(web_server_modifiers("WebService"))