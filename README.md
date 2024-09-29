# Multisite health endpoint
This is a simple go script that gives you the opportunity to monitor multiple site installations for a /health endpoint for monitoring with an automatic load balancer or just for fun. More of a proof of concept at this time, but should be production ready with a few minor changes and perhaps some better error handling.

## Usage
Add `[[endpoints]]` to config.toml file that you wish to monitor, using the example syntax with an `URL` string and a `HeaderHost` string. The `URL` is simply the URL you wish to monitor and the HeaderHost is what is sent in the `Host` header with the request, useful for example with WordPress multisite installations that expect that header to serve the correct site.

You can specify what HTTP response codes that are acceptable, for example 200 for an OK response, 401 if you expect that for sites that require prior auth and any other you see fit.

The health checker will then go through all listed endpoints and get the response code using a GET request. If all are withing the accepted status code range, it will return a `200 OK` response and list each endpoint and their response code.

If one or more supplied endpoints return an unaccepted response code, for example `500 Internal Server Error`, this script will return a `424 Failed Dependency` response, allowing you to monitor the `/health` endpoint for this in your monitoring setup.

## How to use it
1. Clone or download this repo to a suitable location.
2. Copy `config.toml.example` to `config.toml` and change the values to fit your needs.
3. Run with `go run .` or build an executable with `go build`. Change port if needed, 1337 is just a silly temporary port.
4. When running, you can go to `http://localhost:1337/health` for health checking the desired endpoints.

## Known issues
- For now, the port is hard coded into the main.go file. This should change into being setable within the `config.toml` file.
- Probably need better error handling.
- Gin needs to be configured for production environment.
