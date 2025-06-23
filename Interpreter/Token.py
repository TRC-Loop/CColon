class Token:
    def __init__(self, type_, value):
        self.type = type_
        self.value = value
    def __repr__(self):
        return f"{self.type}({self.value})"

class Tokenizer:
    def __init__(self):
        self.keywords = {'var', 'string', 'int', 'function', 'return', 'if', 'else'}
        self.operators = {'=', '+', '-', '*', '/', '(', ')', ';', ',', '.', '{', '}', '=='}
        self.whitespace = {' ', '\t', '\n'}

    def tokenize(self, code):
        tokens = []
        i = 0
        while i < len(code):
            c = code[i]

            if c in self.whitespace:
                i += 1
                continue

            if c == '/' and i + 1 < len(code) and code[i+1] == '/':
                while i < len(code) and code[i] != '\n':
                    i += 1
                continue

            if c == '/' and i + 1 < len(code) and code[i+1] == '*':
                i += 2
                while i + 1 < len(code) and not (code[i] == '*' and code[i+1] == '/'): 
                    i += 1
                i += 2
                continue

            if c == '"':
                i += 1
                start = i
                while i < len(code) and code[i] != '"':
                    i += 1
                value = code[start:i]
                tokens.append(Token('STRING', value))
                i += 1
                continue

            if c.isalpha() or c == '_':
                start = i
                while i < len(code) and (code[i].isalnum() or code[i] == '_'):
                    i += 1
                value = code[start:i]
                type_ = 'KEYWORD' if value in self.keywords else 'ID'
                tokens.append(Token(type_, value))
                continue

            if c.isdigit():
                start = i
                while i < len(code) and code[i].isdigit():
                    i += 1
                value = code[start:i]
                tokens.append(Token('NUMBER', value))
                continue

            if c in self.operators:
                if c == '=' and i + 1 < len(code) and code[i+1] == '=':
                    tokens.append(Token('OP', '=='))
                    i += 2
                    continue
                tokens.append(Token('OP', c))
                i += 1
                continue

            raise SyntaxError(f"Unknown character: {c}")

        return tokens


code = open('example.ccolon').read()
def runDemo():
    tokenizer = Tokenizer()
    tokens = tokenizer.tokenize(code)
    return tokens
