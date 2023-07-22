
COMPONENTS = {
    "Cache": True, 
    "NoSQLDatabase": True, 
    "Queue": True, 
    "RelationalDB": True, 
    "Tracer": True, 
    "MetricCollector": True, 
    "XTracer": True,
    "LoadBalancer": True
}

valid_client_modifiers = {
    "ClientPool", #ClientPoolModifier
    "Retry",
    "CircuitBreaker"
}

valid_server_modifiers = {
    "MetricModifier",
    "TracerModifier", #TracerModifier
    "ThreadedServer",
    "RPCServer",
    "WebServer",
    "Deployer",
    "XTraceModifier",
    "PlatformReplication"
}

class ModifierRegistry:

    def is_valid_client_modifier(self, modifier_name):
        return modifier_name in valid_client_modifiers 

    def is_valid_server_modifier(self, modifier_name):
        return modifier_name in valid_server_modifiers

    def is_valid_modifier(self, modifier_name):
        return self.is_valid_client_modifier(modifier_name) or self.is_valid_server_modifier(modifier_name)