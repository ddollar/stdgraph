package stdgraph

import (
	"bufio"
	"context"
	_ "embed" // embed
	"fmt"
	"net"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
	"github.com/pkg/errors"
)

type ContextGeneratorFunc func(context.Context, *http.Request) (context.Context, error)

type Graph struct {
	ContextGenerator ContextGeneratorFunc
	Trace            bool
	handler          http.Handler
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func New(schema string, resolver any) (*Graph, error) {
	g := &Graph{}

	s, err := graphql.ParseSchema(schema, resolver, graphql.ErrorExtensioner(g.errorTracer))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	g.handler = graphqlws.NewHandlerFunc(s, &relay.Handler{Schema: s}, graphqlws.WithContextGenerator(g)) // support http fallback

	return g, nil
}

func (g *Graph) BuildContext(ctx context.Context, r *http.Request) (context.Context, error) {
	if g.ContextGenerator == nil {
		return ctx, nil
	}

	return g.ContextGenerator(ctx, r)
}

func (g *Graph) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := g.handler.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}

	c, rw, err := h.Hijack()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return c, rw, nil
}

func (g *Graph) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	switch r.Method {
	case "GET", "POST":
		g.handler.ServeHTTP(w, r)
	case "OPTIONS":
		fmt.Fprintf(w, "ok\n")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (g *Graph) errorTracer(err error) map[string]interface{} {
	if g.Trace {
		if st, ok := err.(stackTracer); ok {
			return map[string]interface{}{
				"stacktrace": st.StackTrace(),
			}
		}
	}

	return nil
}
