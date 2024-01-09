package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for scanner.Scan() {
		csses := ParseString(scanner.Text())
		s := OrderedCSSArrToString(csses)
		_, err := writer.WriteString(s)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		err = writer.Flush()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}
}
