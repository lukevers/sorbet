# Sorbet

# Building

#### 0. Before You Build

Make sure you have [Go](http://golang.org/) installed. In order to compile the LESS/JS the preferred way is to use [Gulp](http://gulpjs.com/). To install Gulp you need to have [NPM](https://www.npmjs.org/) installed. Once you have NPM installed you can install Gulp via NPM:

```bash
npm install -g gulp
```

Once everything is installed make sure you have set your [$GOPATH](http://golang.org/doc/code.html#GOPATH) properly, or it will prove difficult to build.

#### 1. Get the Code

Start by cloning the repository and getting all the dependencies.

```bash
git clone https://github.com/lukevers/sorbet
cd sorbet
go get
```

#### 2. Build LESS/JS

Before we can run Gulp we need to make sure we install all of the necessary modules, and download our dependencies:
```bash
npm update
bower update
```

Building our webserver CSS/JS files is easy with Gulp.

```bash
gulp
```

When developing you can run `gulp watch` instead of running `gulp` every time you make changes.

If you'd rather use your own way of compiling LESS to CSS and concating all the CSS files into one file and JS files into one file, feel free. You can checkout `gulpfile.js` in the root of the directory to find out where these files are located and where they end up.

#### 3. Build the Source

```bash
go build
```

# Flags

### Debug

```bash
--debug
```
By including the debug flag, Sorbet will do the following:

* Recompile webserver templates on each page load
* Provide verbose stdout output

By default, Sorbet sets debug to `false`. This is a good option when developing Sorbet.


### Webserver Port

```bash
--port [port]
```

By including the webserver port flag you can change the port that Sorbet webserver listens on by default. By default Sorbet webserver listens on port `6015`.

### Webserver Interface

```bash
--interface [interface]
```

By including the webserver interface flag you can change the interface that Sorbet webserver binds to by default. By default the Sorbet webserver binds to the interface `127.0.0.1`.

### Database Driver

```bash
--driver [driver]
```

By including the database driver flag you can change the type of database that we are connecting to. By default Sorbet uses `sqlite3` as the default driver because Sorbet uses a SQLite3 database as default. If you change the database driver, you need to change the database connection details or it will not work.

Sorbet supports SQLite, MySQL, and PostgreSQL. To see how to use the database driver flag with the database flag, read the information on the database flag.

### Database

```bash
--database [connection]
```

By including the database flag you can change the connection details that we use. By default Sorbet uses `sorbet.db` as the database to connect to. If you change the database, you need to change the database driver or it will not work. 

Sorbet supports SQLite, MySQL, and PostgreSQL. The full list of options for database connection details can be found on each's website respectively. Here is an example for each:

##### SQLite

```bash
--driver sqlite3 --database /etc/sorbet.db
```

##### PostgreSQL

```bash
--driver postgres --database "user=username dbname=sorbet sslmode=disable"
```

##### MySQL

```bash
--driver mysql --database "username:password@tcp(host:port)/database"
```
