package main

import (
	"context"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	flag "github.com/spf13/pflag"
)

const (
	USERNAME = "keycloak"
	PASSWORD = "keycloak"
	PORT     = 5432
	SKIP_SEC = 60
)

func main() {

	// Get the path to the executable file
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[M]  Error:", err)
		return
	}

	// Get the name of the executable file
	exeName := filepath.Base(exePath)

	// Parse command-line arguments
	var (
		username string
		password string
		realm    string
		dbName   string
		dbHost   string
		dbPort   int
		logKey   string
		skipZero bool
		skipSec  int
		logTag   string
	)

	flag.NewFlagSet("Find Disconnected IDP Users in Keycloak", flag.ExitOnError)

	flag.StringVarP(&username, "username", "U", USERNAME, "Database username.")
	flag.StringVarP(&password, "password", "W", PASSWORD, "Database password.")
	flag.StringVarP(&dbName, "dbname", "d", "keycloak", "Database Name.")
	flag.StringVarP(&dbHost, "host", "h", "localhost", "Specifies the host name of the machine on which the server is running.")
	flag.IntVarP(&dbPort, "port", "p", PORT, "Database Port.")
	flag.BoolVarP(&skipZero, "skipZero", "0", false, "Do not emit a message to the logger if the result count is zero.")

	flag.StringVarP(&realm, "realm", "r", "master", "Keycloak Realm")
	flag.StringVarP(&logKey, "logKey", "k", "auth_idp_disconnect_issue_count", "Log Key for the emitting to sys-logger.")
	flag.IntVarP(&skipSec, "skipSec", "s", SKIP_SEC, "Skip seconds, ie if the user was created less than this many seconds ago, do not emit a message to the logger, as it is excluded from the results.")
	flag.StringVarP(&logTag, "logTag", "g", "keycloak", "Log Tag for the emitting to sys-logger.")
	flag.CommandLine.SortFlags = false
	flag.Parse()

	log.Printf("[START] %s", exeName)

	//urlExample := "postgres://username:password@localhost:5432/database_name"
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		username,
		password,
		dbHost, dbPort, dbName)
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Create the SQL Query to run against the keycloak database
	sqlQuery := fmt.Sprintf("select count(*) from (select ue.id, ue.email, ue.first_name, ue.last_name, ue.realm_id, ue.username, ue.created_timestamp, ipm.federated_user_id, ipm.federated_username, ipm.user_id, (round(extract(epoch from current_timestamp) ) - created_timestamp /1000) as age_in_sec from user_entity ue left join federated_identity ipm on ( ue.id = ipm.user_id) where ipm.federated_username is null and ue.realm_id = '%s') AS ISSUES WHERE age_in_sec > %d;", realm, skipSec)

	// Query the database for the number of rows in a table
	var count int
	err = conn.QueryRow(context.Background(), sqlQuery).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	// Send the result to syslog on the local host
	priority := syslog.LOG_INFO | syslog.LOG_LOCAL0

	// Change the default priority to warning if there are any issues
	if count > 0 {
		priority = syslog.LOG_WARNING | syslog.LOG_LOCAL0
	}

	// If we are emitting zero values, or if the count is greater than zero, emit the message
	if (count > 0) || (!skipZero) {
		syslogWriter, err := syslog.New(priority, logTag)
		if err != nil {
			log.Fatal(err)
		}
		message := fmt.Sprintf("%s=%d", logKey, count)
		syslogWriter.Info(message)
	}
	log.Printf("dbhost=%s  dbname=%s  realm=%s  %s=%d\n", dbHost, dbName, realm, logKey, count)
	log.Printf("[END] %s", exeName)
}
