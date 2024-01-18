package form

import (
	"fcs/corpus"
	"net/http"

	"text/template"

	"github.com/gin-gonic/gin"
)

type FormHandler struct {
	conf *corpus.CorporaSetup
	tmpl *template.Template
}

func (a *FormHandler) Handle(ctx *gin.Context) {
	tplData := map[string]any{
		"Corpora": a.conf.Resources.GetCorpora(),
	}
	if err := a.tmpl.ExecuteTemplate(ctx.Writer, "form.html", tplData); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Writer.WriteHeader(http.StatusOK)
}

func NewFormHandler(conf *corpus.CorporaSetup) *FormHandler {
	tmpl := template.Must(
		template.New("").
			ParseGlob("handler/form/templates/*"))
	return &FormHandler{
		conf: conf,
		tmpl: tmpl,
	}
}
