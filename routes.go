package main

func (s *server) routes() {

	s.router.HandleFunc("/healthcheck", s.handleHealthcheck()).Methods("GET")
	s.router.HandleFunc("/pingDB", s.handlePingDB()).Methods("GET")

	s.router.HandleFunc("/marker", s.handleGetAllMarkers()).Methods("GET")
	s.router.HandleFunc("/marker", s.handleInsertMarker()).Methods("PUT")
	s.router.HandleFunc("/marker/{id}", s.handleGetSingleMarker()).Methods("GET")
	s.router.HandleFunc("/marker/{id}", s.handleDeleteMarker()).Methods("DELETE")

}
