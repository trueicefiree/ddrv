package webdav

import (
    "net/http"

    "github.com/spf13/afero"
    "golang.org/x/net/webdav"

    "github.com/forscht/ddrv/config"
)

func New(dfs afero.Fs) *http.Server {
    // Create a new WebDAV handler using the file system.
    handler := &webdav.Handler{
        FileSystem: NewFs(dfs),
        LockSystem: webdav.NewMemLS(),
    }

    // Set up Basic Authentication
    webdavHandler := webdavWithBasicAuth(handler, config.Username(), config.Password())

    // Create a new HTTP server with the provided address and the authentication handler.
    server := &http.Server{
        Addr:    config.WDAddr(),
        Handler: webdavHandler,
    }

    return server
}

// webdavWithBasicAuth is a middleware that wraps the provided handler with Basic Authentication.
func webdavWithBasicAuth(handler http.Handler, username, password string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, pass, ok := r.BasicAuth()
        if !ok || user != username || pass != password {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        handler.ServeHTTP(w, r)
    })
}
