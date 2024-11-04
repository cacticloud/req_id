package reqid

import (
	"net/http"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	nanoid "github.com/matoous/go-nanoid/v2"
)

func init() {
	caddy.RegisterModule(ReqID{})
	httpcaddyfile.RegisterHandlerDirective("req_id", parseCaddyfile)
}

type ReqID struct {
	Length     int            `json:"length"`
	Additional map[string]int `json:"additional,omitempty"`
}

func (ReqID) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.req_id",
		New: func() caddy.Module { return new(ReqID) },
	}
}

func (m *ReqID) Provision(ctx caddy.Context) error {
	if m.Length < 1 {
		m.Length = 21
	}

	if m.Additional == nil {
		m.Additional = make(map[string]int)
	}

	return nil
}

func (m ReqID) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	id := nanoid.Must(m.Length)
	repl.Set("http.request_id", id)

	for key, value := range m.Additional {
		id := nanoid.Must(value)
		repl.Set("http.request_id."+key, id)
	}
	// r.Header.Set("Req-ID", reqID)
	// w.Header().Set("Req-ID", reqID)

	// newContext := context.WithValue(r.Context(), "Req-ID", reqID)
	// newRequest := r.WithContext(newContext)

	return next.ServeHTTP(w, r)
}

func (m *ReqID) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	arg1 := d.NextArg()
	arg2 := d.NextArg()

	// Parse standalone length
	if arg1 && arg2 {
		val := d.Val()
		len, err := strconv.Atoi(val)

		if err != nil {
			return d.Err("failed to convert length to int")
		}

		if len < 1 {
			return d.Err("length cannot be less than 1")
		}

		m.Length = len
	}

	if m.Additional == nil {
		m.Additional = make(map[string]int)
	}

	// Parse additional IDs
	for d.NextBlock(0) {
		key := d.Val()
		if !d.NextArg() {
			return d.ArgErr()
		}

		val := d.Val()
		len, err := strconv.Atoi(val)

		if err != nil {
			return d.Err("failed to convert length to int")
		}

		if len < 1 {
			return d.Err("length cannot be less than 1")
		}

		if _, ok := m.Additional[key]; ok {
			return d.Errf("duplicate key: %v\n", key)
		}

		m.Additional[key] = len
	}

	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	m := new(ReqID)
	err := m.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ReqID)(nil)
	_ caddyhttp.MiddlewareHandler = (*ReqID)(nil)
	_ caddyfile.Unmarshaler       = (*ReqID)(nil)
)
