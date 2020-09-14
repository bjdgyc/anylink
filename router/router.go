package router

import (
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
)

const (
	ANY = "ANY" // 包含所有 method
)

type HttpMux struct {
	no http.Handler // NotFoundHandler
	mu sync.RWMutex
	m  map[string]muxEntry // example: GET/index:muxEntry{}
	es []muxEntry          // 模糊匹配，pattern需要添加后缀 *
}

type muxEntry struct {
	h       http.Handler
	pattern string
	method  string
}

func NewHttpMux() *HttpMux {
	http.NewServeMux()
	return &HttpMux{
		m:  make(map[string]muxEntry),
		es: make([]muxEntry, 0),
	}
}

func (mux *HttpMux) SetNotFound(no http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.no = no
}

func (mux *HttpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h := mux.match(r.Method, r.URL.Path)
	h.ServeHTTP(w, r)
}

func (mux *HttpMux) match(method, rpath string) http.Handler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	path := mux.cleanPath(rpath)
	// any 路径 匹配
	p_a := ANY + path
	if v, ok := mux.m[p_a]; ok {
		return v.h
	}
	// method 路径 匹配
	method = strings.ToUpper(method)
	p_m := method + path
	if e, ok := mux.m[p_m]; ok {
		return e.h
	}

	// Check for longest valid match.  mux.es contains all patterns
	// that end in / sorted from longest to shortest.
	for _, e := range mux.es {
		// trim last *
		pattern := e.pattern[:len(e.pattern)-1]
		// fmt.Println(pattern, p_a, p_m)
		if strings.HasPrefix(p_a, pattern) {
			return e.h
		}
		if strings.HasPrefix(p_m, pattern) {
			return e.h
		}
	}

	if mux.no != nil {
		return mux.no
	}
	return http.NotFoundHandler()
}

func (mux *HttpMux) cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

func (mux *HttpMux) HandleFunc(method, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(method, pattern, http.HandlerFunc(handler))
}

func (mux *HttpMux) Handle(method, pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" || method == "" {
		panic("http: invalid pattern")
	}
	if handler == nil {
		panic("http: nil handler")
	}
	method = strings.ToUpper(method)
	p := method + pattern
	if _, exist := mux.m[p]; exist {
		panic("http: multiple registrations for " + p)
	}

	e := muxEntry{h: handler, pattern: p}
	mux.m[p] = e
	if pattern[len(pattern)-1] == '*' {
		mux.es = mux.appendSorted(mux.es, e)
	}
}

func (mux *HttpMux) appendSorted(es []muxEntry, e muxEntry) []muxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return len(es[i].pattern) < len(e.pattern)
	})
	if i == n {
		return append(es, e)
	}
	// we now know that i points at where we want to insert
	es = append(es, muxEntry{}) // try to grow the slice in place, any entry works.
	copy(es[i+1:], es[i:])      // Move shorter entries down
	es[i] = e
	return es
}

// ANY /static/* /var/www
func (mux *HttpMux) ServeFile(method, pattern string, root http.FileSystem) {
	fs := http.FileServer(root)

	// trim *
	pt := pattern[:len(pattern)-1]
	mux.HandleFunc(method, pattern, func(w http.ResponseWriter, r *http.Request) {
		// 过滤前缀路径
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pt)
		fs.ServeHTTP(w, r)
	})
}
