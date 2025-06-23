"""Basic parser for variable declaration and function call"""

import Token

class ASTNode:
    pass

class VarDecl(ASTNode):
    def __init__(self, var_type, name, value):
        self.var_type = var_type
        self.name = name
        self.value = value
    def __repr__(self):
        return f"VarDecl(type={self.var_type}, name={self.name}, value={self.value})"

class FuncCall(ASTNode):
    def __init__(self, obj, method, args):
        self.obj = obj
        self.method = method
        self.args = args
    def __repr__(self):
        return f"FuncCall(obj={self.obj}, method={self.method}, args={self.args})"

class Import(ASTNode):
    def __init__(self, module):
        self.module = module
    def __repr__(self):
        return f"Import(module={self.module})"

class FuncDef(ASTNode):
    def __init__(self, name, body):
        self.name = name
        self.body = body
    def __repr__(self):
        return f"FuncDef(name={self.name}, body={self.body})"

class IfElse(ASTNode):
    def __init__(self, cond, if_body, else_body=None):
        self.cond = cond
        self.if_body = if_body
        self.else_body = else_body
    def __repr__(self):
        return f"IfElse(cond={self.cond}, if_body={self.if_body}, else_body={self.else_body})"

class Parser:
    def __init__(self, tokens):
        self.tokens = tokens
        self.pos = 0

    def current(self):
        return self.tokens[self.pos] if self.pos < len(self.tokens) else None

    def consume(self, expected_type=None, expected_value=None):
        token = self.current()
        if not token:
            raise Exception("Unexpected end of input")
        if expected_type and token.type != expected_type:
            raise Exception(f"Expected type {expected_type}, got {token.type}")
        if expected_value and token.value != expected_value:
            raise Exception(f"Expected value {expected_value}, got {token.value}")
        self.pos += 1
        return token

    def parse(self):
        ast = []
        while self.pos < len(self.tokens):
            curr = self.current()
            if curr and curr.type == 'KEYWORD' and curr.value == 'import':
                ast.append(self.parse_import())
            elif curr and curr.type == 'KEYWORD' and curr.value == 'function':
                ast.append(self.parse_func_def())
            elif curr and curr.type == 'KEYWORD' and curr.value == 'var':
                ast.append(self.parse_var_decl())
            elif curr and curr.type == 'KEYWORD' and curr.value == 'if':
                ast.append(self.parse_if_else())
            else:
                ast.append(self.parse_expression())
        return ast

    def parse_var_decl(self):
        self.consume('KEYWORD', 'var')
        var_type = self.consume('KEYWORD').value
        var_name = self.consume('ID').value
        self.consume('OP', '=')
        value = self.parse_expression()
        return VarDecl(var_type, var_name, value)

    def parse_import(self):
        self.consume('KEYWORD', 'import')
        module = self.consume('ID').value
        return Import(module)

    def parse_func_def(self):
        self.consume('KEYWORD', 'function')
        name = self.consume('ID').value
        self.consume('OP', '(')
        self.consume('OP', ')')
        self.consume('OP', '{')
        body = self.parse_block()
        self.consume('OP', '}')
        return FuncDef(name, body)

    def parse_if_else(self):
        self.consume('KEYWORD', 'if')
        self.consume('OP', '(')
        cond = self.parse_expression()
        self.consume('OP', ')')
        self.consume('OP', '{')
        if_body = self.parse_block()
        self.consume('OP', '}')
        else_body = None
        curr = self.current()
        if curr and curr.type == 'KEYWORD' and curr.value == 'else':
            self.consume('KEYWORD', 'else')
            self.consume('OP', '{')
            else_body = self.parse_block()
            self.consume('OP', '}')
        return IfElse(cond, if_body, else_body)

    def parse_block(self):
        body = []
        while True:
            curr = self.current()
            if curr is None:
                break
            if curr.type == 'OP' and curr.value == '}':
                break
            
            if curr.type == 'KEYWORD' and curr.value == 'var':
                body.append(self.parse_var_decl())
            elif curr.type == 'KEYWORD' and curr.value == 'if':
                body.append(self.parse_if_else())
            else:
                body.append(self.parse_expression())
        return body

    def parse_expression(self, min_precedence=0):
        precedences = {'+': 1, '-': 1, '*': 2, '/': 2, '==': 0}
        token = self.current()
        if token is None:
            raise Exception("Unexpected end of input in expression")
        if token.type == 'ID':
            expr = self.consume('ID').value
            while True:
                curr = self.current()
                if curr is None:
                    break
                if curr.type == 'OP' and curr.value == '.':
                    self.consume('OP', '.')
                    method = self.consume('ID').value
                    self.consume('OP', '(')
                    args = self.parse_args()
                    self.consume('OP', ')')
                    expr = FuncCall(expr, method, args)
                else:
                    break
        elif token.type == 'STRING':
            expr = self.consume('STRING').value
        elif token.type == 'NUMBER':
            expr = int(self.consume('NUMBER').value)
        elif token.type == 'OP' and token.value == '(':
            self.consume('OP', '(')
            expr = self.parse_expression()
            self.consume('OP', ')')
        else:
            raise Exception(f"Unexpected token in expression: {token}")

        while True:
            curr = self.current()
            if curr is None:
                break
            if not (curr.type == 'OP' and curr.value in precedences and precedences[curr.value] >= min_precedence):
                break
            op = self.consume('OP').value
            next_min_prec = precedences[op] + 1 if op in {'*', '/'} else precedences[op]
            right = self.parse_expression(next_min_prec)
            expr = (expr, op, right)
        return expr

    def parse_args(self):
        args = []
        curr = self.current()
        if curr is not None and curr.type == 'OP' and curr.value == ')':
            return args

        while True:
            arg_expr = self.parse_expression()
            args.append(arg_expr)

            curr = self.current()
            if curr is None:
                raise Exception("Unexpected end of input while parsing arguments. Expected ')' or ','")
            if curr.type == 'OP' and curr.value == ')':
                break
            elif curr.type == 'OP' and curr.value == ',':
                self.consume('OP', ',')
            else:
                raise Exception(f"Unexpected token {curr.value} while parsing arguments. Expected ')' or ','")
        return args

def runDemo():
    tokens = Token.runDemo()
    parser = Parser(tokens)
    ast = parser.parse()
    return ast
