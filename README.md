Sunglasses
====

[![wercker status](https://app.wercker.com/status/e9b3149279373fa4cabb59ead53c3e48/m/master "wercker status")](https://app.wercker.com/project/bykey/e9b3149279373fa4cabb59ead53c3e48)

Sunglasses is a social platform with a great focus on privacy where you can get in touch with the people you want without having to worry about your data.

###Requirements
* Node.js and npm
* Golang 1.1 or higher
* MongoDB
* Redis

###Technologies
Sunglasses is built on top of:
* AngularJS (CoffeeScript)
* Martini (Golang)
* MongoDB
* Redis
* Gulp as a frontend build tool
* npm and bower as frontend package managers

###Install
In order to install a production instance of sunglasses you need nodejs (and npm), Golang, mongodb and redis up and running.

####Grab the source code
```bash
git clone https://github.com/mvader/sunglasses
cd sunglasses
```

####Edit config.sample.json
```bash
cp config.sample.json config.json
vim config.json
```
Now you have to edit all the values you need to setup sunglasses.
**Note:** copy config.sample.json, don't delete it or move it. It is used in the tests! 

####Install backend dependencies and make tests
```bash
go get -t github.com/smartystreets/goconvey
go get -t .
cd tests
go test -v
```
If everything is right we can proceed to build the backend.

####Build backend
```
cd ..
go build
```

####Install frontend dependencies
```bash
cd client
npm install
bower install
```

####Build frontend
```bash
gulp
```
Now you just have to run your app.

###Warning
This was a class project and thus it may be discontinued and the missing features not implemented. Use at your own risk.

