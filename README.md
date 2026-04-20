# JSON Formatter

Uma ferramenta de linha de comando para formatar e validar arquivos JSON.

## Recursos

- Formata arquivos JSON com indentação de 4 espaços
- Suporta chaves sem aspas duplas (JavaScript-like)
- Lê de arquivo ou stdin
- Erros detalhados com linha e coluna

## Instalação

```bash
go install github.com/informeai/json-formatter@latest
```

## Uso

### Via arquivo

```bash
json-formatter arquivo.json
```

### Via stdin

```bash
echo '{nome: "João", idade: 30}' | json-formatter
```

## Exemplos

**Entrada:**
```json
{nome: "João", idade: 30, ativos: true}
```

**Saída:**
```json
{
    "nome": "João",
    "idade": 30,
    "ativos": true
}
```

## Erros

Em caso de erro de sintaxe, o formato mostra a linha e coluna:

```
error: line 2, col 5: unexpected character 'x'
```
