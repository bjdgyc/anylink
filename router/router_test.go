// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Server unit tests

package router

import (
	"fmt"
	"net/http"
	"testing"
)

func BenchmarkServerMatch(b *testing.B) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}
	mux := NewHttpMux()
	mux.HandleFunc("GET", "/", fn)
	mux.HandleFunc("GET", "/index", fn)
	mux.HandleFunc("GET", "/home", fn)
	mux.HandleFunc("GET", "/about", fn)
	mux.HandleFunc("GET", "/contact", fn)
	mux.HandleFunc("GET", "/robots.txt", fn)
	mux.HandleFunc("GET", "/products/", fn)
	mux.HandleFunc("GET", "/products/1", fn)
	mux.HandleFunc("GET", "/products/2", fn)
	mux.HandleFunc("GET", "/products/3", fn)
	mux.HandleFunc("GET", "/products/3/image.jpg", fn)
	mux.HandleFunc("GET", "/admin", fn)
	mux.HandleFunc("GET", "/admin/products/", fn)
	mux.HandleFunc("GET", "/admin/products/create", fn)
	mux.HandleFunc("GET", "/admin/products/update", fn)
	mux.HandleFunc("GET", "/admin/products/delete", fn)

	paths := []string{"/", "/notfound", "/admin/", "/admin/foo", "/contact", "/products",
		"/products/", "/products/3/image.jpg"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		if h := mux.match("GET", path); h == nil {
			b.Error("impossible", path)
		}
	}
	b.StopTimer()
}
