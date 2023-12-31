// Package gqlpq provides a thread-safe implementation of a hot-reloadable,
// schema-verified allowlist of persisted queries.
package gqlpq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	gqlparse "github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	gqlFileExtension       = ".graphql"
	gqlSchemaFileExtension = ".graphqls"
)

type PersistedQueries struct {
	schema *ast.Schema

	list atomic.Value
	// list map[string]string
}

// New reads the schema and creates a new GraphQL persisted queries list instance.
func New(schemaDirPath string) (*PersistedQueries, error) {
	dir, err := os.ReadDir(schemaDirPath)
	if err != nil {
		return nil, fmt.Errorf("reading schema directory path: %w", err)
	}
	schemaSources := []*ast.Source{}
	for _, o := range dir {
		if o.IsDir() {
			continue
		}
		n := o.Name()
		if !strings.HasSuffix(n, gqlSchemaFileExtension) {
			continue
		}
		p := filepath.Join(schemaDirPath, n)
		source, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading schema file %q: %w", p, err)
		}
		schemaSources = append(schemaSources, &ast.Source{
			Name:  n,
			Input: string(source),
		})
	}

	schema, err := gqlparse.LoadSchema(schemaSources...)
	if err != nil {
		return nil, fmt.Errorf("loading schema: %w", err)
	}
	var list atomic.Value
	list.Store(map[string]string{})
	return &PersistedQueries{
		schema: schema,
		list:   list,
	}, nil
}

// GetQuery returns the query by key, or "" if no query is found.
// GetQuery is safe for concurrent use.
func (l *PersistedQueries) GetQuery(key string) string {
	m := l.list.Load().(map[string]string)
	return m[key]
}

// Len returns the length of the list.
// Len is safe for concurrent use.
func (l *PersistedQueries) Len() int {
	m := l.list.Load().(map[string]string)
	return len(m)
}

// ForEach calls fn for every key-query pair stored.
// ForEach is safe for concurrent use.
func (l *PersistedQueries) ForEach(fn func(key, query string)) {
	m := l.list.Load().(map[string]string)
	for k, v := range m {
		fn(k, v)
	}
}

// Load loads the persisted queries from dirPath.
// Load swaps the list atomically and is therefore safe for concurrent use.
func (l *PersistedQueries) Load(dirPath string) error {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("reading query directory path: %w", err)
	}
	newMap := make(map[string]string)
	for _, o := range dir {
		if o.IsDir() {
			continue
		}
		n := o.Name()
		if !strings.HasSuffix(n, gqlFileExtension) {
			continue
		}
		if n != url.PathEscape(n) {
			return fmt.Errorf("invalid file name (not URL safe): %q", n)
		}
		p := filepath.Join(dirPath, n)
		query, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading query file %q: %w", p, err)
		}
		n, _ = strings.CutSuffix(n, gqlFileExtension)

		queryStr := string(query)
		_, errs := gqlparse.LoadQuery(l.schema, queryStr)
		if errs != nil {
			return fmt.Errorf("reading query file %q: %v", p, errs)
		}

		encodedQuery, err := json.Marshal(queryStr)
		if err != nil {
			return fmt.Errorf("encoding query to JSON string: %w", err)
		}
		newMap[n] = string(encodedQuery)
	}

	l.list.Swap(newMap)
	return nil
}

// Watch starts listening to changes on dirPath
// and automatically reloads the persisted queries.
// Watch is safe for concurrent use.
// onReloaded is invoked after a reload.
func (l *PersistedQueries) Watch(
	ctx context.Context,
	schemaDirPath string,
	dirPath string,
	debounce time.Duration,
	onReloaded func(),
) (err error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("initializing watcher: %w", err)
	}
	defer func() {
		err = w.Close()
		if err != nil {
			err = fmt.Errorf("closing watcher %w", err)
		}
	}()
	err = w.Add(dirPath)
	if err != nil {
		return fmt.Errorf("adding watcher dir path %q: %w", dirPath, err)
	}
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done(): // Context canceled
			return ctx.Err()
		case <-timer.C: // Debounce triggered
			err := l.Load(dirPath)
			if err != nil {
				return fmt.Errorf("parsing: %w", err)
			}
			if onReloaded != nil {
				onReloaded()
			}
		case _, ok := <-w.Events: // Allowlist changed
			if !ok {
				return nil
			}
			timer.Reset(debounce)
		case err, ok := <-w.Errors: // Watcher failed
			if !ok {
				return err
			}
		}
	}
}
