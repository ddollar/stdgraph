package stdgraph

import (
	"context"
	_ "embed" // embed
	"fmt"
	"net/http"
	"strings"

	"github.com/ddollar/graphql-go"
	"github.com/ddollar/graphql-go/relay"
	"github.com/pkg/errors"
)

type Handler struct {
	Trace      bool
	handler    http.Handler
	middleware []MiddlewareFunc
}

type MiddlewareFunc func(ctx context.Context, r *http.Request) (context.Context, error)

type contextKey string

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var contextAuthorization = contextKey("authorization")

func Authorization(ctx context.Context, kind string) string {
	prefix := fmt.Sprintf("%s ", kind)

	if v, ok := ctx.Value(contextAuthorization).(string); ok && strings.HasPrefix(v, prefix) {
		return v[len(prefix):]
	} else {
		return ""
	}
}

func NewHandler(schema string, resolver any) (*Handler, error) {
	g := &Handler{
		middleware: []MiddlewareFunc{},
	}

	s, err := graphql.ParseSchema(schema, resolver, graphql.ErrorExtensioner(g.errorTracer))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	g.handler = &relay.Handler{Schema: s}

	return g, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	ctx := context.WithValue(r.Context(), contextAuthorization, r.Header.Get("Authorization"))

	for _, fn := range h.middleware {
		c, err := fn(ctx, r)
		switch et := err.(type) {
		case Error:
			http.Error(w, et.Error(), et.Code())
			return
		case error:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx = c
	}

	r = r.WithContext(ctx)

	switch r.Method {
	case "GET", "POST":
		h.handler.ServeHTTP(w, r)
	case "OPTIONS":
		fmt.Fprintf(w, "ok\n")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Use(fn MiddlewareFunc) {
	h.middleware = append(h.middleware, fn)
}

func (h *Handler) errorTracer(err error) map[string]interface{} {
	if h.Trace {
		if st, ok := err.(stackTracer); ok {
			return map[string]interface{}{
				"stacktrace": st.StackTrace(),
			}
		}
	}

	return nil
}
