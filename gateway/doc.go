/*
Package gateway manages the transfer of data between IP API and Postgres.

The package entry point is NewGatewayController. This returns a controller which
uses the DistLock distributed lock system to guarantee there is only one leader
gateway in the IB Connect cluster.

The leader will then load GatewayService values for each IB API URL. A
GatewayService is a disposable value that is only valid until it encounters an
error. If it encounters an error, it will advise the GatewayController. The
GatewayController will then terminate the failed instance and create a fresh
GatewayService for that IB API URL.

A GatewayService delegates its actual work to Feed values. A Feed deals with a
specific IB API use case. GatewayService uses FeedFactory implementations to
create a set of Feed values for its particular connection to IB. Each Feed uses
the Notifier system to receive refresh requests and send update advisories to
other interested parties (eg the REST server). A Feed can also abnormally
terminate at any time, in which case the owning GatewayService will go into
error and trigger a fresh restart by GatewayController.

The gateway package has been designed with visiblity rules that allow end users
to write their own Feed instances if required.
*/
package gateway
