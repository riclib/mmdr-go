// Command mmdr-demo renders a Mermaid diagram to SVG using github.com/riclib/mmdr-go.
//
// It reads Mermaid source from a file argument or stdin and writes SVG to a
// file (-o) or stdout. It exists to demonstrate the library and to give a quick
// way to eyeball rendered output.
//
//	# from a file to stdout
//	mmdr-demo diagram.mmd
//
//	# from stdin to a file
//	echo 'flowchart LR; A-->B-->C' | mmdr-demo -o out.svg
//
//	# print the upstream renderer version
//	mmdr-demo -version
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	mmdr "github.com/riclib/mmdr-go"
)

func main() {
	os.Exit(run())
}

func run() int {
	out := flag.String("o", "", "write SVG to this file instead of stdout")
	showVersion := flag.Bool("version", false, "print the upstream mermaid-rs-renderer version and exit")
	timing := flag.Bool("t", false, "print render time to stderr")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(mmdr.Version())
		return 0
	}

	source, srcName, err := readSource(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "mmdr-demo:", err)
		return 1
	}

	start := time.Now()
	svg, err := mmdr.Render(source)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mmdr-demo: rendering %s: %v\n", srcName, err)
		// Distinguish a bad diagram (user error) from an internal panic.
		if errors.Is(err, mmdr.ErrInvalidInput) {
			return 2
		}
		return 3
	}

	if err := writeSVG(*out, svg); err != nil {
		fmt.Fprintln(os.Stderr, "mmdr-demo:", err)
		return 1
	}

	if *timing {
		fmt.Fprintf(os.Stderr, "rendered %s in %v (%d bytes, mmdr %s)\n",
			srcName, elapsed, len(svg), mmdr.Version())
	}
	return 0
}

// readSource returns the Mermaid source from the first file argument, or from
// stdin when no argument is given.
func readSource(args []string) (source, name string, err error) {
	if len(args) > 1 {
		return "", "", fmt.Errorf("expected at most one input file, got %d", len(args))
	}
	if len(args) == 1 && args[0] != "-" {
		b, err := os.ReadFile(args[0])
		if err != nil {
			return "", "", err
		}
		return string(b), args[0], nil
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("reading stdin: %w", err)
	}
	if len(b) == 0 {
		return "", "", errors.New("no input (give a file argument or pipe Mermaid source to stdin)")
	}
	return string(b), "<stdin>", nil
}

func writeSVG(path, svg string) error {
	if path == "" {
		_, err := io.WriteString(os.Stdout, svg)
		return err
	}
	return os.WriteFile(path, []byte(svg), 0o644)
}

func usage() {
	fmt.Fprintf(os.Stderr, `mmdr-demo — render a Mermaid diagram to SVG

usage:
  mmdr-demo [flags] [input.mmd]

  Reads Mermaid source from the file argument, or from stdin if none is given
  (or if the argument is "-"). Writes SVG to stdout, or to -o if set.

flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
examples:
  mmdr-demo diagram.mmd > diagram.svg
  echo 'flowchart LR; A-->B-->C' | mmdr-demo -o out.svg -t
  mmdr-demo -version

exit codes:
  0 ok   1 I/O or usage error   2 invalid diagram source   3 renderer panic
`)
}
