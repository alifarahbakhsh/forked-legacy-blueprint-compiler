import copy
import ast
import astunparse

class AnnotationCleaner(ast.NodeTransformer):

    def visit_ClassDef(self, node):
        node.decorator_list = []
        self.generic_visit(node)
        return node

    def visit_AsyncFunctionDef(self, node):
        node.decorator_list = []
        node.type_comment = None
        node.returns = None
        for arg in node.args.args:
            arg.type_comment = None
            arg.annotation = None

        return node

    def visit_FunctionDef(self, node):
        if node.name == '__init__':
            return node
        node.decorator_list = []
        node.type_comment = None
        node.returns = None
        for arg in node.args.args:
            arg.type_comment = None
            arg.annotation = None

        return node

class InitArgsFinder(ast.NodeVisitor):

    def __init__(self):
        self.args = []

    def visit_FunctionDef(self, node):
        if node.name == '__init__':
            self.args = copy.deepcopy(node.args.args)

def create_copy(node):
    return copy.deepcopy(node)

def remove_annotations(node):
    cleaner = AnnotationCleaner()
    cleaner.visit(node)

def add_func_arg(node, name, arg_type):
    prev_arg_names = []
    for a in node.args.args:
        prev_arg_names += [a.arg]

    new_arg = ast.arg(arg=name, annotation=ast.Name(arg_type), type_comment=None)
    node.args.args += [new_arg]

    return prev_arg_names

def get_arg_names(node):
    arg_names = []
    for a in node.args.args:
        arg_names += [a.arg]

    return arg_names

def get_decorators(node):
    dec_list = []
    for d in node.decorator_list:
        if isinstance(d, ast.Name):
            dec_list += [d.id]
    return dec_list

def get_value(node):
    return astunparse.unparse(node).strip()

def get_init_args(node):
    finder = InitArgsFinder()
    finder.visit(node)
    return finder.args[1:]