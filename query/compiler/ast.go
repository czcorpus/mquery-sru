package compiler

type AST interface {
	Generate() string
	AddError(err error)
	Errors() []error
	TranslateWithinCtx(v string) string
}
