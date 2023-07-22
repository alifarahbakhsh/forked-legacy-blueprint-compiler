default_server_conn_opts : Modifier = RPCServer(framework="grpc", timeout="1s")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
cpool_opts : Modifier = ClientPool(max_clients=100)
retry_opts : Modifier = Retry(max_retries=5)
client_modifiers : List[Modifier] = [cpool_opts, retry_opts]

jaegerTracer : Tracer = JaegerTracer().WithServer(default_deployer)

xtracer : XTracer = XTracerImpl().WithServer(default_deployer)

jaegerTraceModifier : Callable[str, Modifier] = lambda x : TracerModifier(tracer=jaegerTracer, service_name= x, sampling_rate= 1)

xTraceModifier : Modifier = XTraceModifier(tracer=xtracer)

localCollector : MetricCollector = LocalMetricCollector(filename="latency.csv").WithServer(default_deployer)

localMetricModifier : Modifier = MetricModifier(collector=localCollector, metrics=["latency"])

replicaModifier : Modifier = PlatformReplication(num_replicas=5)

server_modifiers : Callable[str, List[Modifier]] = lambda x : [jaegerTraceModifier(x), default_server_conn_opts, default_deployer, localMetricModifier, xTraceModifier, replicaModifier]

leafService : Service = LeafServiceImpl().WithServer(server_modifiers("LeafService")).WithClient(client_modifiers) 

nonleafService : Service = NonLeafServiceImpl(leafService=leafService).WithServer(server_modifiers("NonLeafService")).WithClient(client_modifiers)