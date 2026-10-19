package reflect

import "net/http"

// cors lets a page on any origin read the reflection, and reports whether r is a preflight,
// which is answered without a body rather than reflected. Credentials are opt-in: the
// reflection carries the request's cookies, so any site could otherwise read them.
func cors(w http.ResponseWriter, r *http.Request, credentials bool) bool {
	h := w.Header()

	origin := r.Header.Get("Origin")
	if origin == "" {
		h.Set("Access-Control-Allow-Origin", "*")

		return false
	}

	h.Set("Access-Control-Allow-Origin", origin)
	h.Add("Vary", "Origin")
	if credentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	}

	method := r.Header.Get("Access-Control-Request-Method")
	if r.Method != http.MethodOptions || method == "" {
		return false
	}

	h.Set("Access-Control-Allow-Methods", method)
	if headers := r.Header.Get("Access-Control-Request-Headers"); headers != "" {
		h.Set("Access-Control-Allow-Headers", headers)
	}
	h.Set("Access-Control-Max-Age", "3600")

	return true
}
