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

type Handler struct {
	ContextGenerator ContextGeneratorFunc
	Trace            bool
	handler          http.Handler
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func NewHandler(schema string, resolver any) (*Handler, error) {
	g := &Handler{}

	s, err := graphql.ParseSchema(schema, resolver, graphql.ErrorExtensioner(g.errorTracer))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	g.handler = graphqlws.NewHandlerFunc(s, &relay.Handler{Schema: s}, graphqlws.WithContextGenerator(g)) // support http fallback

	return g, nil
}

func (h *Handler) BuildContext(ctx context.Context, r *http.Request) (context.Context, error) {
	if h.ContextGenerator == nil {
		return ctx, nil
	}

	return h.ContextGenerator(ctx, r)
}

func (h *Handler) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := h.handler.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}

	c, rw, err := hj.Hijack()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return c, rw, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	switch r.Method {
	case "GET", "POST":
		h.handler.ServeHTTP(w, r)
	case "OPTIONS":
		fmt.Fprintf(w, "ok\n")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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
