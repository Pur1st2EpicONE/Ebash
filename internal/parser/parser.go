// Package parser parses a command line into a pipeline of Pipe structures.
// It handles conditional operators (&&, ||), pipes (|), and simple
// redirections (<, >). The parser produces a slice of Pipe values that the
// shell executor can run sequentially.
package parser

import (
	"os"
	"strings"
)

// Pipe represents a single conditional block of commands within a shell pipeline.
// Section holds the commands with their arguments, Input and Output handle optional redirections,
// and NextAnd / NextOr indicate conditional execution of the following pipe.
type Pipe struct {
	Section [][]string // Commands (with arguments) forming this conditional pipe section
	Input   *os.File   // Optional input redirection file
	Output  *os.File   // Optional output redirection file
	NextAnd bool       // True if the next pipe runs only if this one succeeds
	NextOr  bool       // True if the next pipe runs only if this one fails
}

// Parse takes a raw command-line string and converts it into a slice of Pipe
// structures. It expands environment variables, normalizes spacing around
// operators, splits the input by conditional operators (&& and ||), and then
// builds each pipe section (handling pipes and redirections). Returns an
// error when building a section or opening redirection files fails.
func Parse(line string) ([]Pipe, error) {

	line = os.ExpandEnv(line)
	line = strings.NewReplacer("&&", " && ", "||", " || ", ">", " > ", "<", " < ").Replace(line)

	var pipeline []Pipe
	var nextAnd, nextOr bool

	conditionals := splitByConditionals(line)

	for i := 0; i < len(conditionals); i++ {

		conditional := conditionals[i]

		if conditional == "" || conditional == "&&" || conditional == "||" {
			continue
		}

		if i+1 < len(conditionals) {
			switch conditionals[i+1] {
			case "&&":
				nextAnd = true
			case "||":
				nextOr = true
			}
		}

		section, input, output, err := buildSection(conditional)
		if err != nil {
			return nil, err
		}

		pipeline = append(pipeline, Pipe{
			Section: section,
			Input:   input,
			Output:  output,
			NextAnd: nextAnd,
			NextOr:  nextOr,
		})

		if nextAnd || nextOr {
			nextAnd, nextOr = false, false
			i++
		}

	}

	return pipeline, nil
}

// splitByConditionals scans the line and splits it into a slice where each
// element is either a conditional operator ("&&" or "||") or the text
// between operators. It preserves ordering and trims whitespace only when
// producing the final slice element.
func splitByConditionals(line string) []string {

	var conditionals []string
	var builder strings.Builder

	for currByte := 0; currByte < len(line); currByte++ {

		if currByte < len(line)-1 && line[currByte] == '&' && line[currByte+1] == '&' {
			saveWithOperator(&builder, "&&", &conditionals, &currByte)
			continue
		} else if currByte < len(line)-1 && line[currByte] == '|' && line[currByte+1] == '|' {
			saveWithOperator(&builder, "||", &conditionals, &currByte)
			continue
		}

		builder.WriteByte(line[currByte])

	}

	conditionals = append(conditionals, strings.TrimSpace(builder.String()))

	return conditionals

}

// saveWithOperator flushes the current builder contents into the
// conditionals slice (if non-empty), appends the operator token, and advances
// the cursor (currByte) to account for the two-character operator.
func saveWithOperator(builder *strings.Builder, operator string, conditionals *[]string, currByte *int) {
	if builder.Len() > 0 {
		*conditionals = append(*conditionals, strings.TrimSpace(builder.String()))
		builder.Reset()
	}
	*conditionals = append(*conditionals, operator)
	(*currByte)++
}

// buildSection takes a conditional string (a part of the input without &&/||)
// and splits it by pipe symbols (|) to produce the section (list of
// commands). It also recognizes input (<) redirection for the first command
// and output (>) redirection for the last command, opening files accordingly
// and returning them alongside the parsed command arguments.
func buildSection(conditional string) ([][]string, *os.File, *os.File, error) {

	var err error
	var section [][]string
	var input, output *os.File

	commands := strings.Split(conditional, "|")

	for i, command := range commands {

		cmdWithArgs := strings.Fields(strings.TrimSpace(command))
		if len(cmdWithArgs) == 0 {
			continue
		}

		if i == 0 && strings.Contains(command, "<") {
			input, cmdWithArgs, err = redirect(cmdWithArgs, "<")
			if err != nil {
				return nil, nil, nil, err
			}
		}

		if i == len(commands)-1 && strings.Contains(command, ">") {
			output, cmdWithArgs, err = redirect(cmdWithArgs, ">")
			if err != nil {
				return nil, nil, nil, err
			}
		}

		section = append(section, cmdWithArgs)

	}

	return section, input, output, nil

}

// redirect searches cmdWithArgs for the redirection operator (`<` or `>`),
// opens the referenced file (for reading or creating), removes the
// redirection tokens from the argument slice, and returns the opened file
// and the cleaned arguments. If no redirection operator is found, the
// original arguments are returned with a nil file.
func redirect(cmdWithArgs []string, direction string) (*os.File, []string, error) {

	for i := range cmdWithArgs {

		if cmdWithArgs[i] == direction && i+1 < len(cmdWithArgs) {

			var err error
			var file *os.File

			if direction == ">" {
				file, err = os.Create(cmdWithArgs[i+1])
			} else {
				file, err = os.Open(cmdWithArgs[i+1])
			}
			if err != nil {
				return nil, nil, err
			}

			argsWithoutRedirect := make([]string, 0)
			argsWithoutRedirect = append(argsWithoutRedirect, cmdWithArgs[:i]...)
			argsWithoutRedirect = append(argsWithoutRedirect, cmdWithArgs[i+2:]...)

			return file, argsWithoutRedirect, nil

		}

	}

	return nil, cmdWithArgs, nil

}
