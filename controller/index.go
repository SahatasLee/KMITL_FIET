package controller

import "github.com/jmoiron/sqlx"

type DBController struct {
	Database *sqlx.DB
}
