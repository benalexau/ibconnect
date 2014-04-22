[![Build Status](https://drone.io/github.com/benalexau/ibconnect/status.png)](https://drone.io/github.com/benalexau/ibconnect/latest)
[![GoDoc](https://godoc.org/github.com/benalexau/ibconnect?status.png)](https://godoc.org/github.com/benalexau/ibconnect)
[![Coverage Status](https://coveralls.io/repos/benalexau/ibconnect/badge.png?branch=master)](https://coveralls.io/r/benalexau/ibconnect?branch=master)

IB Connect is a HTTP service that provides a durable, long-term history of
[Interactive Brokers](https://www.interactivebrokers.com) account performance.

While IB Account Management offers extensive reporting features (eg Portfolio
Analyst), users need to login to their account (which in many cases involves a
hardware security token) and run reports that provide EOD granularity at most.
This quickly becomes tedious, particularly if multiple IB accounts are managed.

IB Connect addresses this issue by using IB API to automatically monitor account
balances at any interval desired. It's intended for financial advisor accounts,
as it will record the status of every detected account. A single IB Connect
instance can also handle multiple IB Gateways at once, allowing a wide variety
of deployment scenarios (eg monitoring your own firm's accounts, monitoring
paper trade accounts, monitoring accounts being advised by another firm etc).
All the data ends up in the same convenient database and REST endpoints. Of
course, you can use IB Connect with a standalone account just fine.

While IB Account Management reports should still be used for compliance and
taxation purposes (eg official balances etc), IB Connect offers a practical
solution for automatically monitoring many IB accounts from a single location.

Features:

* 100% pure Go
* JSON-based REST API is easily consumed by almost any programming language
* Observes [Twelve-Factor Methodology](http://12factor.net/) for easier devops
* Reflects [best practices for pragmatic RESTful APIs](http://www.vinaysahni.com/best-practices-for-a-pragmatic-restful-api)
* Uses Postgres for reliable, consistent data durability
* Normalised schema for space efficient, constraint-verifying data management
* Records snapshot information (eg account balances) automatically and on request
* Recovers from many errors by automatic restarts (eg IB Gateway I/O timeouts)
* Works with financial advisor accounts, managed accounts and standalone accounts
* Designed for financial advisor requirements (eg IB advisor logins, account churn etc)
* Supports multiple IB Gateway instances from a single IB Connect process
* Compatible with multiple IB Connect instances for fault tolerance and scale-out
* Sends cache headers to help clients avoid re-requesting unchanging content
* Clean architecture extensible to additional reporting needs (eg executions, orders, positions etc)

IB Connect builds on the [GoIB](https://github.com/gofinance/ib) library. Go
developers should use GoIB if they wish to perform trading (IB Connect does not
provide trading features; its scope is limited to account information only).

To install, ``go get github.com/benalexau/ibconnect/ibcd``.

Environment Variables
---------------------
Pursuant to the [Twelve-Factor Methodology](http://12factor.net/), all
configuration is performed via environment variables. The following environment
variables are currently supported:

| Variable     | Default                | Comment                              |
| ------------ | ---------------------- | ------------------------------------ |
| ``DB_URL``   | ``postgres://ibc_dev@localhost/ibc_dev?sslmode=disable``|Postgres only|
| ``IB_GW``    | ``127.0.0.1:4002``     | Separate multiple values with commas |
| ``IB_CID``   | ``5555``               | API Client ID (unique to IB Connect) |
| ``ERR_INFO`` | ``false``              | Extra details in HTTP status code 500|
| ``PORT``     | ``3000``               | HTTP listener port number            |
| ``HOST``     | ``localhost``          | HTTP listener IP to bind             |
| ``ACCT_REF`` | ``@hourly``            | Account snapshot cron interval (UTC) |

REST Endpoints
--------------

A HTTP GET of ``http://yourserver:3000/v1/accounts`` will return a JSON list of
all accounts ever found in any IB Gateway instance. You can include a
``Cache-Control`` header of ``max-age=0`` to force a refresh of the IB Gateway
backend.

You can HTTP GET ``http://yourserver:3000/v1/accounts/ACCTNO`` to receive a
HTTP status 303 redirect to the latest report URL for that account number.

Finally, all historical reports are available under the HTTP GET URL format
``http://yourserver:3000/v1/accounts/ACCTNO/RFC3339NANO``. For example,
``http://yourserver:3000/v1/accounts/U12345678/2014-04-22T04:22:05.776394Z``.
The returned JSON contains sections for the account balances and portfolio. 
The account balances are formatted without currency minor unit separators (in
other words, dollar amounts are displayed in cents).

Design Overview
---------------

| Directory           | Description                                               |
| ------------------- | --------------------------------------------------------- |
| [db](db/)           | SQL scripts for ``goose`` database migrations (see below) |
| [core](core/)       | Package ``core`` contains types and values used elsewhere |
| [gateway](gateway/) | Package ``gateway`` transfers between Postgres and IB API |
| [ibcd](ibcd/)       | Package ``main`` contains the IB Connect daemon           |
| [server](server/)   | Package ``server`` offers a REST API for Postgres data    |

In general, loading ``ibcd`` will cause the gateway system to load if it isn't
already running in the cluster. The gateway will refresh its data on request
from the REST server tier (usually by a REST request with ``max-age=0``) or
automatically if (i) the gateway just started, (ii) a time indicated by one of
the environment-variable cron expressions is reached, or (iii) IB Gateway
business logic indicates it's appropriate to do so.

Send a SIGTERM to the ``idbc`` process to exit (or just press C-c).

Prerequisites
-------------
As indicated by the environment variables, you need to run IB Gateway and a
Postgres database (with a dedicated user and database).

We suggest creating an ``ibc_dev`` user and database on the development
machine in order to match the default environment variable noted above.

To install Postgres on Arch Linux:

```
sudo pacman -S postgres
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

To create a Postgres user and database:

```
sudo su postgres
cd ~
psql -c "CREATE USER ibc_dev;"
psql -c "CREATE DATABASE ibc_dev OWNER ibc_dev;"
psql ibc_dev -c "ALTER SCHEMA public OWNER TO ibc_dev;"
exit
```

IB Connect manages schemas via [Goose](https://bitbucket.org/liamstask/goose).
Note the Goose [db/dbconf.yml](db/dbconf.yml) declares a single ``db`` environment
that expects the ``DB_URL`` to have been set. There shouldn't be any need to
edit the ``dbconf.yml`` between development and production, which is a goal of
the [Twelve-Factor Methodology](http://12factor.net/). Use these commands to
install Goose and configure your schema:

```
go get bitbucket.org/liamstask/goose/cmd/goose
DB_URL=postgres://ibc_dev@localhost/ibc_dev?sslmode=disable goose -env db up
DB_URL=postgres://ibc_dev@localhost/ibc_dev?sslmode=disable goose -env db status
```

Production
----------
If you're running IB Connect in production, you may wish to consider:

1. [IBController](http://sourceforge.net/projects/ibcontroller/) can manage your
   IB Gateway instance(s).
2. Postgres has numerous high availability capabilities. The best choice will
   depend on your particular environment's configuration and reliability needs.

Tests
-----
Tests require the ``IB_GW`` endpoint to be running a financial advisor account.
The test server directory provides a suitable IB Gateway testing endpoint (see
the [test server instructions](testserver/README.md) for details).

The following shell command uses ``goose`` to bring the database back to an
empty state, apply all migrations, run the tests and report test coverage
on the console and in a web browser for convenient visual inspection:

``DB_URL=postgres://ibc_dev@localhost/ibc_dev?sslmode=disable ./tests``

Style Guide
-----------
Go code is automatically formatted by ```goimports```.

License
-------
[GNU General Public License](http://www.gnu.org/licenses/gpl.html) version 3
applies to IB Connect.

The goal of the above license is to facilitate improvements in IB Connect being
shared with the wider community. It makes no claim over the license or
distribution of your own separate applications which simply use IB Connect via
its provided network interface.
