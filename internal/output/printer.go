package output

import (
	"fmt"
	"os"
)

type Printer struct {
	JSON  bool
	Quiet bool
}

func NewPrinter(jsonOut, quiet bool) *Printer {
	return &Printer{JSON: jsonOut, Quiet: quiet}
}

func (p *Printer) PrintJSON(data any) error {
	return WriteJSON(os.Stdout, Success(data))
}

func (p *Printer) PrintJSONError(code int, message string) error {
	return WriteJSON(os.Stderr, Failure(code, message))
}

func (p *Printer) PrintJSONErrorDetailed(code int, reason, message, hint string) error {
	return WriteJSON(os.Stderr, FailureDetailed(code, reason, message, hint))
}

func (p *Printer) Println(line string) {
	fmt.Fprintln(os.Stdout, line)
}

func (p *Printer) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}
