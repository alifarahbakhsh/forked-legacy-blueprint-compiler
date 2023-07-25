import sys
from wiring import ModifierRegistry, ModifierPartialEvaluator, WiringParser, ModifierParser, COMPONENTS
import astunparse
import json

class JsonObject:
    def toJSON(self):
        return json.dumps(self, default=lambda o: o.__dict__, 
            sort_keys=True, indent=4)

class ParameterInfo(JsonObject):
    def __init__(self, name, isservice, client_modifiers, param_val, keyword_name):
        self.name = name
        self.isservice = isservice
        self.keyword_name = keyword_name
        self.client_modifiers = client_modifiers
        self.client_node = param_val

class ModifierInfo(JsonObject):
    def __init__(self, modifier_type, modifier_params):
        self.modifier_type = modifier_type
        self.modifier_params = modifier_params

class Node(JsonObject):
    def __init__(self, name, children):
        self.name = name
        self.children = children

    def add_child(self, node):
        self.children += [node]

class ContainerNode(Node):
    def __init__(self, name, children):
        super().__init__(name, children)

class ProcessNode(Node):
    def __init__(self, name, children):
        super().__init__(name, children)
        self.abstract_type = "Process"

class MillenialRootNode(Node):
    def __init__(self, children):
        super().__init__("root", children)

class ServiceNode(Node):
    def __init__(self, name, server_modifiers, client_modifiers, arguments, actual_type, abstract_type):
        super().__init__(name, [])
        self.server_modifiers = server_modifiers
        self.client_modifier = client_modifiers
        self.arguments = arguments
        self.actual_type = actual_type
        self.abstract_type = abstract_type

def main():
    if len(sys.argv) != 3:
        print("Usage: python wiring_translate.py <path/to/wiring_file> <output/path>")
        sys.exit(1)

    wiring_file = sys.argv[1]
    output_file = sys.argv[2]
    modifier_registry = ModifierRegistry()
    wiringParser = WiringParser(service_instances={}, service_infos={}, components=COMPONENTS, choices={}, modifier_registry=modifier_registry)
    wiringParser.parse_wiring(wiring_file)
    service_instances = wiringParser.service_instances
    modifiers = wiringParser.defined_modifiers
    modifier_lists = wiringParser.defined_modifier_lists
    modifier_lambdas = wiringParser.defined_modifier_lambdas
    processes = wiringParser.defined_processes

    modifierEvaluator = ModifierPartialEvaluator(modifiers, modifier_lists, modifier_lambdas, ModifierParser())

    for name, instance in service_instances.items():
        server_modifiers, client_modifiers = modifierEvaluator.partial_eval(instance)
        instance.default_client_modifiers = client_modifiers
        instance.default_server_modifiers = server_modifiers
        service_instances[name] = instance

    root_node = MillenialRootNode(children=[])
    service_node_map = {}
    container_counter = 1
    for name, instance in service_instances.items():
        if instance.abstract_type in COMPONENTS:
            # COMPONENTS only need a Container Node
            client_modifiers = []
            server_modifiers = []
            arguments = []
            for cm in instance.default_client_modifiers:
                modifier_params = []
                for p in cm.modifier_params:
                    client_val = None
                    if p.node is not None:
                        client_val = astunparse.unparse(p.node).strip()
                    param = ParameterInfo(name=p.instance_name, keyword_name=p.keyword_name, param_val=client_val, isservice=p.isinstance, client_modifiers=[])
                    modifier_params += [param]
                modifier = ModifierInfo(modifier_type=cm.modifier_type, modifier_params=modifier_params)
                client_modifiers += [modifier]
            for sm in instance.default_server_modifiers:
                modifier_params = []
                for p in sm.modifier_params:
                    client_val = None
                    if p.node is not None:
                        client_val = astunparse.unparse(p.node).strip()
                    param = ParameterInfo(name=p.instance_name, keyword_name=p.keyword_name, param_val=client_val, isservice=p.isinstance, client_modifiers=[])
                    modifier_params += [param]
                modifier = ModifierInfo(modifier_type=sm.modifier_type, modifier_params=modifier_params)
                server_modifiers += [modifier]
            for a in instance.parameters:
                param_modifiers = []
                for pm in a.client_modifiers:
                    modifier = ModifierInfo(modifier_type=pm.modifier_type, modifier_params=pm.modifier_params)
                    param_modifiers += [modifier]
                client_val = None
                if a.node is not None:
                    client_val = astunparse.unparse(a.node).strip()
                param = ParameterInfo(name=a.instance_name, isservice=a.isinstance, client_modifiers=param_modifiers, param_val=client_val, keyword_name=a.keyword_name)
                arguments += [param]
            service_node = ServiceNode(name=instance.name, client_modifiers=client_modifiers, server_modifiers=server_modifiers, arguments=arguments, actual_type=instance.actual_type, abstract_type=instance.abstract_type)
            container_node = ContainerNode("container" + str(container_counter), children=[service_node])
            container_counter += 1
            root_node.add_child(container_node)
        elif instance.abstract_type == 'Service' or instance.abstract_type == 'QueueService' or instance.abstract_type.endswith("Service"):
            # Service instances need both Container Node and Process Node            
            client_modifiers = []
            server_modifiers = []
            arguments = []
            for cm in instance.default_client_modifiers:
                modifier_params = []
                for p in cm.modifier_params:
                    client_val = None
                    if p.node is not None:
                        client_val = astunparse.unparse(p.node).strip()
                    param = ParameterInfo(name=p.instance_name, keyword_name=p.keyword_name, param_val=client_val, isservice=p.isinstance, client_modifiers=[])
                    modifier_params += [param]
                modifier = ModifierInfo(modifier_type=cm.modifier_type, modifier_params=modifier_params)
                client_modifiers += [modifier]
            for sm in instance.default_server_modifiers:
                modifier_params = []
                for p in sm.modifier_params:
                    client_val = None
                    if p.node is not None:
                        client_val = astunparse.unparse(p.node).strip()
                    param = ParameterInfo(name=p.instance_name, keyword_name=p.keyword_name, param_val=client_val, isservice=p.isinstance, client_modifiers=[])
                    modifier_params += [param]
                modifier = ModifierInfo(modifier_type=sm.modifier_type, modifier_params=modifier_params)
                server_modifiers += [modifier]
            for a in instance.parameters:
                param_modifiers = []
                for pm in a.client_modifiers:
                    modifier = ModifierInfo(modifier_type=pm.modifier_type, modifier_params=pm.modifier_params)
                    param_modifiers += [modifier]
                client_val = None
                if a.node is not None:
                    client_val = astunparse.unparse(a.node).strip()
                param = ParameterInfo(name=a.instance_name, isservice=a.isinstance, client_modifiers=param_modifiers, param_val=client_val, keyword_name=a.keyword_name)
                arguments += [param]
            service_node = ServiceNode(name=instance.name, client_modifiers=client_modifiers, server_modifiers=server_modifiers, arguments=arguments, actual_type=instance.actual_type, abstract_type=instance.abstract_type)
            service_node_map[instance.name] = service_node

    seen_services = {}
    for proc_name, process in processes.items():
        service_nodes = []
        for instance in process.instances:
            if instance.instance_name in seen_services:
                raise Exception("Double use of a service instance")
            seen_services[instance.instance_name] = True
            service_node = service_node_map[instance.instance_name]
            service_nodes += [service_node]
        process_node = ProcessNode(proc_name, children=service_nodes)
        container_node = ContainerNode("container" + str(container_counter), children=[process_node])
        container_counter += 1
        root_node.add_child(container_node)

    # Take care of orphan services
    num_procs = len(processes)
    for name, service_node in service_node_map.items():
        if name in seen_services:
            continue
        num_procs += 1
        process_node = ProcessNode("Proc" + str(num_procs), children=[service_node])
        container_node = ContainerNode("container" + str(container_counter), children=[process_node])
        container_counter += 1
        root_node.add_child(container_node)


    with open(output_file, 'w+') as outf:
        outf.write(root_node.toJSON())

    print("Wiring Translation Completed")

if __name__ == '__main__':
    main()