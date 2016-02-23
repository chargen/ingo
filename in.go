package ingo

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	obsoleteKeys = make(map[string]string)
)

func Parse(configPath string) {
	if flag.Parsed() {
		log.Fatalf("flags have been parsed already.")
	}

	loadConfig(configPath)
	saveConfig(configPath)
	flag.Parse()
}

func loadConfig(configPath string) {
	fin, err := os.Open(configPath)
	if _, ok := err.(*os.PathError); ok {
		log.Printf("No config file found. Creating %s ...", configPath)
		return
	} else if err != nil {
		log.Fatalf("Unable to read config file %v: %v", configPath, err)
	}
	defer fin.Close()

	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		// find first assignment symbol and parse key, val
		i := strings.IndexAny(line, "=:")
		if i == -1 {
			continue
		}
		key, val := strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:])

		if err := flag.Set(key, val); err != nil {
			obsoleteKeys[key] = val
			continue
		}
	}
}

func saveConfig(configPath string) {
	fout, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("Unable to open file %v for writing: %v", configPath, err)
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)
	defer writer.Flush()
	flag.VisitAll(func(f *flag.Flag) {
		_, usage := flag.UnquoteUsage(f)
		fmt.Fprintln(writer, "#", strings.Replace(usage, "\n    \t", "\n# ", -1))
		fmt.Fprintf(writer, "%v=%v\n", f.Name, f.Value.String())
	})

	// if we have obsolete keys left from the old config, preserve them in an
	// additional section at the end of the file
	if len(obsoleteKeys) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!")
	fmt.Fprintln(os.Stderr, "! WARNING: The application was probably updated,")
	fmt.Fprintln(os.Stderr, "! Check and update", configPath, " as necessary and")
	fmt.Fprintln(os.Stderr, "! remove the last \"deprecated\" paragraph to disable this message!")
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!")
	fmt.Fprintln(writer, "\n\n# The following options are probably deprecated and not used currently!")
	for key, val := range obsoleteKeys {
		fmt.Fprintf(writer, "%v=%v\n", key, val)
	}
}
