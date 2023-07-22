
from enum import Enum, auto


class FunctionInfo:
    def __init__(self, func_name, ast_node, is_local, throws, is_async, is_entry):
        self.func_name = func_name
        self.ast_node = ast_node
        self.is_local = is_local
        self.throws = throws
        self.is_async = is_async
        self.is_entry = is_entry

class ClassInfo:
    def __init__(self, name, ast_node, bases):
        self.name = name
        self.ast_node = ast_node
        self.bases = bases

class FileInfo:
    def __init__(self, filename, classes, imports, functions):
        self.filename = filename
        self.classes = classes
        self.imports = imports
        self.functions = functions

class ModifierInfo:
    def __init__(self, name, modifier_type, modifier_params):
        self.name = name
        self.modifier_type = modifier_type
        self.modifier_params = modifier_params

class ModifierListInfo:
    def __init__(self, list_name, value_node):
        self.name = list_name
        self.elements = value_node

class ModifierLambdaInfo:
    def __init__(self, lambda_name, value_node):
        self.name = lambda_name
        self.function = value_node

class ServiceType(Enum):
    RPCSERVICE = auto()
    QUEUESERVICE = auto()
    HYBRIDSERVICE = auto()
    EXTERNALSERVICE = auto()

class ServiceInfo:
    def __init__(self, name):
        self.name = name
        self.service_type = ServiceType.RPCSERVICE
        self.is_user_defined = True
        self.functions = []
        self.imports = []
        self.classes = []
        self.init_args = {}
        self.filename = ""
        self.ast_node = None
        self.bases = []

    def is_rpc_service(self):
        return self.service_type == ServiceType.RPCSERVICE

    def is_queue_service(self):
        return self.service_type == ServiceType.QUEUESERVICE

class Component:

    def __init__(self, name, func_decls):
        self.name = name
        self.func_decls = func_decls
        self.choices = []

    def add_choice(self, choice):
        self.choices += [choice]

class Choice:
    def __init__(self, name, components, ast_node, filename):
        self.name = name
        self.components = components
        self.ast_node = ast_node
        self.filename = filename

class ServiceInstance:
    def __init__(self, name, address, port):
        self.name = name
        self.address = address
        self.port = port
        self.abstract_type = None
        self.actual_type = None
        self.parameters = []
        self.server_opts = None
        self.client_opts = None
        self.default_server_modifiers = []
        self.default_client_modifiers = []
        self.server_framework = None
        self.dependencies = []
        self.choice_deployer = None
        self.modifier_instance_opts = {}

    def add_parameter(self, keyword_name, node, instance_name=None, isserviceinstance=False):
        p = ServiceParameterInfo(keyword_name, node, instance_name, isserviceinstance)
        self.parameters += [p]

class ServiceParameterInfo:
    '''
    Supports only keyworded arguments.
    '''
    def __init__(self, keyword_name, node, instance_name=None, isserviceinstance=False):
        self.isinstance = isserviceinstance
        self.instance_name = instance_name
        self.node = node
        self.keyword_name = keyword_name
        self.client_opts = None
        self.client_modifiers = []
        self.client_node = None

class ConfigOptions:

    def __init__(self):
        self.app_name = ""
        self.src_dir = ""
        self.output_dir = ""
        self.target = ""
        self.wiring_file = ""
        self.dependency_file = ""
        self.deploy_address = ""