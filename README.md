# Snippetbox guided project

A web application where users can register and post snippets for everyone to see.

This is my introducty project for backend development. The code in this project is nearly all guided by the book "Let's Go" by Alex Edwards. The main difference is that I chose to use PostgreSQL instead of SQL, hence using a postgres driver in the code.

I learnt a lot about secure and maintainable backend development through writing this project. 

# Build instructions

Since this web app uses HTTPS we need TLS keys to run it. The TLS keys should be placed in the `./tls` directory. You can generate a self-signed TLS certificate using Go's crypto/tls standard library package by running the following command inside the `./tls` directory:
```bash
go run $GOPATH/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
```
A postgres database needs to be set up with the relevant tables. I have yet to automate this with SQL migrations. After setting up the tables and a role with permissions to the tables, keep a note of the DSN you want to the clients to use to connect to your database.

Then execute the following command to start the web app:
```bash
go run ./cmd/web -dsn=$YOUR_DSN
```
If everything worked as expected then you should see the console message "starting server on :4000". You should be able to visit the web app by going to https://localhost:4000/. If you are already using port 4000 for something else you can change the port with the `-addr` option.