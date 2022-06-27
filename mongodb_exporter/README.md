# Mongodb Exporter

An exporter for [Prometheus](https://prometheus.io) that collects logs from [Mongo Atlas](https://www.mongodb.com/atlas/database) database nodes and provides analytic metrics based on [mtools](https://github.com/rueckstiess/mtools).

## Install

Build the executable to `./exporter`

```bash
make build
```

Or from Dockerfile

```bash
docker build -t shopline/mongodb-exporter .
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| CRAWLER_INTERVAL_TIME | The interval of crawling on mongodb logs | `"1h"` |
| CRAWLER_SINCE_TIME | The time range of logs to be retrieved | `"1h"` |
| MONGO_ATLAS_GROUP_ID | Project ID from altas path, e.g. https://cloud.mongodb.com/v2/GROUP_ID_HERE#clusters | `""` |
| MONGO_ATLAS_CLUSTER_NAME | Just a label for record | `""` |
| MONGO_ATLAS_CLUSTER_IDS | Domain of nodes to be crawled separated by comma, e.g. XXX-shard-00-00.XXX.mongodb.net,XXX-shard-00-01.XXX.mongodb.net | `""` |
| MONGO_ATLAS_PUBLIC_KEY | Public key of [API Key](https://cloud.mongodb.com/v2#/org/GROUP_ID_HERE/access/apiKeys) | `""` |
| MONGO_ATLAS_PRIVATE_KEY | Private key of [API Key](https://cloud.mongodb.com/v2#/org/GROUP_ID_HERE/access/apiKeys) | `""` |

## Development

Make sure you have Docker installed, run `make bash` to kickstart a docker container and uses your host network (as the API key might be IP-restricted) and required to running `mtools` with non-TTY context.

It will mount the current folder in and running Go version 1.18.3 (min required version >= 1.17).

```bash
% make bash 
... docker build ...
root@3d275f4bc235:/go/src#
```

## Limitation

- mtools does not work with Mongodb version 4.4+ (https://github.com/rueckstiess/mtools/issues/806) for now but might be workarounded.
