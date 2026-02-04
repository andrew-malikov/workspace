package flr

type Context struct {
	Type    string
	Details any
}

type Failure struct {
	Context *Context
	Error   error
}

func Is[TDetails any](ctxType string, failure Failure) (*TDetails, bool) {
	if failure.Context.Type == ctxType {
		return failure.Context.Details.(*TDetails), true
	}
	return nil, false
}

func On(onError func(), onCtx func()) {
	
}

func OfError(error error) *Failure {
	return &Failure{Error: error}
}

func OfCtx(context Context) *Failure {
	return &Failure{Context: &context}
}
