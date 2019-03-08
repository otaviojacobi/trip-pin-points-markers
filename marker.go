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

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return nil, err
	}

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

func getMarker(user string, markerID string, db *sql.DB) (*Marker, error) {

	sqlStatement := `
	SELECT * FROM markers 
	WHERE username=$1
	AND id=$2
	`

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(user, markerID)
	if err != nil {
		return nil, err
	}

	var note, dbUser string
	var lat, lng float64
	var id int

	err = row.Scan(&id, &dbUser, &lat, &lng, &note)

	if err != nil {
		return nil, err
	}

	return &Marker{Lat: lat, Lng: lng, Note: note, User: dbUser}, nil
}

func deleteMarker(user string, markerID string, db *sql.DB) error {

	sqlStatement := `
	DELETE FROM markers
	WHERE username=$1
	AND id=$2
	`
	result, err := db.Exec(sqlStatement, user, markerID)
	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return errors.New("Could not find marker to delete")
	}
	return err
}
