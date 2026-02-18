package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
)

type BashParams struct {
	Command string `json:"command"`
}

func Bash(command string) (string, error) {
	fmt.Fprintf(os.Stderr, "Executing command: %s\n", command)
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func RegisterBashTools(reg *registry.Registry) {
	reg.Register("Bash", func(args string) (string, error) {
		fmt.Fprintf(os.Stderr, "DEBUG: Bash tool called with args: %q\n", args)
		var params BashParams
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}
		fmt.Fprintf(os.Stderr, "DEBUG: Parsed command: %s\n", params.Command)
		return Bash(params.Command)
	})
}
