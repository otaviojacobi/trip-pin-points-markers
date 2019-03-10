package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

const key = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAnntAL976IVa75LVQoa9maOlFFclICOpf4DowXrUlXMyYi0HvEzFxpsk8ZyHg
AUxxQ+MV5qUZvSFYghtldF0PpyJkdJb9E8pUz/oM+/XBgjXHxOXNAn/fghrtOswZ5gqUjne/
kMnjg3wiUPIwlAf+lL1JlRvX0X9iwFjrFshbU/+ExdngPRyC8slKOUQa+cOgnie2oOC1bYgs
BCFITaM4AfKaRs+vgDLfRUHU9DUAzrJ62nJLeauFZ0M54s9pH8RXSB3BECjbEyDhk9wIRKTp
I9sxOpIDBy5S0WktkoxkN+O6kE8t7pXqOD1o7P+rhcXmYYvRCLnUZPRhPOmO7aCqqQIDAQAB
-----END RSA PUBLIC KEY-----`

const stubAuthHeader = `Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsidXNlciJdLCJ6aWQiOiJzdHJpbmczIiwiaWF0IjoxNTUyMjI5NzEwLCJleHAiOjE1NTIyNzI5MTB9.Wq8Zy8RAAFW1IHlSQxUvpCi3rDiDvNB4b5LDaKPKJgVdbD-NJOAu66ZcXZIEpnwxDkNj2fq8xqBaY1QxVPK6rDX4eeNiEThd5OLDHnUeGa0ySb-obBsWyBLaoH3-51DVkHwrm8T7zsHExCt54WRfF4XyS5TADZmd4NB3eNjzfdMea1VdUGo71G9eX3JW4VsXivv0KxsaaJ1yYdYoLDga5NeibGizYhBX87zSCG64rUrU1clCMSiKqYOBNlQQxEoBylQl96ImHG89o60uVtrQlMRK4ip6P3AehnUGLk4Cj_Fv467M4QqrxuT3vo1gjgslrWvL8H2m4CVy2NSRZpPkag`

func getMockServer() (*server, sqlmock.Sqlmock) {

	block, _ := pem.Decode([]byte(key))
	authKey, _ := x509.ParsePKCS1PublicKey(block.Bytes)
	zapLogger, _ := zap.NewProduction()
	db, mock, _ := sqlmock.New()

	s := &server{
		router:  mux.NewRouter(),
		logger:  zapLogger,
		authKey: authKey,
		db:      db,
	}

	s.routes()
	return s, mock
}

func TestHandleHealthcheck(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("GET", "/healthcheck", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleHealthcheck()
	fun(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OK", res.Body.String())
}

func TestInsertNewMarker(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":2.32, "lng":5.55}`))

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.ExpectExec("INSERT INTO markers").WithArgs("string3", 2.32, 5.55, "").WillReturnResult(sqlmock.NewResult(1, 1))
	fun := s.handleInsertMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusCreated, res.Code)
	assert.Equal(t, res.Body.String(), `{"user":"string3","lat":2.32,"lng":5.55,"note":""}`)
}

func TestInsertOnDbFail(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":2.32, "lng":5.55}`))

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.ExpectExec("INSERT INTO markers").WithArgs("string3", 2.32, 5.55, "").WillReturnError(errors.New("test error"))
	fun := s.handleInsertMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not insert in database"}`)
}

func TestInsertNewMarkerNoLatLng(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)

}
func TestInsertNewMarkerNoLat(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lng":3.2}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)
}

func TestInsertNewMarkerNoLng(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":5.1}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)
}

func TestInsertNewMarkerZeroLatLng(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":0,"lng":0}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)
}

func TestInsertNewMarkerZeroLat(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":0,"lng":3.5}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)
}

func TestInsertNewMarkerZeroLng(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":1.2,"lng":0}`))
	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not parse given body"}`)

}

func TestInsertNoAuthHeader(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":1.2,"lng":1.1}`))
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not find Authorization header"}`)
}

func TestInsertInvalidAuthHeader(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("PUT", "/marker", strings.NewReader(`{"lat":1.2,"lng":1.1}`))
	req.Header.Set("Authorization", "Bearer 1234")
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleInsertMarker()
	fun(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Invalid Token"}`)
}

func TestGetAllMarkers(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("GET", "/marker", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "username", "lat", "long", "note"}).
		AddRow(1, "string3", 3.21, 5.2, "teste").
		AddRow(2, "string3", -2.5, -5.2, "")

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs("string3").
		WillReturnRows(rows)

	fun := s.handleGetAllMarkers()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, res.Body.String(), `{"markers":[{"user":"string3","lat":3.21,"lng":5.2,"note":"teste"},{"user":"string3","lat":-2.5,"lng":-5.2,"note":""}]}`)
}

func TestGetAllMarkersDBError(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("GET", "/marker", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs("string3").
		WillReturnError(errors.New("test error"))

	fun := s.handleGetAllMarkers()
	fun(res, req)

	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not find markers"}`)
}

func TestGetSingleMarker(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("GET", "/marker/2/3", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	rows := sqlmock.NewRows([]string{"id", "username", "lat", "long", "note"}).
		AddRow(2, "string3", -2.5, -5.2, "")

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(rows)

	fun := s.handleGetSingleMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, `{"user":"string3","lat":-2.5,"lng":-5.2,"note":""}`, res.Body.String())
}

func TestGetSingleMarkerDBError(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("GET", "/marker/2", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnError(errors.New("test error"))

	fun := s.handleGetSingleMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, `{"message":"Could not find marker"}`, res.Body.String())
}

func TestDeleteMarker(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("DELETE", "/marker/2/3", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.
		ExpectExec("DELETE").
		WillReturnResult(sqlmock.NewResult(1, 1))

	fun := s.handleDeleteMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusNoContent, res.Code)
}

func TestDeleteMarkerWithNotFoundMarker(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("DELETE", "/marker/2/3", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.
		ExpectExec("DELETE").
		WillReturnResult(sqlmock.NewResult(0, 0))

	fun := s.handleDeleteMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not delete marker"}`)
}

func TestDeleteMarkerWithDBFail(t *testing.T) {
	s, mock := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("DELETE", "/marker/2/3", nil)

	req.Header.Set("Authorization", stubAuthHeader)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	mock.
		ExpectExec("DELETE").
		WillReturnError(errors.New("test error"))

	fun := s.handleDeleteMarker()
	fun(res, req)

	mock.ExpectationsWereMet()
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not delete marker"}`)
}

func TestDeleteNoAuthHeader(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("DELETE", "/marker/2/3", nil)
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleDeleteMarker()
	fun(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Could not find Authorization header"}`)
}

func TestDeleteInvalidAuthHeader(t *testing.T) {
	s, _ := getMockServer()
	defer s.finalize()

	req, err := http.NewRequest("DELETE", "/marker/2/3", nil)
	req.Header.Set("Authorization", "Bearer 1234")
	req.Header.Set("Content-Type", "application/json")

	assert.NoError(t, err)
	res := httptest.NewRecorder()

	fun := s.handleDeleteMarker()
	fun(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
	assert.Equal(t, res.Body.String(), `{"message":"Invalid Token"}`)
}
