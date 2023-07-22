from .dataModels import ModifierInfo, ServiceParameterInfo
import ast

class ModifierParser:
    'Standalone modifier parser to be used by the Modifier Evaluator'

    def parse(self, name, body):
        params = []
        for k in body.keywords:
            keyword_name = k.arg
            isserviceinstance = False
            instance_name = None
            if isinstance(k.value, ast.Name):
                isserviceinstance = True
                instance_name = k.value.id
            pinfo = ServiceParameterInfo(keyword_name, k.value, instance_name, isserviceinstance)
            params += [pinfo]
        modifier_type = body.func.id
        return ModifierInfo(name, modifier_type, params)
