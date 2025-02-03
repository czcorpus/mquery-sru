# MQuery-SRU

MQuery-SRU is an easy to set up endpoint for [Clarin FCS](https://www.clarin.eu/content/federated-content-search-clarin-fcs-technical-details) 2.0 (Federated Content Search) based on the [Manatee-open](https://nlp.fi.muni.cz/trac/noske) corpus search engine and developed and maintained by the [Institute of the Czech National Corpus](https://ucnk.ff.cuni.cz/en/).

## Features

* Full support for the [FCS-QL](https://clarin-eric.github.io/fcs-misc/fcs-core-2.0-specs/fcs-core-2.0.html#_fcs_ql_ebnf) query language
    * definable mapping between FCS-QL layers and Manatee-open positional attributes
* Level 1 support for basic search via CQL (Context Query
Language)
* simultaneous search in multiple defined corpora
* (optional) backlinks to respective concordances in KonText


## Requirements

* a working Linux server with installed [Manatee-open](https://nlp.fi.muni.cz/trac/noske) library
* [Redis](https://redis.io/) database
* [Go](https://go.dev/)  language compiler and tools
* (optional) an HTTP proxy server (Nginx, Apache, ...)


## How to install

1. Install `Go` language environment, either via a package manager or manually from Go [download page](https://go.dev/dl/)
   1. make sure `/usr/local/go/bin` and `~/go/bin` are in your `$PATH` so you can run any installed Go tools without specifying a full path
2. Install Manatee-open from the [download page](https://nlp.fi.muni.cz/trac/noske). No specific language bindings are required.
   1. `configure --with-pcre --disable-python && make && sudo make install && sudo ldconfig`
3. Get MQuery-SRU sources (`git clone --depth 1 https://github.com/czcorpus/mquery-sru.git`)
4. Run `./configure`
5. Run `make`
6. Run `make install`
      * the application will be installed in `/opt/mquery-sru`
      * for data and registry, `/var/opt/corpora/data` and `/var/opt/corpora/registry` directories will be created
      * systemd services `mquery-sru-server.service` and `mquery-sru-worker-all.target` will be created
8. Copy at least one corpus and its configuration (registry) into respective directories (`/var/opt/corpora/data`, `/var/opt/corpora/registry`)
9. Update corpora entries in `/opt/mquery-sru/conf.json` file to match your installed corpora
10. start the service:
      * `systemctl start mquery-sru-server`
      * `systemctl start mquery-sru-worker-all.target`

## HTTP access

In most cases, it is not recommended to expose the server directly to the Internet. It is therefore advisable to put the service behind an HTTP proxy.
E.g. in Nginx, the configuration may look like this:

```
location /mquery-fcs/ {
    proxy_pass http://127.0.0.1:8080/;
    proxy_set_header Host $http_host;
    proxy_redirect off;
    proxy_read_timeout 30;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;    
}
```

## Worker considerations

It's important to understand that endpoints experiencing low traffic can still benefit from having multiple workers. Specifically, if an endpoint is configured to search across multiple corpora, MQuery-SRU can leverage these workers to execute searches in parallel. This approach can significantly reduce the response time by querying all configured corpora simultaneously, thereby improving efficiency even under conditions of minimal load.

## Configuration

To run the endpoint, you need at least

1. to configure listening address and port
2. defined path to your Manatee corpora registry (= configuration) files
2. defined corpora along with:
    * positional attributes to be exposed and also layer names they belong to
    * mapping of FCS-QL's `within` structures (`s`, `sentence`, `p` etc.) to your specific corpora structures
3. address of your Redis service plus a number of database to be used for passing queries and results around

See [configuration reference](https://github.com/czcorpus/mquery-sru/blob/main/config-reference.md) and/or [conf.sample.json](https://github.com/czcorpus/mquery-sru/blob/main/conf.sample.json) for detailed info.

## OS integration (systemd)

This applies in case `make install` is not used.

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
