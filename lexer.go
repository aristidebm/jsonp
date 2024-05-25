package jsonp

import (
	"bufio"
	"io"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type TokenType string

const (
	LBRACE     TokenType = "LBRACE"
	RBRACE               = "RBRACE"
	LBRACKET             = "LBRACKET"
	RBRACKET             = "RBRACKET"
	COMMA                = "COMMA"
	COLON                = "COLON"
	TRUE                 = "TRUE"
	FALSE                = "FALSE"
	NULL                 = "NULL"
	NUMBER               = "NUMBER"
	STRING               = "STRING"
	EOF                  = "EOF"
	WHITESPACE           = "WHITESPACE"
	INVALID              = "\\"
)

var TokenDefinition = map[TokenType]*regexp.Regexp{
	LBRACE:     regexp.MustCompile(`{`),
	RBRACE:     regexp.MustCompile(`}`),
	LBRACKET:   regexp.MustCompile(`\[`),
	RBRACKET:   regexp.MustCompile(`\]`),
	COMMA:      regexp.MustCompile(`,`),
	COLON:      regexp.MustCompile(`:`),
	TRUE:       regexp.MustCompile(`true`),
	FALSE:      regexp.MustCompile(`false`),
	NULL:       regexp.MustCompile(`null`),
	NUMBER:     regexp.MustCompile(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:e[-+]\d+)?`),
	STRING:     regexp.MustCompile(`".*?[^\\]"`),
	WHITESPACE: regexp.MustCompile(`\s+`),
	EOF:        regexp.MustCompile(`^$`),
}

type Token struct {
	Type  TokenType `json:"type"`
	Value string    `json:"value"`
	Loc   Location  `json:"loc"`
}

type Location struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

func newPosition(line int, col int, offset int) Position {
	return Position{
		Line:   line,
		Column: col,
		Offset: offset,
	}
}

func newToken(tokenType TokenType, value string, start Position, end Position) Token {
	return Token{
		Type:  tokenType,
		Value: value,
		Loc: Location{
			Start: start,
			End:   end,
		},
	}
}

type Lexer struct {
	Input        io.Reader
	lineNumber   int
	numberOfChar int
}

func NewLexer(input io.Reader) *Lexer {
	return &Lexer{
		Input: input,
	}
}

func (l *Lexer) Lex() ([]Token, error) {
	var tokens []Token

	sc := bufio.NewScanner(l.Input)

	for sc.Scan() {
		line := sc.Text() + "\n"
		tokens = append(tokens, l.lex(line)...)
		l.lineNumber++
		l.numberOfChar += utf8.RuneCountInString(line)
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc[[]Token, Token](tokens, func(a, b Token) int {
		if a.Loc.Start.Line == b.Loc.Start.Line {
			return a.Loc.Start.Column - b.Loc.Start.Column
		}
		return a.Loc.Start.Line - b.Loc.Start.Line
	})

	return tokens, nil
}

func (l *Lexer) lex(line string) []Token {
	var tokens []Token
	row := line
	for k, v := range TokenDefinition {
		for {
			loc := v.FindStringIndex(row)
			if len(loc) == 0 {
				break
			}
			value := v.FindString(row)
			if loc != nil {
				start := newPosition(
					l.lineNumber+1, // 1-index based on output
					utf8.RuneCountInString(line[:loc[0]])+1, // 1-index based on output
					l.numberOfChar+utf8.RuneCountInString(line[:loc[0]]),
				)
				end := newPosition(
					l.lineNumber+1,                        // 1-index based on output
					utf8.RuneCountInString(line[:loc[1]]), // 1-index based on output
					l.numberOfChar+utf8.RuneCountInString(line[:loc[1]]),
				)
				token := newToken(k, value, start, end)
				tokens = append(tokens, token)
				row = row[0:loc[0]] + strings.Repeat(string(INVALID), loc[1]-loc[0]) + row[loc[1]:]
			}
		}
	}
	return tokens
}
