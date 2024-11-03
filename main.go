package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Mapping struct {
	VanityDomain string
	RepoBase     string
}

const tmpl = `<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="{{.VanityDomain}}{{.ImportPath}} git https://{{.RepoBase}}{{.ImportPath}}">
</head>
</html>
`

func parseMapping(s string) (Mapping, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Mapping{}, fmt.Errorf("invalid mapping format: %s. Expected format: vanity-domain:repo-base", s)
	}
	return Mapping{
		VanityDomain: strings.TrimSpace(parts[0]),
		RepoBase:     strings.TrimSpace(parts[1]),
	}, nil
}

func main() {
	var mappingsRaw arrayFlags
	flag.Var(&mappingsRaw, "map", "Mapping in the format vanity-domain:repo-base (can be specified multiple times)")

	var port string
	flag.StringVar(&port, "port", "8080", "Port to listen on")

	flag.Parse()

	if len(mappingsRaw) == 0 {
		log.Fatal("At least one mapping must be specified using -map")
	}

	// Parse mappings
	mappings := make(map[string]string) // domain -> repo base
	for _, m := range mappingsRaw {
		mapping, err := parseMapping(m)
		if err != nil {
			log.Fatal(err)
		}
		mappings[mapping.VanityDomain] = mapping.RepoBase
	}

	t := template.Must(template.New("page").Parse(tmpl))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only respond to go-get=1 queries
		if r.URL.Query().Get("go-get") != "1" {
			http.NotFound(w, r)
			return
		}

		// Get the host from headers, handling proxy scenarios
		host := r.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = r.Host
		}

		// Remove port if present
		if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
			host = host[:colonIdx]
		}

		repoBase, ok := mappings[host]
		if !ok {
			http.Error(w, fmt.Sprintf("no mapping found for host: %s", host), http.StatusNotFound)
			return
		}

		// Get the import path from the URL
		importPath := strings.TrimSuffix(r.URL.Path, "/")

		data := struct {
			VanityDomain string
			RepoBase     string
			ImportPath   string
		}{
			VanityDomain: host,
			RepoBase:     repoBase,
			ImportPath:   importPath,
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Printf("Starting server on port %s", port)
	for domain, repo := range mappings {
		log.Printf("  %s -> %s", domain, repo)
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// arrayFlags allows multiple flag values
type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
