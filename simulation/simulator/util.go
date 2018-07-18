package simulator

import "fmt"

const author = "André Pires"
const email = "pardal.pires@tecnico.ulisboa.pt"

/*
Prints the banner of the CARAVELA Simulator system.
*/
func PrintSimulatorBanner() {
	fmt.Printf("##################################################################\n")
	fmt.Printf("#      CARAVELA: A Cloud @ Edge (SIMULATOR)         000000       #\n")
	fmt.Printf("#            Author: %s                 00000000000     #\n", author)
	fmt.Printf("#  Email: %s           | ||| |      #\n", email)
	fmt.Printf("#              IST/INESC-ID                        || ||| ||     #\n")
	fmt.Printf("##################################################################\n")
}
