package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// errNotConfirmed is returned when the user did not confirm a destructive
// action. It makes the command exit non-zero, so a script that runs without a
// terminal does not take the skipped action for a success.
var errNotConfirmed = errors.New("aborted: not confirmed (type 'yes' at the prompt, or pass --force to skip it)")

// confirm prints question and asks the user to type "yes". Unless --force is
// set, any other answer, including none because stdin is closed or not a
// terminal, returns errNotConfirmed.
func confirm(cmd *cobra.Command, question string) error {
	if force, _ := cmd.Flags().GetBool("force"); force {
		return nil
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, question)
	_, _ = fmt.Fprint(out, "Type 'yes' to confirm: ")

	answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if strings.TrimSpace(answer) != "yes" {
		if !strings.HasSuffix(answer, "\n") {
			// End the prompt line, so the error starts on a line of its own.
			_, _ = fmt.Fprintln(out)
		}
		return errNotConfirmed
	}
	return nil
}
