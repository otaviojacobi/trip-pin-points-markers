package main

func (s *server) routes() {
	s.router.HandleFunc("/healthcheck", s.handleHealthcheck())
	s.router.HandleFunc("/marker", s.handleMarker())
	s.router.HandleFunc("/marker/{id}", s.handleSingleMarker())
}
