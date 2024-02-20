# Configuration documentation

## Global settings

`listenAddress`: a network address the internal HTTP web server will listen to. It is recommended to use a local network and expose the service via an HTTP Proxy (Nginx, Apache) which allow
more fine-tuned configuration.

`listenPort`: a network port the internal HTTP web server will listen to. This is tightly related to `listenAddress`.

`serverReadTimeoutSecs` - ReadTimeout is the maximum duration in seconds for reading the entire
HTTP request, including the body. For an endpoint in Clarin FCU, this should be quite fast so there is
no need to set high values (like many tens of seconds).

`serverWriteTimeoutSecs` - the maximum duration in seconds an HTTP response can be written. Please note that in
case of a node in Clarin FCU, the response time should be ideally quite short so using values in many tens
of seconds provides no advantage here.

`sourcesRootDir` - specifies a local filesystem path where source codes of the project are located. We are mostly interested in `handler/(v12|v20)/templates`. (:construction:)
:exclamation: this value will be probably redefined in `v0.2`

`assetsURLPath` - specifies an external URL where assets (e.g. XSLT templates) can be found. This is not needed for basic endpoint functionality.

`logFile` (optional) - a file to write application log. If omitted, `stderr` is used.

`logLevel` (optional) - one of `debug`, `info`, `warning`, `error`. Defaults to `info`.

`timeZone` - local time zone. Defaults to `Europe/Prague`.

## SRU server info

`serverInfo.serverHost` - a public hostname of the endpoint (as required by SRU specification)

`serverInfo.serverPort` - a public port number of the endpoint (as required by SRU specification)

`serverInfo.database` - a resource database name
(defined in SRU specification)

`serverInfo.databaseTitle[lang]` - a human readable name for the endpoint database (defined in SRU specification)

`serverInfo.databaseDescription[lang]` - detailed information about the endpoint (defined in SRU specification)

## Corpora (resources)

`corpora.registryDir` - a local filesystem path where Manatee-open configuration (aka the "registry") files are located

`corpora.resources[i].id` - an ID of a defined corpus. By ID we mean its configuration/registry file name

`corpora.resources[i].pid` - a persistent ID of a defined corpus. This should be ideally an identifier registered with a respective authority

`corpora.resources[i].fullName[lang]` - a name of a defined corpus

`corpora.resources[i].description[lang]` - a detailed information about a defined corpus

`corpora.resources[i].viewContextStruct` - a structure used to specify KWIC range. In most cases, we need something like a sentence or a speach (so structures like `s`, `sp` etc.)

`corpora.resources[i].languages[]` - a list of languages (3-letter codes) a defined corpus contains

`corpora.resources[i].posAttrs[i].name` - name of a defined positional attribute (e.g. `word`, `lemma`,...)

`corpora.resources[i].posAttrs[i].id` - id of the attribute used within explain XML. This does not have to be a human readable value (e.g. `attr1`) - but it must be unique per corpus.

`corpora.resources[i].posAttrs[i].layer` - a text layer the attribute belongs to


`corpora.resources[i].posAttrs[i].isBasicSearchAttr` - specifies whether the attribute should be used for basic search. Multiple attributes can be set to true -
in such case the query looks like `[attr1="query" | attr2="query" | ... | attrN="query"]`

`corpora.resources[i].posAttrs[i].isLayerDefault` - tells whether the attribute should be used by default when searching using a layer it belongs to.

`corpora.resources[i].structureMapping[structType]` -
for different structure types (`utteranceStruct`,
`paragraphStruct`, `turnStruct`, `textStruct`, `sessionStruct`) defines actual structures matching those
general types (e.g. `"paragraphStruct": "p"`)

## Redis database

`redis.host` - an IP or hostname of available Redis instance

`redis.port` (optional) - a port used to connect to a Redis instance (defaults to 6379)

`redis.db` - Redis database number (1-16)

`redis.password` (optional) - a password used to connect to a Redis instance

`redis.channelQuery` (optional) - Redis channel to use for server-worker communication (defaults to `mquerysru`)

`redis.channelResultPrefix` (optional) - a prefix used for channels notifying about finished jobs in workers (defaults to `res`)

`redis.queryAnswerTimeoutSecs`(optional) - a time in seconds to wait for a worker to provide a result
(defaults to `30`)

