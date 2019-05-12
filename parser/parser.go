package parser

import (
	"fmt"
)

// Grammar
//
// production
//     : lhs ":" rhs ";"
//     ;
// lhs
//     : id
//     ;
// rhs
//     : alternative ("|" alternative)*
//     ;
// alternative
//     : (id)*
//     ;

type State string

func (s State) String() string {
	return string(s)
}

const (
	stateStart       = State("start")
	stateProduction  = State("production")
	stateLHS         = State("lhs")
	stateRHS         = State("rhs")
	stateAlternative = State("alternative")
)

type AST struct {
	State    State
	Tokens   []Token
	Children []*AST
}

func (ast *AST) appendChild(child *AST) {
	if ast.Children == nil {
		ast.Children = []*AST{}
	}
	ast.Children = append(ast.Children, child)
}

type Frame struct {
	state  State
	tokens []Token
	ast    *AST
}

type Parser interface {
	Parse() (*AST, error)
}

type parser struct {
	lex          Lexer
	peekedTok    Token
	stateStack   []*Frame
	currentState *Frame
	ast          *AST
}

func NewParser(lex Lexer) (Parser, error) {
	if lex == nil {
		return nil, fmt.Errorf("Lexer is nil")
	}

	return &parser{
		lex:        lex,
		peekedTok:  nil,
		stateStack: []*Frame{},
	}, nil
}

func (p *parser) Parse() (ast *AST, err error) {
	defer func() {
		rErr := recover()
		if rErr != nil {
			err = rErr.(error)
			return
		}
	}()

	p.start()

	ast = p.ast
	return
}

func (p *parser) start() {
	p.entry(stateStart)

	for {
		if p.isNext(TokenTypeEOF) {
			break
		}
		p.production()
	}

	p.exit()
}

func (p *parser) production() {
	p.entry(stateProduction)

	p.lhs()
	p.match(TokenTypeColon)
	p.rhs()
	p.match(TokenTypeSemicolon)

	p.exit()
}

func (p *parser) lhs() {
	p.entry(stateLHS)

	p.matchAndPush(TokenTypeID)

	p.exit()
}

func (p *parser) rhs() {
	p.entry(stateRHS)

	for {
		p.alternative()
		if !p.isNext(TokenTypeVBar) {
			break
		}
		p.match(TokenTypeVBar)
	}

	p.exit()
}

func (p *parser) alternative() {
	p.entry(stateAlternative)

	for {
		if !p.isNext(TokenTypeID, TokenTypeString) {
			break
		}
		p.matchAndPush(TokenTypeID, TokenTypeString)
	}

	p.exit()
}

func (p *parser) entry(s State) {
	ast := &AST{
		State: s,
	}
	if p.currentState != nil {
		p.currentState.ast.appendChild(ast)
	}

	f := &Frame{
		state:  s,
		tokens: []Token{},
		ast:    ast,
	}
	p.stateStack = append(p.stateStack, f)
	p.currentState = f
}

func (p *parser) exit() {
	stackLen := len(p.stateStack)
	f := p.stateStack[stackLen-1]
	p.stateStack = p.stateStack[:stackLen-1]

	if stackLen >= 2 {
		p.currentState = p.stateStack[stackLen-2]
	} else {
		p.currentState = nil
	}

	f.ast.Tokens = f.tokens
	p.ast = f.ast
}

func (p *parser) matchAndPush(expected ...TokenType) {
	tok := p.consume(expected...)
	p.currentState.tokens = append(p.currentState.tokens, tok)
}

func (p *parser) match(expected ...TokenType) {
	p.consume(expected...)
}

func (p *parser) isNext(expected ...TokenType) bool {
	tokType := p.peek()
	for _, e := range expected {
		if tokType == e {
			return true
		}
	}

	return false
}

func (p *parser) consume(expected ...TokenType) Token {
	var tok Token
	var err error
	if p.peekedTok != nil {
		tok = p.peekedTok
		p.peekedTok = nil
	} else {
		tok, err = p.lex.Next()
		if err != nil {
			panic(err)
		}
	}
	for _, e := range expected {
		if tok.Type() == e {
			return tok
		}
	}

	panic(fmt.Errorf("unexpected token. want: %v, got: %v", expected, tok.Type()))
}

func (p *parser) peek() TokenType {
	if p.peekedTok == nil {
		tok, err := p.lex.Next()
		if err != nil {
			panic(err)
		}
		p.peekedTok = tok
	}

	return p.peekedTok.Type()
}
