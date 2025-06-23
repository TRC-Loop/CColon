import AST

from AST import VarDecl, FuncCall, Import, FuncDef, IfElse

class Interpreter:
    def __init__(self):
        self.vars = {}

    def eval(self, node):
        if isinstance(node, Import):
            return None
        if isinstance(node, FuncDef):
            self.vars[node.name] = node
            if node.name == 'main':
                for stmt in node.body:
                    self.eval(stmt)
            return None
        if isinstance(node, VarDecl):
            value = self.eval(node.value) if not isinstance(node.value, (str, int)) else node.value
            self.vars[node.name] = value
            return None
        if isinstance(node, IfElse):
            cond = self.eval(node.cond)
            if cond:
                for stmt in node.if_body:
                    self.eval(stmt)
            elif node.else_body is not None:
                for stmt in node.else_body:
                    self.eval(stmt)
            return None
        if isinstance(node, FuncCall):
            obj_val = self.eval(node.obj) if isinstance(node.obj, FuncCall) else node.obj
            if obj_val == 'console' and node.method == 'println':
                out = []
                for arg in node.args:
                    out.append(str(self.eval(arg)))
                print(''.join(out))
                return None
            if obj_val == 'console' and node.method == 'scanp':
                prompt = str(self.eval(node.args[0])) if node.args else ''
                return input(prompt)
            if node.method == 'tostring':
                val = self.eval(node.obj) if hasattr(node, 'obj') else node.args[0]
                return str(val)
        if isinstance(node, tuple) and len(node) == 3:
            left = self.eval(node[0])
            right = self.eval(node[2])
            if node[1] == '+':
                if isinstance(left, int) and isinstance(right, int):
                    return left + right
                return str(left) + str(right)
            elif node[1] == '-':
                return int(left) - int(right)
            elif node[1] == '*':
                return int(left) * int(right)
            elif node[1] == '/':
                return int(left) // int(right)
            elif node[1] == '==':
                return left == right
        if isinstance(node, str):
            return self.vars.get(node, node)
        if isinstance(node, int):
            return node
        if isinstance(node, bool):
            return node
        raise Exception(f"Unknown node type: {node}")

    def run(self, ast):
        for node in ast:
            self.eval(node)

ast = AST.runDemo()
interpreter = Interpreter()
interpreter.run(ast)
