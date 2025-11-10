# Development strategy 

## features

1. design the architecture and choose the technology stack;
2. define the API endpoints and data structures;
3. bootstrap the project with `POST /tags` without real implementation; rely on dummy adapters;
4. Set up the database schema and connections;
5. implement the db part for POST /tags; 
6. implement expvar metrics;
7. implment GET /tags without pagination;
8. implement pagination for GET /tags;
9. implement POST /media with db but without s3;
10. implement s3 adapter and integrate with POST /media;
11. develop tools for uploading media files;
12. implement POST /media/<id>/finalize 
13. implement GET /media/<id>
14. develop tools for downloading media files;

## tech choices

### Database driver: pgx vs lib/pq

**Decision: Use `github.com/jackc/pgx/v5` with `pgxpool`**

**Rationale:**
- **Performance**: 2-3x faster than lib/pq for most operations due to native PostgreSQL protocol implementation
- **Prepared statements**: Automatic prepared statement caching by default, reducing database overhead for repeated queries
- **Connection pooling**: Native `pgxpool` provides better connection management with configurable min/max connections, lifecycle, and health checks

## building 

1. locally with go build and go run;
2. Docker image;
3. github actions for CI;
4. docker compose with db and s3 simulator;

## testing

1. unit tests for handlers, and usecases;
2. unit tests with dockertest for db and s3 adapters;
3. manual testing with httpie and developed tools;
4. unit tests on CI;


