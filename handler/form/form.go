package form

import (
	"net/http"
	"path/filepath"

	"github.com/czcorpus/mquery-sru/cnf"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/handler/common"

	"text/template"

	"github.com/gin-gonic/gin"
)

type FormHandler struct {
	serverInfo *cnf.ServerInfo
	conf       *corpus.CorporaSetup
	tmpl       *template.Template
}

func (a *FormHandler) Handle(ctx *gin.Context) {
	tplData := map[string]any{
		"Corpora":    a.conf.Resources.GetCorpora(),
		"ServerInfo": a.serverInfo,
	}
	if err := a.tmpl.ExecuteTemplate(ctx.Writer, "form.html", tplData); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
}

func NewFormHandler(
	serverInfo *cnf.ServerInfo,
	conf *corpus.CorporaSetup,
	projectRootDir string,
) *FormHandler {
	path := filepath.Join(projectRootDir, "handler", "form", "templates")
	tmpl := template.Must(
		template.New("").
			Funcs(common.GetTemplateFunctions()).
			ParseGlob(path + "/*"))
	return &FormHandler{
		serverInfo: serverInfo,
		conf:       conf,
		tmpl:       tmpl,
	}
}
