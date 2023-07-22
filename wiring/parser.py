import ast
import os
from .dataModels import ProcessInfo, ServiceParameterInfo, ModifierInfo, ModifierListInfo, ModifierLambdaInfo, ServiceInstance
from enum import Enum, auto

class AnnotationType(Enum):
    NORMAL = auto()
    LIST = auto()
    LAMBDA = auto()

class InvalidWiringSyntaxError(Exception):
    def __init__(self, lineno, msg):
        self.lineno = lineno
        self.msg = msg
        super().__init__("Line " + str(self.lineno) + ": " + self.msg)

class UnAnnotatedAssignError(Exception):
    def __init__(self, lineno):
        self.lineno = lineno
        super().__init__("UnAnnotated Assignment statement found in wiring file at line " + str(self.lineno))

class UndeclaredServiceError(Exception):
    def __init__(self, lineno, name):
        self.lineno = lineno
        self.name = name
        super().__init__("Line " + str(self.lineno) + ": Service Type " + self.name + " was not declared in the specification")

class DuplicateInstanceNameError(Exception):
    def __init__(self, lineno, name):
        self.lineno = lineno
        self.name = name
        super().__init__("Line " + str(self.lineno) + ": Instance Name " + self.name + " has been previously used!")

class UnknownModifierError(Exception):
    def __init__(self, lineno, name):
        self.lineno = lineno
        self.name = name
        super().__init__("Line " + str(self.lineno) + ": Modifier " + self.name + " was never registered.")

class UseBeforeDeclaredError(Exception):
    def __init__(self, lineno, name):
        self.lineno = lineno
        self.name = name
        super().__init__("Line " + str(self.lineno) + ": Instance " + self.name + " used before declaration")

class InvalidAnnotationTypeError(Exception):
    def __init__(self, lineno, annotation_name):
        self.lineno = lineno
        self.name = annotation_name
        super().__init__("Line " + str(self.lineno) + ": Invalid annoation type " + self.name)

class WiringDataCollector(ast.NodeVisitor):

    def __init__(self, service_instances, service_infos, components, choices, modifier_registry):
        self.service_instances = service_instances
        self.service_infos = service_infos
        self.components = components
        self.choices = choices
        self.defined_instances = {}
        self.defined_modifiers = {}
        self.defined_modifier_lists = {}
        self.defined_modifier_lambdas = {}
        self.defined_processes = {}
        self.modifier_registry = modifier_registry

    def parse_client_overriding(self, value_node):
        client_opts_node = None
        instance_name = ""
        if isinstance(value_node.func, ast.Attribute):
            instance_name = value_node.func.value.id
            if value_node.func.attr != "WithClient":
                raise InvalidWiringSyntaxError(value_node.lineno, "Instance parameter can only be modified using '.WithClient' method")
            client_opts_node = value_node.args
        else:
            raise InvalidWiringSyntaxError(value_node.lineno, "Instance parameter can only be modified using '.WithClient' method")
        return instance_name, client_opts_node

    def get_param_infos(self, params):
        pinfos = []
        for p in params:
            keyword_name = p.arg
            isserviceinstance = False
            instance_name = None
            client_opts_node = None
            if isinstance(p.value, ast.Name):
                isserviceinstance = True
                instance_name = p.value.id
                #if instance_name not in self.defined_instances:
                #    raise UseBeforeDeclaredError(p.lineno, instance_name)
            elif isinstance(p.value, ast.Call):
                isserviceinstance = True
                instance_name, client_opts_node = self.parse_client_overriding(p.value)
            elif isinstance(p.value, ast.Constant) or isinstance(p.value, ast.Dict) or isinstance(p.value, ast.List):
                # Constants are allowed!
                pass
            else:
                raise InvalidWiringSyntaxError(p.lineno, "Argument to service instantiation must be a constant or an instance")
            pinfo = ServiceParameterInfo(keyword_name, p.value, instance_name, isserviceinstance)
            pinfo.client_opts = client_opts_node
            pinfos += [pinfo]
        return pinfos

    def get_process_params(self, params):
        pinfos = []
        if len(params) > 1:
            raise InvalidWiringSyntaxError(params[0].lineno, "Process only takes 1 argument named 'services'")
        for p in params:
            keyword_name = p.arg
            if keyword_name != 'services':
                raise InvalidWiringSyntaxError(p.lineno, "Process only takes 1 argument named 'services'") 
            if not isinstance(p.value, ast.List):
                raise InvalidWiringSyntaxError(p.lineno, "'services' argument expected list")
            for element in p.value.elts:
                if not isinstance(element, ast.Name):
                    raise InvalidWiringSyntaxError(p.lineno, "Only names of service instances are permitted in the 'services' argument")
                pinfo = ServiceParameterInfo("", element, element.id, True)
                pinfos += [pinfo]
        return pinfos

    
    def visit_Assign(self, node):
        raise UnAnnotatedAssignError(node.lineno)

    def parse_annotation(self, annotation):
        annotation_type = None
        abstract_types = []
        if isinstance(annotation, ast.Name):
            annotation_type = AnnotationType.NORMAL
            abstract_types += [annotation.id]
        elif isinstance(annotation, ast.Subscript):
            if annotation.value.id == 'List':
                annotation_type = AnnotationType.LIST
                abstract_types += [annotation.slice.value.id]
            elif annotation.value.id == 'Callable':
                annotation_type = AnnotationType.LAMBDA
                abstract_types += [annotation.slice.value.elts[1]]
            else:
                raise InvalidAnnotationTypeError(annotation.lineno, annotation.value.id)
        else:
            raise InvalidAnnotationTypeError(annotation.lineno, annotation.value.id)
        return annotation_type, abstract_types

    def parse_service_instantiation(self, value_node,with_client=False,with_server=False):
        server_opts_node = None
        client_opts_node = None
        actual_type = None
        params = []
        is_client = False
        is_server = False
        if isinstance(value_node.func, ast.Name):
            actual_type = value_node.func.id
            params = self.get_param_infos(value_node.keywords)
        elif isinstance(value_node.func, ast.Attribute):
            attribute_name = value_node.func.attr
            if attribute_name == "WithClient":
                if with_client:
                    raise InvalidWiringSyntaxError(value_node.lineno, "WithClient can only be used once")
                with_client = True
                is_client = True
            elif attribute_name == "WithServer":
                if with_server:
                    raise InvalidWiringSyntaxError(value_node.lineno, "With")
                with_server = True
                is_server = True
            actual_type, params, server_opts_node, client_opts_node = self.parse_service_instantiation(value_node.func.value, with_client, with_server)
            if is_client:
                client_opts_node = value_node.args
            elif is_server:
                server_opts_node = value_node.args
        
        return actual_type, params, server_opts_node, client_opts_node

    def visit_AnnAssign(self, node):
        instance_name = node.target.id
        annotation_type, abstract_types = self.parse_annotation(node.annotation)

        is_modifier = False
        is_service = False
        is_process = False

        if annotation_type == AnnotationType.NORMAL:
            abstract_type = abstract_types[0]
            if abstract_type == "Service" or abstract_type == "QueueService" or abstract_type.endswith("Service"):
                is_service = True
                actual_type, params, server_opts_node, client_opts_node = self.parse_service_instantiation(node.value)
                if not isinstance(node.value, ast.Call):
                    raise InvalidWiringSyntaxError(node.lineno, "Instance of Service type must be an object of a known service")
                #if actual_type not in self.service_infos:
                    #raise UndeclaredServiceError(node.lineno, node.value.func.id)
                if instance_name not in self.service_instances:
                    self.service_instances[instance_name] = ServiceInstance(instance_name, address="", port=None)
                elif instance_name in self.defined_instances:
                    raise DuplicateInstanceNameError(node.lineno, instance_name)
            elif abstract_type in self.components:
                is_service = True
                actual_type, params, server_opts_node, client_opts_node = self.parse_service_instantiation(node.value)
                if not isinstance(node.value, ast.Call):
                    raise InvalidWiringSyntaxError(node.lineno, "Instance of a Component type must be an object of a known Choice")
                if instance_name not in self.service_instances:
                    self.service_instances[instance_name] = ServiceInstance(instance_name, address="", port=None)
                elif instance_name in self.defined_instances:
                    raise DuplicateInstanceNameError(node.lineno, instance_name)
            elif abstract_type == 'Modifier':
                is_modifier = True
                # Check if actual_type is in the appropriate list of modifier types!
                if not self.modifier_registry.is_valid_modifier(node.value.func.id):
                    raise UnknownModifierError(node.lineno, node.value.func.id)
                actual_type = node.value.func.id
                params = self.get_param_infos(node.value.keywords)
            elif abstract_type == 'Process':
                is_process = True
                actual_type = node.value.func.id
                params = self.get_process_params(node.value.keywords)

            else:
                raise InvalidWiringSyntaxError(node.lineno, "Currently only instances of services or components are supported")
            
            if is_service:
                instance = self.service_instances[instance_name]            
                instance.abstract_type = abstract_type
                instance.actual_type = actual_type
                instance.parameters = params
                instance.server_opts = server_opts_node
                instance.client_opts = client_opts_node
                self.service_instances[instance_name] = instance
            elif is_modifier:
                modifierInfo = ModifierInfo(instance_name, actual_type,  params)
                self.defined_modifiers[instance_name] = modifierInfo
            elif is_process:
                processInfo = ProcessInfo(instance_name, params)
                self.defined_processes[instance_name] = processInfo       
        elif annotation_type == AnnotationType.LIST:
            element_type = abstract_types[0]
            if element_type != 'Modifier':
                raise InvalidWiringSyntaxError(node.lineno, "List annotation type is only supported with Modifiers")
            modifier_list = ModifierListInfo(instance_name, node.value)
            self.defined_modifier_lists[instance_name] = modifier_list
        elif annotation_type == AnnotationType.LAMBDA:
            lambda_result_node = abstract_types[0]
            if isinstance(lambda_result_node, ast.Name):
                if lambda_result_node.id != 'Modifier':
                    raise InvalidWiringSyntaxError(node.lineno, "Lambda annotaiton type must produce a Modifier or a List[Modifier]")
                if not self.modifier_registry.is_valid_modifier(node.value.body.func.id):
                    raise UnknownModifierError(node.lineno, node.value.body.func.id)
            elif isinstance(lambda_result_node, ast.Subscript):
                if lambda_result_node.value.id != 'List':
                    raise InvalidWiringSyntaxError(node.lineno, "Lambda annotaiton type must produce a Modifier or a List[Modifier]")
                if lambda_result_node.slice.value.id != 'Modifier':
                    raise InvalidWiringSyntaxError(node.lineno, "Lambda annotaiton type must produce a Modifier or a List[Modifier]")
            modifier_lambda = ModifierLambdaInfo(instance_name, node.value)
            self.defined_modifier_lambdas[instance_name] = modifier_lambda
        
        self.defined_instances[instance_name] = True


class WiringParser:
    def __init__(self, service_instances, service_infos, components, choices, modifier_registry):
        self.service_instances = service_instances
        self.service_infos = service_infos
        self.components = components
        self.choices = choices
        self.defined_modifiers = {}
        self.defined_modifier_lists = {}
        self.defined_modifier_lambdas = {}
        self.modifier_registry = modifier_registry
        self.defined_processes = {}

    def parse_file(self, filename):
        with open(filename, 'r') as inf:
            tree = ast.parse(inf.read(), type_comments=True)
            return tree

    def parse_wiring(self, filename):
        tree = self.parse_file(filename)
        wiringParser = WiringDataCollector(self.service_instances, self.service_infos, self.components, self.choices, self.modifier_registry)
        wiringParser.visit(tree)
        self.service_instances = wiringParser.service_instances
        self.defined_modifiers = wiringParser.defined_modifiers
        self.defined_modifier_lists = wiringParser.defined_modifier_lists
        self.defined_modifier_lambdas = wiringParser.defined_modifier_lambdas
        self.defined_processes = wiringParser.defined_processes