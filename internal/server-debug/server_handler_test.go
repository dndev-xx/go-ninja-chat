package serverdebug //nolint:testpackage // special hack

import "net/http"

func (s *Server) Handler() http.Handler {
	return s.srv.Handler
}
