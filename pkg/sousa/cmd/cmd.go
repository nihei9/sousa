package cmd

import (
	"os"

	"github.com/nihei9/sousa/ast2grammar"
	"github.com/nihei9/sousa/grammar"
	"github.com/nihei9/sousa/parser"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sousa",
		Short:         "Sousa is a parsing table generator",
		Long:          `Sousa is a parsing table generator.`,
		Args:          cobra.ExactArgs(1),
		RunE:          run,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	filepath := args[0]
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	lexer := parser.NewLexer(file)
	parser, err := parser.NewParser(lexer)
	if err != nil {
		return err
	}
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	g, err := ast2grammar.Convert(ast)
	if err != nil {
		return err
	}
	first, err := grammar.GenerateFirstSets(g.Productions)
	if err != nil {
		return err
	}
	follow, err := grammar.GenerateFollowSets(g.Productions, first)
	if err != nil {
		return err
	}
	automaton, err := grammar.GenerateLR0Automaton(g.SymbolTable, g.Productions, g.AugmentedStartSymbol)
	if err != nil {
		return err
	}
	parsingTable, err := grammar.GenerateSLRParsingTable(automaton, follow)
	if err != nil {
		return err
	}

	prodsFile, err := os.OpenFile("production", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer prodsFile.Close()
	prodsWriter := NewProductionsWriter(g.Productions)
	prodsWriter.Write(prodsFile)

	actionFile, err := os.OpenFile("action", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer actionFile.Close()
	actionWriter := NewActionWriter(parsingTable, g.Productions)
	actionWriter.Write(actionFile)

	gotoFile, err := os.OpenFile("goto", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer gotoFile.Close()
	gotoWriter := NewGoToWriter(parsingTable)
	gotoWriter.Write(gotoFile)

	return nil
}
