# kc_emit_idp_disconnected
Create a syslog message if Keycloak users in a database are not connected to their parent or proxy IdP.

Based on the following SQL, the code will connect to the postgresql database, run the query, and if there are any results, send a syslog message to the configured localhost syslog server.

```sql
select count(*)
from (
  select ue.id,
    ue.email,
    ue.first_name,
    ue.last_name,
    ue.realm_id,
    ue.username,
    ue.created_timestamp,
    ipm.federated_user_id,
    ipm.federated_username,
    ipm.user_id,
    (round(extract
      (epoch from current_timestamp)
        ) - created_timestamp /1000)
        as age_in_sec
    from user_entity ue
    left join federated_identity ipm
      on ( ue.id = ipm.user_id)
      where ipm.federated_username is null
      and ue.realm_id = 'realm'
) AS ISSUES
WHERE age_in_sec > 58;
```

## To Execute

```bash
Usage of kc_emit_idp_disconnected:
-dbHost string
Database Name (default "localhost")
-dbName string
Database Name (default "keycloak")
-dbPassword string
Database password (default "keycloak")
-dbPort string
Database Port (default "5432")
-dbUsername string
Database username (default "keycloak")
-emitZero
If not set, do not emit zero values, ie if none found, do not emit a message to the logger.
-logKey string
Log Key for the emitting to sys-logger (default "auth_idp_disconnect_issue_count")
-realm string
Keycloak Realm (default "master")
```