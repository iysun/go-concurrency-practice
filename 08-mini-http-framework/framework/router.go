package framework

import (
	"strings"
)

// node is a trie node for the path router.
// Each segment of a URL path (split by "/") is one node level.
//
// Example trie for routes GET /user/:id and GET /user/profile:
//
//	root
//	 └── "user"
//	      ├── ":id"     (wildcard, matches any segment)
//	      └── "profile" (static, takes priority over wildcard)
type node struct {
	part     string           // path segment this node represents, e.g. "user" or ":id"
	children map[string]*node // static children, keyed by part
	wildcard *node            // child for ":param" wildcard (at most one per node)
	handlers []HandlerFunc    // non-nil means this node is a registered route endpoint
}

func newNode(part string) *node {
	return &node{part: part, children: make(map[string]*node)}
}

// insert registers handlers under the given path pattern.
// Path parameters are prefixed with ":", e.g. "/user/:id/posts".
// TODO: implement trie insertion
func (n *node) insert(pattern string, handlers []HandlerFunc) {
	parts := splitPath(pattern)
	cur := n
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			if cur.wildcard == nil {
				cur.wildcard = newNode(part)
			}
			cur = cur.wildcard
		} else {
			if _, ok := cur.children[part]; !ok {
				cur.children[part] = newNode(part)
			}
			cur = cur.children[part]
		}
	}
	cur.handlers = handlers
}

// search finds the node matching path and populates params.
// Returns nil if no route matches.
// TODO: implement trie search (static match takes priority over wildcard)
func (n *node) search(path string, params map[string]string) *node {
	parts := splitPath(path)
	cur := n
	for _, part := range parts {
		if child, ok := cur.children[part]; ok {
			cur = child
		} else if cur.wildcard != nil {
			// wildcard match — capture the value
			params[cur.wildcard.part[1:]] = part
			cur = cur.wildcard
		} else {
			return nil
		}
	}
	if cur.handlers == nil {
		return nil
	}
	return cur
}

func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// router maps HTTP methods to their own trie root.
type router struct {
	trees map[string]*node // method -> trie root
}

func newRouter() *router {
	return &router{trees: make(map[string]*node)}
}

func (r *router) addRoute(method, pattern string, handlers []HandlerFunc) {
	if _, ok := r.trees[method]; !ok {
		r.trees[method] = newNode("/")
	}
	r.trees[method].insert(pattern, handlers)
}

// match looks up handlers for (method, path) and fills params.
func (r *router) match(method, path string) ([]HandlerFunc, map[string]string) {
	params := make(map[string]string)
	root, ok := r.trees[method]
	if !ok {
		return nil, nil
	}
	n := root.search(path, params)
	if n == nil {
		return nil, nil
	}
	return n.handlers, params
}
