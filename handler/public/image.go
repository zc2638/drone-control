/**
 * Created by zc on 2020/9/6.
 */
package public

import (
	"net/http"
)

func Image() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/image/avatar.jpg")
	}
}