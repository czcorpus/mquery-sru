# MQuery-SRU

MQuery-SRU is an easy to set up endpoint for [Clarin FCS](https://www.clarin.eu/content/federated-content-search-clarin-fcs-technical-details) 2.0 (Federated Content Search) based on the [Manatee-open](https://nlp.fi.muni.cz/trac/noske) corpus search engine and developed and maintained by the [Institute of the Czech National Corpus](https://ucnk.ff.cuni.cz/en/).

## Features

* Full support for the FCQ-QL query language
    * definable layer &#8596; Manatee-open positional attribute mapping
* Level 1 support for basic search via CQL (Context Query
Language)
* simultaneous search in multiple defined corpora
* (optional) backlinks to concordances in KonText


## Requirements

* a working Linux server with installed [Manatee-open](https://nlp.fi.muni.cz/trac/noske) library
* [Redis](https://redis.io/) database
* [Go](https://go.dev/)  language compiler and tools
* (optional) an HTTP proxy server (Nginx, Apache, ...)


## How to install

1. Install `Go` language environment, either via a package manager or manually from Go [download page](https://go.dev/dl/)
   1. make sure `/usr/local/go/bin` and `~/go/bin` are in your `$PATH` so you can run any installed Go tools without specifying a full path
2. Install Manatee-open from the [download page](https://nlp.fi.muni.cz/trac/noske). No specific language bindings are required.
   1. `configure --with-pcre && make && sudo make install && sudo ldconfig`
3. Get MQuery-SRU sources (`git clone --depth 1 github.com:czcorpus/mquery-sru.git`)
4. Run `./configure`
5. Run `make`
6. copy `mquery-sru` to a desired location and create a config file (conf.sample.json can be used as a starting point)
7. run:
   * main server: `mquery-sru server /path/to/conf.json` and
   * one or more workers: `WORKER_ID=0 mquery-sru worker /path/to/conf.json` (multiple workers can be run to utilize higher service load; in such case, set `WORKER_ID` properly for each one)
   * for OS integration, see <a href="#os-integration-systemd">OS integration (systemd)</a>

## Worker considerations

It's important to understand that endpoints experiencing low traffic can still benefit from having multiple workers. Specifically, if an endpoint is configured to search across multiple corpora, MQuery-SRU can leverage these workers to execute searches in parallel. This approach can significantly reduce the response time by querying all configured corpora simultaneously, thereby improving efficiency even under conditions of minimal load.

## Configuration

To run the endpoint, you need at least

1. to configure listening address and port
2. defined path to your Manatee corpora registry (= configuration) files
2. defined corpora along with:
    * positional attributes to be exposed
    * mapping of FCS-QL's `within` structures (`s`, `sentence`, `p` etc.) to your specific corpora structures
3. address and port of your Redis service plus a number of database to be used for passing queries and results around

See [configuration reference](https://github.com/czcorpus/mquery-sru/blob/main/config-reference.md) and/or [conf.sample.json](https://github.com/czcorpus/mquery-sru/blob/main/conf.sample.json) for detailed info.

## OS integration (systemd)

(Here we assume the service will run with user `www-data`)

Create a directory for logging (e.g. `/var/log/mquery-sru`) and set proper permissions for `www-data` to be able to write there.

You can use predefined systemd files from [/scripts/systemd/*](https://github.com/czcorpus/mquery-sru/tree/main/scripts/systemd). Copy (or link) them to `/etc/systemd/system` and then run:

```
systemctl enable mquery-sru-server.service
systemctl enable mquery-sru-worker-all.target
```

Now you can try to run the service:

```
systemctl start mquery-sru-server
systemctl start mquery-sru-worker-all.target
```

## See MQuery-SRU in action

A CNC instance of MQuery-SRU is running as one of the endpoints for Clarin [Content Search](https://contentsearch.clarin.eu/) page.
