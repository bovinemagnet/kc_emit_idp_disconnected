package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Parse command-line arguments
	var (
		username string
		password string
		realm    string
		dbName   string
		dbHost   string
		dbPort   string
		logKey   string
		emitZero bool
	)

	flag.StringVar(&username, "dbUsername", "keycloak", "Database username")
	flag.StringVar(&password, "dbPassword", "keycloak", "Database password")
	flag.StringVar(&dbName, "dbName", "keycloak", "Database Name")
	flag.StringVar(&dbHost, "dbHost", "localhost", "Database Name")
	flag.StringVar(&dbPort, "dbPort", "5432", "Database Port")
	flag.BoolVar(&emitZero, "emitZero", false, "If not set, do not emit zero values, ie if none found, do not emit a message to the logger.")

	flag.StringVar(&realm, "realm", "master", "Keycloak Realm")
	flag.StringVar(&logKey, "logKey", "auth_idp_disconnect_issue_count", "Log Key for the emitting to sys-logger")
	flag.Parse()

	//urlExample := "postgres://username:password@localhost:5432/database_name"
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		username,
		password,
		dbHost, dbPort, dbName)
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	sqlQuery := fmt.Sprintf("select count(*) from (select ue.id, ue.email, ue.first_name, ue.last_name, ue.realm_id, ue.username, ue.created_timestamp, ipm.federated_user_id, ipm.federated_username, ipm.user_id, (round(extract(epoch from current_timestamp) ) - created_timestamp /1000) as age_in_sec from user_entity ue left join federated_identity ipm on ( ue.id = ipm.user_id) where ipm.federated_username is null and ue.realm_id = '%s') AS ISSUES WHERE age_in_sec > 58;", realm)

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
	if (emitZero) || (count > 0) {
		syslogWriter, err := syslog.New(priority, "keycloak")
		if err != nil {
			log.Fatal(err)
		}
		message := fmt.Sprintf("auth_idp_disconnect_issue_count=%d", count)
		syslogWriter.Info(message)
	}

	fmt.Printf("[%s] auth_idp_disconnect_issue_count=%d\n", realm, count)
}
