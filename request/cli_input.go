package request

import (
	"bufio"
	"fmt"
	"os"

	"github.com/atolix/clyst/spec"
)

type CLIInput struct{}

func (CLIInput) GetPathParam(p spec.Parameter) string {
	fmt.Printf("Enter %s (%s): ", p.Name, p.Schema.Type)
	var v string
	fmt.Scan(&v)
	return v
}

func (CLIInput) GetQueryParam(p spec.Parameter) string {
	fmt.Printf("Enter %s (%s) [optional]: ", p.Name, p.Schema.Type)
	var v string
	fmt.Scanln(&v)
	return v
}

func (CLIInput) GetRequestBody() string {
	fmt.Println("Enter JSON body:")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}
