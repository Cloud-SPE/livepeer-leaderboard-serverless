# Leaderboard-Serverless

## Overview

This repository contains the Livepeer Leaderboard API server - A set of HTTP Rest Endpoints that provide statistics from job testing the Livepeer Orchestrator Network. 
see [figure 1](#figure-1---logical-overview-of-livepeers-testing).

### Key Features Supported
* APIs support POST (submit new test job statistics) and GET (query test job statistics).
  * Includes Raw Stats - query every job test results.
  * Includes Aggregated Stats - query the aggregated raw stats  
* Supports AI and Transcoding Test Job Types.
* Uses Postgres for data storage.
* Has a database migration tool to support different applications upgrades.
* Deploy using Docker or Binaries.

The serverless functions can be deployed using `vercel-cli` or self-hosted. 

### Production API Consumers 
* [Livepeer Inc](https://explorer.livepeer.org) (Performance Leaderboard)
* [Interptr](https://interptr-latest-test-streams.vercel.app) (Transcoding Test Stream UI)
* [Livepeer.Cloud SPE](https://inspector.livepeer.cloud/) (AI and Transcoding Test Job UI) 


## Required Software

This software can run on many operating systems.  Make sure you have the below software installed on you system before proceeding.

* **Git** 
  * [Install Guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* **Go** 1.23.1 or newer
  * [Install Go](https://go.dev/doc/install)
* **Postgres**
  * A postgres database is required. 
  * The database must be running prior to running the API server
  * see section *"Local Postgres local DB (optional)"*
* Vercel CLI
  * [Vercel CLI docs](https://vercel.com/docs/cli)
  * Optional, unless you deploy to Vercel
* Docker
  * Optional, unless you run a local Postgres DB

## Development

### Clone the repo

`git clone https://github.com/livepeer/leaderboard-serverless.git`
`cd <repo-folder>`

### Run Build

Run the go build:

`go build -o leaderboard-serverless main.go`

When the build completes, check for the binary file:
`leaderboard-serverless`

### Local Postgres local DB (optional)
The [`docker-compose.yml`](docker-compose.yml) file will allow you to spin up Postgres with a database 
called `leaderboad` and user named `leaderboard` and password `leaderboard` **DO NOT USE THESE VALUES FOR PRODUCTION!**

To run a local DB run the following docker command

`docker compose up -d`

This will enable the database to run on 

`<your ip>:5432`

When configuring the API Server, you will use the following environment variable:

`POSTGRES=postgres://leaderboard:leaderboard@<your ip>:5432/leaderboard?sslmode=disable;`

### Configure Environment Variables

*Note: Make sure to set the environment variables for your specific operating system*

#### Required 
* `POSTGRES=postgres://<YOUR_DATABASE_URL>` - The database conneciton URL. For example: postgres://leaderboard:leaderboard@localhost:5432/leaderboard
* `SECRET=<YOUR_SECRET KEY>`
  - The `SECRET` environment variable is part of the API security. The `/api/post_stats` HTTP POST Endpoint requires HMAC authentication.Any HTTP client posting data will have to use the same `SECRET` to create a HMAC message based on the `SECRET` that must be provided in the HTTP Authorization Header. The value of `SECRET` needs to be changed to your own private value. **DO NOT USE THIS VALUE FOR PRODUCTION!**

#### Optional
* `START_TIME_WINDOW` - The lookback period in hours for retrieving stats in aggregate or raw stats. Default is 24h.
* `DB_TIMEOUT` - The time in seconds used for database operations before they will timeout. Default is 20s.
* `LOG_LEVEL`  - The logging level of the application. Default is INFO.
* `SECRET` - The secret used in HTTP Authorization headers to authenitcate callers of protected endpoints.  See the section on Endpoint Security.  This is optional is you do not intend to post stats.
* `REGIONS_CACHE_TIMEOUT` - The timeout for the application to cache regions before retrieving them from the database.  The default is 60 seconds.
* `PIPELINES_CACHE_TIMEOUT` - The timeout for the application to cache pipelines before retrieving them from the database.  The default is 60 seconds.
* `CATALYST_REGION_URL` - A custom URL point to the Catlyst JSON representing regions to be inserted into the database.

#### ClickHouse (required for platform metrics endpoints)

The `/api/gpu/metrics`, `/api/network/demand`, and `/api/sla/compliance` endpoints are backed by ClickHouse. These environment variables must be set for those endpoints to return data.

* `CLICKHOUSE_HOST` - ClickHouse server hostname. Default is `localhost`.
* `CLICKHOUSE_PORT` - ClickHouse server port. Default is `8123`.
* `CLICKHOUSE_DB` - ClickHouse database name. Default is `livepeer_analytics`.
* `CLICKHOUSE_USER` - ClickHouse username. Default is `analytics_user`.
* `CLICKHOUSE_PASS` - ClickHouse password. Default is `analytics_password`.
* `CLICKHOUSE_PROTOCOL` - ClickHouse connection protocol (`http` or `native`). Default is `http`.

### Run the App

Next, execute the `leaderboard-serverless` binary

Your console will show the following output:
```
2024/09/06 17:09:50 Setting up log level to default: INFO
time=2024-09-06T17:09:50.771Z level=INFO msg="Server starting on port 8080"
```

In your browser, go to `http://<YOUR_SERVER_IP>:8080/api/regions` to verify the server works

You will see similar JSON output:
```
{
  "regions": [
    {
      "id": "TOR",
      "name": "Toronto",
      "type": "transcoding"
    },
    {
      "id": "HND",
      "name": "Tokyo",
      "type": "transcoding"
    },
    ...
  ]
}
```

### Troubleshooting Tip

If your `POSTGRES` environment variable is misconfigured, you may see an error like this:

```
failed to connect to `host=localhost user=leaderboard database=leaderboard`: dial error (dial tcp [::1]:5432: connect: connection refused)
```

Make sure you can connect to the database url defined with proper user and password credentials.

### Running The Unit Tests

Tests can be run with the following command in the project root:

```
go build && go test -p 1-v ./...
```

Since we use an emebedded database for the entire test run between packages, test packages must be run one at a time (-p 1 flag). 

See [embedded-postgres issue #115](https://github.com/fergusstrange/embedded-postgres/issues/115) for more details.

## Production

Livepeer Inc hosts a version of this API to support the Livepeer Explorer Performance Leaderboard.

### Livepeer Inc's API
- Production API: [leaderboard-serverless.vercel.app/api/](https://leaderboard-serverless.vercel.app/api/)
- Staging API: [staging-leaderboard-serverless.vercel.app/api/](https://staging-leaderboard-serverless.vercel.app/api/)

## API Reference

All API endpoint documentation has moved to [API Reference](docs/api-reference.md).

## Database

The database is responsible for storing the results of test data for each job executed as well as some reference data (regions).

The new schema allows for easy addition for Regions and Job Types (AI, Transcoding, etc...)

The new schema has index all queries to allow fast searching on common fields:
orchestrator
timestamp
model/pipeline
region
job type 

![Leaderboard Database Entity Relation Diagram](docs/new-db-entity-relation.png)

## Migrations

As reference data and schema design evolves, it is necessary to deploy these changes to your backend database.  In order to avoid human error and manual tasks, database migrations are automated in this project.  This means one can update DDL and DML in the databsae with the addition of a SQL script.  In other words, you can alter the structure of the database or the data stored in the databse with these migrations.

These scripts must follow the naming convention of <migration_number>_descriptive_text>.<up | down>.sql.  As an example:

```
8_create_users_table.up.sql
8_create_users_table.down.sql
```
This example defines the eighth migration for the datbaase with two migrations, one to upgrade the database and one to revert it.  When the application starts its connection to the database, it will run all upgrade (or up) migrations automatically.

Database migrations are found in [`assets/migrations`](assets/migrations). These are embeded in the golang binary built from this project for ease of access regardless of where the application is deployed.  This also allows Vercel to use these migrations.  The migrations are loaded and processed by the [golang-migrate](https://github.com/golang-migrate/migrate) project.  Please read their documentation for more details.

### Upgrading the database

The database upgrade runs automatically once the first API is called.  This is because the first API call instantiates a connection to the database, which first checks that the database exists and then applies any migrations necessary to the database.

### Downgrading the database

IMPORTANT: This is not a substitute for backing up your database.  Before attempting any of the below steps, ensure you back up your database.

Run the following command using docker to downgrade the database to it's original state (-all) or to a specific version (N).  The example below shows how to downgrade by one version (N=1).

> docker run --rm -it -v ./assets/migrations:/migrations --network host migrate/migrate -path=/migrations/ -verbose -database $POSTGRES down 1

The resulting output:

```2024/09/04 17:00:25 Start buffering 2/d migrate_data_to_events
2024/09/04 17:00:25 Start buffering 1/d event_design_refactor
2024/09/04 17:00:25 Read and execute 2/d migrate_data_to_events
2024/09/04 17:00:25 Finished 2/d migrate_data_to_events (read 8.673955ms, ran 8.698581ms)
2024/09/04 17:00:25 Read and execute 1/d event_design_refactor
2024/09/04 17:00:25 Finished 1/d event_design_refactor (read 26.303185ms, ran 27.874975ms)
2024/09/04 17:00:25 Finished after 55.855055ms
2024/09/04 17:00:25 Closing source and database
```

## Old DB Schema Reference

Each Region was a separate table. The only tests that existed was the Transcoding work. As part of the [Livepeer.Cloud SPE Proposal #2](https://explorer.livepeer.org/treasury/69112973991711207069799657820129915730234258793790128205157315299386501373337), the database was migrated to a table called "EVENTS".

![Leaderboard Database Entity Relation Diagram](docs/old-db-entity-relation.png)

## Appendix

### Figure 1 - Logical Overview of Livepeer's Testing

This repository documents the *"Livepeer API Server"* and *"Database Server"* boxes defined in figure 1.

![Job Tester Logical Architecture](docs/job-tester-logical_architecture.png)
