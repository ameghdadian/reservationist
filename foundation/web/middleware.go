package web

type MidFunc func(HandlerFunc) HandlerFunc

func wrapMiddleware(mw []MidFunc, h HandlerFunc) HandlerFunc {
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			h = mwFunc(h)
		}
	}

	return h
}
