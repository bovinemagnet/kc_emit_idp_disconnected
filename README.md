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
WHERE age_in_sec > 60;
```

## To Execute

```bash
Usage of kc_emit_idp_disconnected:
-U, --username string   Database username. (default "keycloak")
-W, --password string   Database password. (default "keycloak")
-d, --dbname string     Database Name. (default "keycloak")
-h, --host string       Specifies the host name of the machine on which the server is running. (default "localhost")
-p, --port int          Database Port. (default 5432)
-0, --skipZero          Do not emit a message to the logger if the result count is zero.
-r, --realm string      Keycloak Realm (default "master")
-k, --logKey string     Log Key for the emitting to sys-logger. (default "auth_idp_disconnect_issue_count")
-s, --skipSec int       Skip seconds, ie if the user was created less than this many seconds ago, do not emit a message to the logger, as it is excluded from the results. (default 60)
-g, --logTag string     Log Tag for the emitting to sys-logger. (default "keycloak")
pflag: help requested
```