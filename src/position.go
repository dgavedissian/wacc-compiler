package main

type Position struct {
	name   string
	line   int
	column int
	length int
}

func (p *Position) Name() string {
	return p.name
}
func (p *Position) Line() int {
	return p.line
}
func (p *Position) Column() int {
	return p.column
}
func (p *Position) End() *Position {
	if p.length == 0 {
		return p
	}
	return p.Add(p.length)
}
func (p *Position) Add(x int) *Position {
	return &Position{
		name:   p.name,
		line:   p.line,
		column: p.column + x,
		length: 0,
	}
}
func ComparePositions(a, b *Position) int {
	if a.name > b.name {
		return 1
	} else if a.name < b.name {
		return -1
	}
	if a.line > b.line {
		return 1
	} else if a.line < b.line {
		return -1
	}
	if a.column > b.column {
		return 1
	} else if a.column < b.column {
		return -1
	}
	return 0
}

type Pos *Position

func NewPositionFromLexer(l *Lexer) *Position {
	return &Position{
		name:   "<unknown>",
		line:   l.Line() + 1,
		column: l.Column() + 1,
		length: len(l.Text()),
	}
}
