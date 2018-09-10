# Feature Flag Service

`ffs` is a simple implementation of a feature flag service.

A feature flag service enables software to make decisions at runtime about how
to operate. This can be used as a way to deploy "dark features" and later
activate them for some or all traffic. You might want to do this to decouple
deployment and release of new code, to perform canary releases or as a way
to do A/B testing.

# Parts

This repo includes the ffs:

* server, holds the authoritative view of feature flags and serves flags
* client library, which synchronises application flag state with server
* cli tool, for performing flag admin tasks on the server
* demo service, which is our service we want to deploy with dark features

# How to use

Run the feature flag server.

    go run github.com/aelse/ffs/server

Run the demo service.

    go run github.com/aelse/ffs/demo

You can now access the demo at [http://localhost:8081](http://localhost:8081)
in your web browser.

Using curl you can modify the feature flags by POSTing to the feature flag
server listening on port 8080.

    curl -i -X POST http://localhost:8080/flags -H "Content-Type: application/json" -d '{ "name": "myflag", "value": true }'

Reload the browser page to see the new flag value.

# How it works

The feature flag server keeps a representation of all feature flags in memory.

The client also keeps a representation in memory.

Clients connect over websocket and receive the list of all current flag names
and values. When you POST to the `/flags` endpoint the added or modified flag
is sent to all connected clients. In this way clients always have an up to date
view of all feature flags.

In the application code making use of feature flags, the client provides an
api for the application code to check the value it should use for a flag.
The application asks the client "what do I do for flag X?" and then acts
accordingly.