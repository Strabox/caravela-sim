package cli

import (
	"fmt"
	"github.com/urfave/cli"
)

// printBanner prints the banner of the CARAVELA Simulator system.
func printBanner(_ *cli.Context) error {
	fmt.Printf("##################################################################\n")
	fmt.Printf("#      CARAVELA: A Cloud @ Edge (SIMULATOR)         000000       #\n")
	fmt.Printf("#            author: %s                 00000000000     #\n", author)
	fmt.Printf("#  Email: %s           | ||| |      #\n", email)
	fmt.Printf("#              IST/INESC-ID                        || ||| ||     #\n")
	fmt.Printf("##################################################################\n")
	return nil
}
