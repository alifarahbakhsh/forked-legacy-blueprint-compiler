import ast
import copy

class InvalidModifierArgError(Exception):

    def __init__(self, instance_name):
        super().__init__("Invalid modifier args provided for instance " + instance_name)

class UndefinedModifierError(Exception):

    def __init__(self, modifier_name):
        super().__init__("Modifier named " + modifier_name + " is undefined")

class LambdaEvaluator(ast.NodeTransformer):

    def __init__(self, mappings):
        self.mappings = mappings

    def visit_Name(self, node):
        if node.id in self.mappings:
            val = self.mappings[node.id]
            return val
        return node

class ModifierPartialEvaluator:

    def __init__(self, modifiers, modifier_lists, modifier_lambdas, modifier_parser):
        self.modifier = modifiers
        self.modifier_lists = modifier_lists
        self.modifier_lambdas = modifier_lambdas
        self.modifier_parser = modifier_parser

    def get_mapping(self, function, value_node):
        mapping = {}
        for name, value in zip(function.args.args, value_node.args):
            mapping[name.arg] = value
        return mapping

    def evaluate_lambda(self, func_body, context, instance):
        modifiers = []
        body_evaluator = LambdaEvaluator(context)
        body_evaluator.visit(func_body)
        if isinstance(func_body, ast.List):
            for element in func_body.elts:
                if isinstance(element, ast.Name):
                    if element.id not in self.modifier:
                        raise UndefinedModifierError(element.id)
                    modifiers += [self.modifier[element.id]]
                if isinstance(element, ast.Call):
                    if element.func.id not in self.modifier_lambdas:
                        raise UndefinedModifierError(element.func.id)
                    function = self.modifier_lambdas[element.func.id].function
                    mapping = self.get_mapping(function, element)
                    modifiers = modifiers + self.evaluate_lambda(copy.deepcopy(function.body), mapping, instance)
        elif isinstance(func_body, ast.Call):
            minfo = self.modifier_parser.parse(instance.name + ' ' + func_body.func.id, func_body)
            modifiers += [minfo]
        return modifiers

    def get_modifier_list(self, arg, instance):
        modifiers = []
        if not isinstance(arg, ast.Name) and not isinstance(arg, ast.Call):
            raise InvalidModifierArgError(instance.name)
        if isinstance(arg, ast.Name):
            if arg.id in self.modifier:
                modifiers += [self.modifier[arg.id]]
            elif arg.id in self.modifier_lists:
                elements = self.modifier_lists[arg.id].elements
                for modifier in elements.elts:
                    if modifier.id not in self.modifier:
                        raise UndefinedModifierError(modifier.id)
                    modifiers += [self.modifier[modifier.id]]
            else:
                raise UndefinedModifierError(arg.id)
        if isinstance(arg, ast.Call):
            if arg.func.id not in self.modifier_lambdas:
                raise UndefinedModifierError(arg.func.id)
            function = self.modifier_lambdas[arg.func.id].function
            mapping = self.get_mapping(function, arg)
            modifiers = self.evaluate_lambda(copy.deepcopy(function.body), mapping, instance)
        return modifiers


    def partial_eval(self, instance):
        server_modifiers = []
        client_modifiers = []
        if instance.server_opts is not None:
            arg = instance.server_opts[0]
            server_modifiers = self.get_modifier_list(arg, instance)

        if instance.client_opts is not None:
            arg = instance.client_opts[0]
            client_modifiers = self.get_modifier_list(arg, instance)

        for p in instance.parameters:
            if p.client_opts is not None:
                arg = p.client_opts[0]
                p.client_modifiers = self.get_modifier_list(arg, instance)

        return server_modifiers, client_modifiers