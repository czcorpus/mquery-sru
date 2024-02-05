<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:sruResponse="http://docs.oasis-open.org/ns/search-ws/sruResponse"
    xmlns:hits="http://clarin.eu/fcs/dataview/hits"
    xmlns:fcs="http://clarin.eu/fcs/resource"
    xmlns:diag="http://docs.oasis-open.org/ns/search-ws/diagnostic">
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
                max-width: 60em;
                margin: 0 auto;
            }
            header {
                display: flex;
                align-items: center;
                background-color: #333333;
                color: #DEDEDE;
                padding: 1em;
                border-radius: 5px;
                margin-bottom: 1em;
            }
            header p {
                margin: 0;
            }
            header .summary {
                flex-grow: 1;
            }
            header h1 {
                margin: 0;
            }
            h1 {
                text-align: center;
                font-size: 20px;
                margin-top: 0;
            }
            div.rec {
                border: 1px solid #ababab;
                border-radius: 5px;
                margin-bottom: 0.5em;
            }
            h2.error {
                color: #dd1111;
            }
            h3.rec-pid {
            	   display: flex;
                background-color: #ababab;
                color: #eeeeee;
                margin: 0;
                padding-right: 1em;
                font-size: 11px;
                text-align: right;
            }
            h3.rec-pid span.record-idx {
                flex-grow: 1;
                text-align: left;
            }
            h3.rec-pid span.record-idx span.num {
                text-align: right;
                display: block;
                width: 3em;
            }
            .rec p {
                padding-left: 1em;
                padding-right: 1em;
            }
            .code {
                font-family: 'Courier New', Courier, monospace;
            }
            .hit {
                color: rgb(226, 0, 122);
            }
        </style>

    </head>
    <body>
        <header>
            <div class="summary">
                <p class="query">
                    <xsl:apply-templates select="/sruResponse:searchRetrieveResponse/sruResponse:echoedSearchRetrieveRequest" />
                </p>
                <p>
                    number of records: <xsl:value-of select="sruResponse:numberOfRecords" />
                </p>
            </div>
            <h1>MQuery-SRU</h1>
        </header>
        <xsl:apply-templates select="sruResponse:diagnostics" />
        <xsl:apply-templates select="sruResponse:records" />
    </body>
</html>
</xsl:template>

<xsl:template match="sruResponse:records">
   <xsl:apply-templates select="sruResponse:record" />
</xsl:template>

<xsl:template match="sruResponse:record">
    <div class="rec">
        <h3 class="rec-pid">
	<span  class="record-idx">
		<span class="num">
			<xsl:value-of select="./sruResponse:recordPosition" />
		</span>
	</span>
	<span>
	        <xsl:value-of select="./sruResponse:recordData/fcs:Resource/@pid" />
	</span>
        </h3>
        <xsl:apply-templates select="sruResponse:recordData/fcs:Resource" />
    </div>
</xsl:template>

<xsl:template match="fcs:Resource">
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
    <strong class="hit"><xsl:value-of select="." /></strong>
</xsl:template>

<xsl:template match="/sruResponse:searchRetrieveResponse/sruResponse:echoedSearchRetrieveRequest">
    query: <span class="code"><xsl:value-of select="sruResponse:query" /></span>
</xsl:template>

<!--
    error/diagnostics output
 -->

<xsl:template match="sruResponse:diagnostics">
    <h2 class="error">ERROR</h2>
    <h3>Detail:</h3>
    <p><xsl:value-of select="diag:diagnostic/diag:details" />
    </p>
    <h3>Message:</h3>
    <p>
    <xsl:value-of select="diag:diagnostic/diag:message" />
    </p>

</xsl:template>

</xsl:stylesheet>

