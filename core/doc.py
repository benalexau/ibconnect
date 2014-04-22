Package core provides types and values shared between IB Connect packages.

A given IB Connect cluster shares a common Postgres database. That Postgres
database is used for (i) data storage, (ii) distributed lock management and
(iii) distributed pub-sub messaging.

Each instance of IB Connect will load a GatewayController to manage the transfer
of data between IB API and the database, and a worker to make representations of
the database available via a HTTP REST server API.

*/
package core
