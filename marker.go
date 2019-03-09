package main

import (
	"database/sql"
	"errors"
)

// MarkerCollection represents a collection of many markers all of the same user
type MarkerCollection struct {
	Markers []Marker `json:"markers"`
}

// Marker represents a marker in the trip pin points
type Marker struct {
	User string  `json:"user"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	Note string  `json:"note"`
}

func (m *Marker) save(db *sql.DB) error {
	sqlStatement := `
	INSERT INTO markers (username, lat, long, note)
	VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(sqlStatement, m.User, m.Lat, m.Lng, m.Note)

	return err
}

func getMarkerCollection(user string, db *sql.DB) (*MarkerCollection, error) {

	sqlStatement := `
	SELECT * FROM markers 
	WHERE username=$1
	`

	stmt, _ := db.Prepare(sqlStatement)
	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markerColelction MarkerCollection
	var note, dbUser string
	var lat, lng float64
	var id int

	for rows.Next() {
		err := rows.Scan(&id, &dbUser, &lat, &lng, &note)
		if err != nil {
			return nil, err
		}
		marker := Marker{Lat: lat, Lng: lng, Note: note, User: dbUser}
		markerColelction.Markers = append(markerColelction.Markers, marker)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &markerColelction, nil
}

func getMarker(user string, lat string, lng string, db *sql.DB) (*Marker, error) {

	sqlStatement := `
	SELECT * FROM markers 
	WHERE username=$1
	AND lat=$2
	AND lng=$3
	`

	stmt, _ := db.Prepare(sqlStatement)
	row := stmt.QueryRow(user, lat, lng)

	var note, dbUser string
	var resultLat, resultLng float64
	var id int

	err := row.Scan(&id, &dbUser, &resultLat, &resultLng, &note)

	if err != nil {
		return nil, err
	}

	return &Marker{Lat: resultLat, Lng: resultLng, Note: note, User: dbUser}, nil
}

func deleteMarker(user string, lat string, lng string, db *sql.DB) error {

	sqlStatement := `
	DELETE FROM markers
	WHERE username=$1
	AND lat=$2
	AND lng=$3
	`
	result, err := db.Exec(sqlStatement, user, lat, lng)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("Could not find marker to delete")
	}
	return nil
}
