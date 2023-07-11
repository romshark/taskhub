# GraphQL Query Whitelisting Demo

This tech demo demonstrates how GraphQL APIs can be protected using
query whitelisting. Queries can be defined in `backend/whitelist`
as `<name>.graphql` files. This directory will be read by the server
at initialization time and the each query can then be executed via
its name: `POST /e/<name>`. If necessary, variables are to be provided
as JSON object in the request body with the `Content-Type: encoding/json`
content type header.

in `DEBUG` mode the server exposes direct querying via `/query` and the
GraphiQL playground via `/` as well as the whitelisted queries under `/e/`.
In `PRODUCTION` mode only the whitelisted query endpoints are available
to make sure clients can't execute arbitrary queries.

# Workflow 

Frontend developers add their queries to `backend/whitelist` to allow their
frontend application to call the API in production. Backend developers
are expected to review and control the whitelist.