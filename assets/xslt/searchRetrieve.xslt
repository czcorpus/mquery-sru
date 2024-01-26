<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:sruResponse="http://docs.oasis-open.org/ns/search-ws/sruResponse"
    xmlns:hits="http://clarin.eu/fcs/dataview/hits"
    xmlns:fcs="http://clarin.eu/fcs/resource">

<xsl:template match="/sruResponse:searchRetrieveResponse">
<html>
    <head>
        <meta charset="utf-8" />
        <title>MQuery-SRU searchRetrieve result</title>
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <style>
            body {
                font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
                font-size: 16px;
                line-height: 1.5;
                color: #333;
                background-color: #fff;
            }
            h1 {
                text-align: center;
                font-size: 20px;
            }
            div.rec {
                border: 1px solid #ababab;
                border-radius: 5px;
                margin-bottom: 0.5em;
            }
            h3.rec-pid {
                background-color: #ababab;
                color: #eeeeee;
                margin: 0;
                padding-right: 1em;
                font-size: 11px;
                text-align: right;
            }
            .rec p {
                padding-left: 1em;
                padding-right: 1em;
            }
        </style>

    </head>
    <body>
        <h1>query result</h1>
        <xsl:apply-templates select="sruResponse:records" />
    </body>
</html>
</xsl:template>

<xsl:template match="sruResponse:record">
    <div class="rec">
        <xsl:apply-templates select="sruResponse:recordData" />
    </div>
</xsl:template>

<xsl:template match="fcs:Resource">
    <h3 class="rec-pid"><xsl:value-of select="@pid" /></h3>
    <xsl:apply-templates />
</xsl:template>

<xsl:template match="fcs:ResourceFragment/fcs:DataView">
    <xsl:apply-templates />
</xsl:template>

<xsl:template match="hits:Result">
    <p>
        <xsl:apply-templates />
    </p>
</xsl:template>


<xsl:template match="hits:Hit">
    <strong><xsl:value-of select="." /></strong>
</xsl:template>

</xsl:stylesheet>