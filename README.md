[![Build Status](https://travis-ci.com/otaviojacobi/trip-pin-points-markers.png)](https://travis-ci.com/otaviojacobi/trip-pin-points-markers)

# Trip Pin Points Markers Service

## AVAILABLE AT: http://trip-pin-points-markers.sa-east-1.elasticbeanstalk.com

## API Documentation: https://app.swaggerhub.com/apis-docs/otaviojacobi/trip-pin-points-markers/1.0.2

## Running

 - To run this application have [Go](https://golang.org/doc/install) set up.
 - [Start the postgresql database](https://www.postgresql.org/docs/9.1/server-start.html)
 - Set a the following environment variables pointing to your postgresql db
    ```
    RDS_USERNAME=postgres_username_here
    RDS_PASSWORD=postgres_password_here
    RDS_HOSTNAME=localhost
    RDS_PORT=5432
    RDS_DB_NAME=postges_db_name_here
    ```
 - `go run .` will start the service


## Running unit tests and reports
 - To run the tests run `go test . ./...`
 - To run the tests and see coverage run `go test -coverprofile=c.out . ./... && go tool cover -html=c.out`
