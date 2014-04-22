Drone Config
============

You can setup free CI and code coverage tracking via [Drone](http://drone.io)
and [Coveralls](http://coveralls.io).

To integrate between Drone and Coveralls, note down the private "repo token"
assigned to your Coveralls project. Then adjust your Drone settings as shown
below.

Environment Variables
---------------------

```
DB_URL=postgres://postgres@localhost/ibc_dev?sslmode=disable
COVERALLS_TOKEN=the_repo_token
```

Command
-------

```
psql -c 'create database ibc_dev;' -U postgres
cd ibcd
go get
go build
cd ..
lsb_release -a
sudo start xvfb
sudo apt-get install -y telnet
cd testserver
./ibgwstart
sleep 15
cd ..
go get github.com/axw/gocov/gocov
go get github.com/mattn/goveralls
go get bitbucket.org/liamstask/goose/cmd/goose
./tests
./testserver/ibgwstop
goveralls -service drone.io -repotoken $COVERALLS_TOKEN -coverprofile all.cover
```
