package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/lukevers/golem"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

var (
	db gorm.DB
)

func initalizeDB() {
	// Lowercase our driver flag to make it easier to parse
	*driverFlag = strings.ToLower(*driverFlag)

	// Check if driver is mysql
	if *driverFlag == "mysql" {
		// Check if ?parseTime=true is not included
		if !strings.Contains(*databaseFlag, "?parseTime=true") {
			// If `?parseTime=true` was not included then we need to
			// add it so MySQL works properly.
			*databaseFlag += "?parseTime=true"
		}
	}

	// Check if driver flag is `sqlite`. It's a common typo to accidently
	// type `sqlite` instead of `sqlite3` (which we support), so to avoid
	// this completely, if we see `sqlite` anywhere in the driver string
	// we're just going to set it to `sqlite3` since that's the only
	// version of sqlite that we support.
	if strings.Contains(*driverFlag, "sqlite") {
		*driverFlag = "sqlite3"
	}

	// Check if driver flag contains `postgres`, and if it does then we just
	// want to change it to only be `postgres`. We're trying to avoid errors
	// in as many places as possible for the user.
	if strings.Contains(*driverFlag, "postgres") {
		*driverFlag = "postgres"
	}

	// Open connection
	db, err = gorm.Open(*driverFlag, *databaseFlag)
	if err != nil {
		golem.Warnf("Error connecting to database: %s", err)
		golem.Warn("Exiting with exit status 1")
		os.Exit(1)
	}

	// Test connection
	err = db.DB().Ping()
	if err != nil {
		golem.Warnf("Error pinging database: %s", err)
		golem.Warn("Exiting with exit status 1")
		os.Exit(1)
	}

	// Migrate database
	if *debugFlag {
		golem.Verb("Running database auto migrate")
	}

	db.AutoMigrate(User{})

	// Check to see if we have any users created.
	// If we don't have any users at all then we
	// need to make a default user.
	if *debugFlag {
		golem.Verb("Checking if any users exist")
	}

	db.FirstOrCreate(&User{
		Username:    "admin",
		Password:    HashPassword("admin"),
		Admin:       true,
		Twofa:       false,
		TwofaSecret: "",
	}, &User{})
}
