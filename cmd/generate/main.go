package main

import (
	"bufio"
	"bytes"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

// tpl which holds the main template and partial to render the plots.
const tpl = `{{- define "field" }}
{{- range . }}
	{{- template "plot" .}}
{{- end }}
{{- end}}

{{- define "plot" }}
<pre>{{ . }}</pre>
{{- end }}
`

var grasChars = []byte{'"', '.', ',', '\''}

func main() {
	var (
		plotsDir = flag.String("plots.dir", "cmd/generate/fixture", "Direcotry of garden plot files")
		seedFile = flag.String("seed.file", "cmd/generate/seed.html", "Seed file with cool plots")
	)
	flag.Parse()

	seed, err := ioutil.ReadFile(*seedFile)
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(*plotsDir)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})

	plots := []string{}

	for _, f := range files {
		// if f.Size() > 168 {
		// 	continue
		// }

		if filepath.Ext(f.Name()) != ".txt" {
			continue
		}

		path := filepath.Join(*plotsDir, f.Name())
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		isValid, err := isValidPlot(buf)
		if err != nil {
			log.Fatal(err)
		}
		if !isValid {
			continue
		}

		plots = append(plots, string(buf))
	}

	var (
		additionalPlots = 30
		l               = len(plots)
	)

	if l < additionalPlots {
		l = additionalPlots
	}

	var (
		out = bufio.NewWriter(os.Stdout)
		t   = template.Must(template.New("garden").Parse(tpl))
	)

	// Write the templated plots.
	if err := t.ExecuteTemplate(out, "field", plots); err != nil {
		log.Fatal(err)
	}

	// Write the seed to fill up space.
	if _, err := out.Write(seed); err != nil {
		log.Fatal(err)
	}

	// Flush to make sure no bits left behind.
	if err := out.Flush(); err != nil {
		log.Fatal(err)
	}
}

func isValidPlot(content []byte) (bool, error) {
	var (
		r     = bufio.NewReader(bytes.NewReader(content))
		valid = true
		lines = 0
	)

	for {
		if lines > 8 {
			valid = false
			break
		}

		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}

		if utf8.RuneCountInString(string(line)) > 20 {
			valid = false
			break
		}

		lines++
	}

	return valid, nil
}
