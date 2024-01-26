<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:sruResponse="http://docs.oasis-open.org/ns/search-ws/sruResponse"
    xmlns:hits="http://clarin.eu/fcs/dataview/hits"
    xmlns:fcs="http://clarin.eu/fcs/resource"
    xmlns:scan="http://docs.oasis-open.org/ns/search-ws/scan">

<xsl:template match="/scan:scanResponse">
    <html>
    <head>
        <meta charset="utf-8" />
        <title>MQuery-SRU - scan</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
    </head>
    <body>
        <h1>Scan</h1>
        <xsl:apply-templates />
    </body>
    </html>
</xsl:template>
</xsl:stylesheet>