package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func DatabaseConnect() {
	// urlExample := "postgres://username:password@localhost:5432/database_name"

	databaseURL := "postgres://postgres:admin@localhost:5432/batch_42"

	var err error
	Conn, err = pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Succesfully connect to database.")
}