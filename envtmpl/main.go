package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("ERROR: incorrect number of arguments provided.")
		os.Exit(1)
	}

	tmpl, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR: couldn't read input file: %v\n", err)
		os.Exit(1)
	}

	envMap := envToMap()
	t := template.Must(template.New("tmpl").Parse(string(tmpl)))

	output, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Printf("ERROR: couldn't create output file: %v\n", err)
		os.Exit(1)
	}

	t.Execute(output, envMap)
}

func envToMap() map[string]string {
	envMap := make(map[string]string)

	for _, v := range os.Environ() {
		split_v := strings.SplitN(v, "=", 2)
		envMap[split_v[0]] = split_v[1]
	}

	return envMap
}
