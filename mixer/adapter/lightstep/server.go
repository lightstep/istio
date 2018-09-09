package lightstep

// Server is basic server interface
type Server interface {
	Addr() string
	Close() error
	Run(shutdown chan error)
}

type ServerOptions struct {
	Address string
}

// Addr returns the listening address of the server
func (s *Adapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *Adapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *Adapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}
