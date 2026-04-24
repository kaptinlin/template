// Package main demonstrates registering a custom tag on an Engine.
// External packages can implement the Statement interface directly.
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kaptinlin/template"
)

// SetNode implements template.Statement to support {% set x = expr %}.
type SetNode struct {
	VarName string
	Expr    template.Expression
	Line    int
	Col     int
}

func (n *SetNode) Position() (int, int) { return n.Line, n.Col }
func (n *SetNode) String() string       { return fmt.Sprintf("Set(%s)", n.VarName) }

func (n *SetNode) Execute(renderCtx *template.RenderContext, _ io.Writer) error {
	val, err := n.Expr.Evaluate(renderCtx)
	if err != nil {
		return err
	}
	renderCtx.Set(n.VarName, val.Interface())
	return nil
}

func main() {
	runMain(os.Stdout, log.Fatal)
}

func runMain(out io.Writer, fatal func(...any)) {
	if err := run(out); err != nil {
		fatal(err)
	}
}

func run(out io.Writer) error {
	engine := template.New()

	if err := registerSetTag(engine); err != nil {
		return err
	}

	tpl, err := engine.ParseString(`{% set greeting = "Hello" %}{{ greeting }}, World!`)
	if err != nil {
		return err
	}

	rendered, err := tpl.Render(nil)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, rendered) // Hello, World!
	return err
}

func registerSetTag(engine *template.Engine) error {
	// Register a {% set varname = expr %} tag on this engine.
	return engine.RegisterTag("set", func(_ *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
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
			VarName: varToken.Value,
			Expr:    expr,
			Line:    start.Line,
			Col:     start.Col,
		}, nil
	})
}
