package main

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

//Run SQL
func failSQL(sqlStr string) error {
	return sql.ErrNoRows
}

//DAO
func find(entity string) error {
	sqlStr := "select * from" + entity
	if err := failSQL(sqlStr); err != nil {
		return errors.Wrap(err, "DAO: failed SQL --"+sqlStr)
	}
	return nil
}

//Service
func getAllStudents() error {
	if err := find("Student"); err != nil {
		return errors.WithMessage(err, "service: failed to getAllStudents")
	}
	return nil
}

//Controller
func controller() {
	if err := getAllStudents(); err != nil {
		fmt.Printf("original error: %T %v\n", errors.Cause(err), errors.Cause(err))
		fmt.Printf("stack trace: \n%+v\n", err)
	}
}

func main() {
	controller()
}
