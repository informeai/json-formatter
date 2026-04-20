package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func main() {
	var (
		data []byte
		err  error
		src  string
	)

	if len(os.Args) > 1 {
		src = os.Args[1]
		data, err = os.ReadFile(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not read file %q: %v\n", src, err)
			os.Exit(1)
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not read stdin: %v\n", err)
			os.Exit(1)
		}
	}

	// normaliza chaves sem aspas duplas antes de formatar
	data = quoteUnquotedKeys(data)

	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "    "); err != nil {
		line, col := offsetToLineCol(data, syntaxErrorOffset(err))
		fmt.Fprintf(os.Stderr, "error: line %d, col %d: %v\n", line, col, err)
		os.Exit(1)
	}

	fmt.Println(buf.String())
}

// quoteUnquotedKeys percorre o input byte a byte e adiciona aspas duplas
// em chaves de objetos JSON que não estejam entre aspas.
// Strings já entre aspas são copiadas integralmente sem modificação.
// Um identificador só é tratado como chave se for seguido de ':'.
func quoteUnquotedKeys(src []byte) []byte {
	var buf bytes.Buffer
	i := 0
	n := len(src)

	for i < n {
		b := src[i]

		// copia string entre aspas sem modificar o conteúdo
		if b == '"' {
			buf.WriteByte(b)
			i++
			for i < n {
				c := src[i]
				buf.WriteByte(c)
				i++
				if c == '\\' && i < n {
					// caractere escapado: copia junto e avança
					buf.WriteByte(src[i])
					i++
				} else if c == '"' {
					break
				}
			}
			continue
		}

		// após '{' ou ',' pode vir uma chave — verifica se é um identificador sem aspas
		// ou uma chave com apenas a aspa de abertura (ex: {"name: "value"})
		if b == '{' || b == ',' {
			buf.WriteByte(b)
			i++

			// copia espaços em branco
			for i < n && isSpace(src[i]) {
				buf.WriteByte(src[i])
				i++
			}

			// chave com apenas aspa de abertura: lookahead para ver se ':' vem antes de '"'
			if i < n && src[i] == '"' {
				j := i + 1
				for j < n && src[j] != '"' && src[j] != ':' {
					j++
				}
				if j < n && src[j] == ':' {
					// ':' encontrado antes do fechamento '"': aspa de fechamento ausente
					// lê o conteúdo da chave (sem a aspa de abertura), trimando espaços no final
					i++ // pula o '"' de abertura
					var key []byte
					for i < n && src[i] != ':' {
						key = append(key, src[i])
						i++
					}
					key = bytes.TrimRight(key, " \t")
					buf.WriteByte('"')
					buf.Write(key)
					buf.WriteByte('"')
					continue
				}
				// caso contrário, aspa bem formada — deixa o loop principal tratar normalmente
				continue
			}

			// se o próximo char inicia um identificador, lê o nome todo
			if i < n && isIdentStart(src[i]) {
				start := i
				for i < n && isIdentChar(src[i]) {
					i++
				}
				ident := src[start:i]

				// pula espaços após o identificador para checar se vem ':'
				// também pula uma '"' solta de fechamento (ex: name":"value")
				j := i
				for j < n && isSpace(src[j]) {
					j++
				}
				hasStrayQuote := j < n && src[j] == '"'
				if hasStrayQuote {
					j++
				}
				for j < n && isSpace(src[j]) {
					j++
				}

				if j < n && src[j] == ':' {
					// é uma chave: envolve com aspas e descarta a '"' solta do source
					if hasStrayQuote {
						i++ // pula a '"' solta
					}
					buf.WriteByte('"')
					buf.Write(ident)
					buf.WriteByte('"')
				} else {
					// não é chave (ex: true, false, null em array): escreve sem aspas
					buf.Write(ident)
				}
			}
			continue
		}

		buf.WriteByte(b)
		i++
	}

	return buf.Bytes()
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '$'
}

func isIdentChar(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}

// syntaxErrorOffset extrai o offset de um *json.SyntaxError, retornando 0 para outros erros.
func syntaxErrorOffset(err error) int64 {
	if se, ok := err.(*json.SyntaxError); ok {
		return se.Offset
	}
	return 0
}

// offsetToLineCol converte um offset em bytes para linha e coluna (1-indexed).
func offsetToLineCol(data []byte, offset int64) (line, col int) {
	line = 1
	col = 1
	if offset <= 0 {
		return
	}
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	for _, b := range data[:offset] {
		if b == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return
}
