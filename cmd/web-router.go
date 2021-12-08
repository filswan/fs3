// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/minio/minio/browser"
	"github.com/minio/minio/internal/logger"
	jsonrpc "github.com/minio/rpc"
	"github.com/minio/rpc/json2"
)

// webAPI container for Web API.
type webAPIHandlers struct {
	ObjectAPI func() ObjectLayer
	CacheAPI  func() CacheObjectLayer
}

// indexHandler - Handler to serve index.html
type indexHandler struct {
	handler http.Handler
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = minioReservedBucketPath + SlashSeparator
	h.handler.ServeHTTP(w, r)
}

const assetPrefix = "release"

// specialAssets are files which are unique files not embedded inside index_bundle.js.
const specialAssets = "index_bundle.*.js|loader.css|logo.svg|firefox.png|safari.png|chrome.png|favicon-16x16.png|favicon-32x32.png|favicon-96x96.png"

// registerWebRouter - registers web router for serving minio browser.
func registerWebRouter(router *mux.Router) error {
	// Initialize Web.
	web := &webAPIHandlers{
		ObjectAPI: newObjectLayerFn,
		CacheAPI:  newCachedObjectLayerFn,
	}

	// Initialize a new json2 codec.
	codec := json2.NewCodec()

	// MinIO browser router.
	webBrowserRouter := router.PathPrefix(minioReservedBucketPath).HeadersRegexp("User-Agent", ".*Mozilla.*").Subrouter()

	// Initialize json rpc handlers.
	webRPC := jsonrpc.NewServer()
	webRPC.RegisterCodec(codec, "application/json")
	webRPC.RegisterCodec(codec, "application/json; charset=UTF-8")
	webRPC.RegisterAfterFunc(func(ri *jsonrpc.RequestInfo) {
		if ri != nil {
			claims, _, _ := webRequestAuthenticate(ri.Request)
			bucketName, objectName := extractBucketObject(ri.Args)
			ri.Request = mux.SetURLVars(ri.Request, map[string]string{
				"bucket": bucketName,
				"object": objectName,
			})
			if globalTrace.NumSubscribers() > 0 {
				globalTrace.Publish(WebTrace(ri))
			}
			ctx := newContext(ri.Request, ri.ResponseWriter, ri.Method)
			logger.AuditLog(ctx, ri.ResponseWriter, ri.Request, claims.Map())
		}
	})

	// Register RPC handlers with server
	if err := webRPC.RegisterService(web, "web"); err != nil {
		return err
	}

	// RPC handler at URI - /minio/webrpc
	webBrowserRouter.Methods(http.MethodPost).Path("/webrpc").Handler(webRPC)
	webBrowserRouter.Methods(http.MethodPut).Path("/upload/{bucket}/{object:.+}").HandlerFunc(httpTraceHdrs(web.Upload))
	webBrowserRouter.Methods(http.MethodPost).Path("/deal/{bucket}/{object:.+}").HandlerFunc(httpTraceHdrs(web.SendDeal))
	webBrowserRouter.Methods(http.MethodPost).Path("/deals/{bucket}").HandlerFunc(httpTraceHdrs(web.SendDeals))
	webBrowserRouter.Methods(http.MethodGet).Path("/retrieve/{bucket}/{object}").HandlerFunc(httpTraceHdrs(web.RetrieveDeal))
	webBrowserRouter.Methods(http.MethodGet).Path("/bucket/retrieve/{bucket}").HandlerFunc(httpTraceHdrs(web.RetrieveDeals))
	webBrowserRouter.Methods(http.MethodPost).Path("/offlinedeal/{bucket}/{object:.+}").HandlerFunc(httpTraceHdrs(web.SendOfflineDeal))
	webBrowserRouter.Methods(http.MethodPost).Path("/offlinedeals/{bucket}").HandlerFunc(httpTraceHdrs(web.SendOfflineDeals))
	//webBrowserRouter.Methods(http.MethodPost).Path("/backup/add/plan").HandlerFunc(httpTraceHdrs(web.BackupVolumeAddPlan))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/add/plan").HandlerFunc(httpTraceHdrs(web.PsqlBackupVolumeAddPlan))
	//webBrowserRouter.Methods(http.MethodPost).Path("/backup/update/plan").HandlerFunc(httpTraceHdrs(web.BackupVolumeUpdatePlan))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/update/plan").HandlerFunc(httpTraceHdrs(web.PsqlBackupVolumeUpdatePlan))
	//webBrowserRouter.Methods(http.MethodGet).Path("/backup/retrieve/plan").HandlerFunc(httpTraceHdrs(web.RetrieveBackupPlan))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/retrieve/plan").HandlerFunc(httpTraceHdrs(web.PsqlRetrieveBackupPlan))
	//webBrowserRouter.Methods(http.MethodPost).Path("/backup/add/job").HandlerFunc(httpTraceHdrs(web.BackupAddJob))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/add/job").HandlerFunc(httpTraceHdrs(web.PsqlBackupAddJob))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/volume").HandlerFunc(httpTraceHdrs(web.SendOfflineDealsVolume))
	//webBrowserRouter.Methods(http.MethodGet).Path("/backup/retrieve/volume").HandlerFunc(httpTraceHdrs(web.RetrieveOfflineDealsVolume))
	webBrowserRouter.Methods(http.MethodPost).Path("/backup/retrieve/volume").HandlerFunc(httpTraceHdrs(web.PsqlRetrieveOfflineDealsVolume))
	//webBrowserRouter.Methods(http.MethodPost).Path("/rebuild/add/job").HandlerFunc(httpTraceHdrs(web.RebuildAddJob))
	webBrowserRouter.Methods(http.MethodPost).Path("/rebuild/add/job").HandlerFunc(httpTraceHdrs(web.PsqlRebuildAddJob))
	//webBrowserRouter.Methods(http.MethodPost).Path("/rebuild/volume").HandlerFunc(httpTraceHdrs(web.RebuildVolume))
	webBrowserRouter.Methods(http.MethodPost).Path("/rebuild/volume").HandlerFunc(httpTraceHdrs(web.PsqlRebuildVolume))
	//webBrowserRouter.Methods(http.MethodGet).Path("/rebuild/retrieve/volume").HandlerFunc(httpTraceHdrs(web.RetrieveRebuildVolume))
	webBrowserRouter.Methods(http.MethodPost).Path("/rebuild/retrieve/volume").HandlerFunc(httpTraceHdrs(web.PsqlRetrieveRebuildVolume))

	// These methods use short-expiry tokens in the URLs. These tokens may unintentionally
	// be logged, so a new one must be generated for each request.
	webBrowserRouter.Methods(http.MethodGet).Path("/download/{bucket}/{object:.+}").Queries("token", "{token:.*}").HandlerFunc(httpTraceHdrs(web.Download))
	webBrowserRouter.Methods(http.MethodPost).Path("/zip").Queries("token", "{token:.*}").HandlerFunc(httpTraceHdrs(web.DownloadZip))

	// Create compressed assets handler
	assetFS, err := fs.Sub(browser.GetStaticAssets(), assetPrefix)
	if err != nil {
		panic(err)
	}
	compressAssets := handlers.CompressHandler(http.StripPrefix(minioReservedBucketPath, http.FileServer(http.FS(assetFS))))

	// Serve javascript files and favicon from assets.
	webBrowserRouter.Path(fmt.Sprintf("/{assets:%s}", specialAssets)).Handler(compressAssets)

	// vue support
	webBrowserRouter.Path("/static/{assets:.*}").Handler(compressAssets)
	// Serve index.html from assets for rest of the requests.
	webBrowserRouter.Path("/{index:.*}").Handler(indexHandler{compressAssets})

	return nil
}
