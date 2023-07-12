# GraphQL Persisted Queries Demo

This tech demo demonstrates how publicly exposed GraphQL APIs can be protected
from arbitrary queries using persisted queries (white/-allowlisting).
Persisted queries are generally easier to implement correctly compared to
[query complexity analysis](https://gqlgen.com/reference/complexity/)
(more about the topic here:
[howtographql.com - security](https://www.howtographql.com/advanced/4-security/)).

Queries are defined in `backend/persisted_queries`
as `<name>.graphql` files. This directory will be watched by the server
hot-reloading the persisted queries without downtime.
Еach query can then be executed via its name: `POST /e/<name>`.
If necessary, variables are to be provided as a JSON object
in the request body with the `Content-Type: encoding/json` header.

`GQL_PQ_MODE="ON_INIT"` disables hot-reloading of persisted queries.

with `MODE="DEBUG"` the server exposes direct querying via `/query` and the
GraphiQL playground via `/` as well as the persisted queries under `/e/`.
`MODE="PRODUCTION"` will only make the persisted query endpoints available
making sure clients can't execute arbitrary queries.

# Workflow 

Frontend developers add their queries to `backend/persisted_queries` to allow their
frontend application to call the API in production. Backend developers
are expected to review and control the persisted queries.

## FAQ

### Why names instead of hashes?
Usually, in the context of persisted queries, a hash of the query is sent
instead of a name. This demo uses human-readable names because it makes
monitoring and rate-limiting easier to configure.

### Why not just use REST/gRPC/etc.?
The difference between a GraphQL API with a query white-/allowlist
and a REST API or other RPC-like approaches to API design is that the former
is less tightly coupled to its consumers.
REST API endpoints usually need to be implemented and tested manually.
The maintenance of many such REST endpoints can become very expensive, thus
their number tends to be limited which can cause under- and overfetching.
A GraphQL API with query white-/allowlisting combines both the flexibility
and loose coupling of GraphQL with the predictability and safety of REST.

See [howtographql.com - GraphQL is the better REST](https://www.howtographql.com/basics/1-graphql-is-the-better-rest/)
for more information.