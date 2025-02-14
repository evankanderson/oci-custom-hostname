package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Proxy struct {
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	if r.URL.Path != "/v2/" &&
		!strings.HasPrefix(r.URL.Path, "/v2/img") &&
		!strings.HasPrefix(r.URL.Path, "/v2/stacklok/codegate") {
		w.WriteHeader(404)
		return
	}

	newUrl := r.URL
	newUrl.Host = "ghcr.io"
	newUrl.Scheme = "https"

	if strings.HasPrefix(newUrl.Path, "/v2/img") {
		newUrl.Path = "/v2/stacklok/codegate" + newUrl.Path[7:]
	}

	fmt.Printf("Headers: %+v\n", r.Header)

	// We might want to be able to redirect blobs to a registry, but it turns out
	// that existing container registries will expect an "authorization" header, which
	// is cleared when redirecting.
	//
	// if strings.Contains(newUrl.Path, "/blobs/sha256:") {
	// 	 fmt.Printf("Redirect to %s\n", newUrl)
	// 	 // Redirect to blobs/manifests to reduce bandwidth costs
	// 	 w.Header().Add("Location", newUrl.String())
	// 	 w.WriteHeader(http.StatusFound)
	// 	 return
	// }

	req, err := http.NewRequest(r.Method, newUrl.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header

	fmt.Printf("Proxying to %s\n", newUrl)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("Error fetching %s: %v\n", newUrl, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Printf("Got a %d for %s: %+v\n", resp.StatusCode, newUrl, resp.Header)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if written, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body after %d bytes: %v\n", written, err)
	}
}
