package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (s *server) handleHealthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}
}

func (s *server) handleMarker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userZid := getUserZID(s, w, r)
		if userZid == "" {
			return
		}

		switch r.Method {
		case "GET":
			handleGetMarker(s, w, r, userZid)
		case "PUT":
			handlePutMarker(s, w, r, userZid)
		default:
			handleNotSupportedMethod(w, r.Method)
		}
	}
}

func (s *server) handleSingleMarker() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userZid := getUserZID(s, w, r)
		if userZid == "" {
			return
		}

		switch r.Method {
		case "GET":
			handleGetSingleMarker(s, w, r, userZid)
		case "DELETE":
			handleDeleteMarker(s, w, r, userZid)
		default:
			handleNotSupportedMethod(w, r.Method)
		}
	}
}

func getUserZID(s *server, w http.ResponseWriter, r *http.Request) string {
	regex := regexp.MustCompile("^Bearer (.*)")
	rawHeader := r.Header.Get("Authorization")
	matches := regex.FindStringSubmatch(rawHeader)
	if len(matches) <= 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"Could not find Authorization header"}`)
		return ""
	}
	bearerToken := matches[1]

	token, err := jwt.Parse(bearerToken, func(*jwt.Token) (interface{}, error) {
		return s.authKey, nil
	})

	if err != nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"Invalid Token"}`)
		return ""
	}

	claims := token.Claims.(jwt.MapClaims)

	if claims["zid"] == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"Could not find zid in given token"}`)
		return ""
	}

	return fmt.Sprintf("%v", claims["zid"])
}

func handleGetMarker(s *server, w http.ResponseWriter, r *http.Request, userZid string) {
	w.Header().Set("Content-Type", "application/json")
	markers, err := getMarkerCollection(userZid, s.db)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Info("Could not find markers", zap.Error(err))
		fmt.Fprint(w, `{"message":"Could not find markers"}`)
		return
	}

	response, _ := json.Marshal(markers)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(response))
}

func handleGetSingleMarker(s *server, w http.ResponseWriter, r *http.Request, userZid string) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	markers, err := getMarker(userZid, params["id"], s.db)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Info("Could not find markers", zap.Error(err))
		fmt.Fprint(w, `{"message":"Could not find marker"}`)
		return
	}

	response, _ := json.Marshal(markers)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(response))
}

func handlePutMarker(s *server, w http.ResponseWriter, r *http.Request, userZid string) {
	w.Header().Set("Content-Type", "application/json")
	marker, err := getNewMarker(r.Body, userZid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Info("Could not parse given body", zap.Error(err))
		fmt.Fprint(w, `{"message":"Could not parse given body"}`)
		return
	}

	if err = marker.save(s.db); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("Could not insert in database", zap.Error(err))
		fmt.Fprint(w, `{"message":"Could not insert in database"}`)
		return
	}

	response, _ := json.Marshal(marker)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(response))
}

func handleDeleteMarker(s *server, w http.ResponseWriter, r *http.Request, userZid string) {
	params := mux.Vars(r)

	if err := deleteMarker(userZid, params["id"], s.db); err != nil {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Error("Could not insert in database", zap.Error(err))
		fmt.Fprint(w, `{"message":"Could not delete marker"}`)
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleNotSupportedMethod(w http.ResponseWriter, method string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, `{"message":"Method %s is not supported"}`, method)
	return
}

func getNewMarker(body io.ReadCloser, user string) (*Marker, error) {

	var marker Marker

	decoder := json.NewDecoder(body)
	err := decoder.Decode(&marker)

	if err != nil || marker.Lat == 0 || marker.Lng == 0 {
		return nil, errors.New("Invalid marker format")
	}

	marker.User = user

	return &marker, nil
}
