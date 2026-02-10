// Package main demonstrates registering a custom tag via RegisterTag.
// External packages can implement the Statement interface directly.
package main

import (
	"fmt"
	"io"
	"log"

	"github.com/kaptinlin/template"
)

// SetNode implements template.Statement to support {% set x = expr %}.
type SetNode struct {
	VarName    string
	Expression template.Expression
	Line       int
	Col        int
}

func (n *SetNode) Position() (int, int) { return n.Line, n.Col }
func (n *SetNode) String() string       { return fmt.Sprintf("Set(%s)", n.VarName) }

func (n *SetNode) Execute(ctx *template.ExecutionContext, _ io.Writer) error {
	val, err := n.Expression.Evaluate(ctx)
	if err != nil {
		return err
	}
	ctx.Set(n.VarName, val.Interface())
	return nil
}

func main() {
	// Register a {% set varname = expr %} tag.
	err := template.RegisterTag("set", func(_ *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
		varToken, err := arguments.ExpectIdentifier()
		if err != nil {
			return nil, arguments.Error("expected variable name after 'set'")
		}

		if arguments.Match(template.TokenSymbol, "=") == nil {
			return nil, arguments.Error("expected '=' after variable name")
		}

		expr, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}

		if arguments.Remaining() > 0 {
			return nil, arguments.Error("unexpected tokens after expression")
		}

		return &SetNode{
			VarName:    varToken.Value,
			Expression: expr,
			Line:       start.Line,
			Col:        start.Col,
		}, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	output, err := template.Render(`{% set greeting = "Hello" %}{{ greeting }}, World!`, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output) // Hello, World!
}
