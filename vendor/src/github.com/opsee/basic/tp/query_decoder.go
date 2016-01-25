package tp

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"golang.org/x/net/context"
)

func QueryDecoder(key interface{}) DecodeFunc {
	return func(ctx context.Context, rw http.ResponseWriter, r *http.Request, p httprouter.Params) (context.Context, int, error) {
		newContext := context.WithValue(ctx, key, r.URL.Query())
		return newContext, 0, nil
	}
}
