default_server_conn_opts : Modifier = WebServer(framework="default")
default_deployer : Modifier = Deployer(framework="docker", public_ports=True)
server_modifiers : List[Modifier] = [default_server_conn_opts, default_deployer]

helloEvaluatorsService : HelloEvaluatorsService = HelloEvaluatorsImpl().WithServer(server_modifiers)