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
	"bytes"
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filedrive-team/go-graphsplit"
	"github.com/google/uuid"
	csv "github.com/minio/csvparser"
	"github.com/minio/minio/internal/config"
	"github.com/minio/minio/logs"
	oshomedir "github.com/mitchellh/go-homedir"
	"github.com/syndtr/goleveldb/leveldb"
	"mime/multipart"

	//"github.com/filedrive-team/go-graphsplit"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/klauspost/compress/zip"
	"github.com/minio/minio-go/v7"
	miniogo "github.com/minio/minio-go/v7"
	miniogopolicy "github.com/minio/minio-go/v7/pkg/policy"
	"github.com/minio/minio-go/v7/pkg/s3utils"

	"github.com/minio/minio/internal/auth"
	objectlock "github.com/minio/minio/internal/bucket/object/lock"
	"github.com/minio/minio/internal/bucket/replication"
	"github.com/minio/minio/internal/config/dns"
	"github.com/minio/minio/internal/config/identity/openid"
	"github.com/minio/minio/internal/crypto"
	"github.com/minio/minio/internal/etag"
	"github.com/minio/minio/internal/event"
	"github.com/minio/minio/internal/handlers"
	"github.com/minio/minio/internal/hash"
	xhttp "github.com/minio/minio/internal/http"
	"github.com/minio/minio/internal/ioutil"
	"github.com/minio/minio/internal/logger"
	"github.com/minio/pkg/bucket/policy"
	iampolicy "github.com/minio/pkg/iam/policy"
	"github.com/minio/rpc/json2"

	ioioutil "io/ioutil"
	"os/exec"
)

const (
	SuccessResponseStatus = "success"
	FailResponseStatus    = "fail"
	NoFileInBucket        = "No file in the bucket.Please upload files"
)

func extractBucketObject(args reflect.Value) (bucketName, objectName string) {
	switch args.Kind() {
	case reflect.Ptr:
		a := args.Elem()
		for i := 0; i < a.NumField(); i++ {
			switch a.Type().Field(i).Name {
			case "BucketName":
				bucketName = a.Field(i).String()
			case "Prefix":
				objectName = a.Field(i).String()
			case "ObjectName":
				objectName = a.Field(i).String()
			}
		}
	}
	return bucketName, objectName
}

// WebGenericArgs - empty struct for calls that don't accept arguments
// for ex. ServerInfo
type WebGenericArgs struct{}

// WebGenericRep - reply structure for calls for which reply is success/failure
// for ex. RemoveObject MakeBucket
type WebGenericRep struct {
	UIVersion string `json:"uiVersion"`
}

// ServerInfoRep - server info reply.
type ServerInfoRep struct {
	MinioVersion    string
	MinioMemory     string
	MinioPlatform   string
	MinioRuntime    string
	MinioGlobalInfo map[string]interface{}
	MinioUserInfo   map[string]interface{}
	UIVersion       string `json:"uiVersion"`
}

// ServerInfo - get server info.
func (web *webAPIHandlers) ServerInfo(r *http.Request, args *WebGenericArgs, reply *ServerInfoRep) error {
	ctx := newWebContext(r, args, "WebServerInfo")
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}
	host, err := os.Hostname()
	if err != nil {
		host = ""
	}
	platform := fmt.Sprintf("Host: %s | OS: %s | Arch: %s",
		host,
		runtime.GOOS,
		runtime.GOARCH)
	goruntime := fmt.Sprintf("Version: %s | CPUs: %d", runtime.Version(), runtime.NumCPU())

	reply.MinioVersion = Version
	reply.MinioGlobalInfo = getGlobalInfo()

	// Check if the user is IAM user.
	reply.MinioUserInfo = map[string]interface{}{
		"isIAMUser": !owner,
	}

	if !owner {
		creds, ok := globalIAMSys.GetUser(claims.AccessKey)
		if ok && creds.SessionToken != "" {
			reply.MinioUserInfo["isTempUser"] = true
		}
	}

	reply.MinioPlatform = platform
	reply.MinioRuntime = goruntime
	reply.UIVersion = Version
	return nil
}

// StorageInfoRep - contains storage usage statistics.
type StorageInfoRep struct {
	Used      uint64 `json:"used"`
	UIVersion string `json:"uiVersion"`
}

// StorageInfo - web call to gather storage usage statistics.
func (web *webAPIHandlers) StorageInfo(r *http.Request, args *WebGenericArgs, reply *StorageInfoRep) error {
	ctx := newWebContext(r, args, "WebStorageInfo")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}
	_, _, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}
	dataUsageInfo, _ := loadDataUsageFromBackend(ctx, objectAPI)
	reply.Used = dataUsageInfo.ObjectsTotalSize
	reply.UIVersion = Version
	return nil
}

// MakeBucketArgs - make bucket args.
type MakeBucketArgs struct {
	BucketName string `json:"bucketName"`
}

// MakeBucket - creates a new bucket.
func (web *webAPIHandlers) MakeBucket(r *http.Request, args *MakeBucketArgs, reply *WebGenericRep) error {
	ctx := newWebContext(r, args, "WebMakeBucket")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// For authenticated users apply IAM policy.
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.CreateBucketAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errAccessDenied)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, true) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	opts := BucketOptions{
		Location:    globalServerRegion,
		LockEnabled: false,
	}

	if globalDNSConfig != nil {
		if _, err := globalDNSConfig.Get(args.BucketName); err != nil {
			if err == dns.ErrNoEntriesFound || err == dns.ErrNotImplemented {
				// Proceed to creating a bucket.
				if err = objectAPI.MakeBucketWithLocation(ctx, args.BucketName, opts); err != nil {
					return toJSONError(ctx, err)
				}

				if err = globalDNSConfig.Put(args.BucketName); err != nil {
					objectAPI.DeleteBucket(ctx, args.BucketName, false)
					return toJSONError(ctx, err)
				}

				reply.UIVersion = Version
				return nil
			}
			return toJSONError(ctx, err)
		}
		return toJSONError(ctx, errBucketAlreadyExists)
	}

	if err := objectAPI.MakeBucketWithLocation(ctx, args.BucketName, opts); err != nil {
		return toJSONError(ctx, err, args.BucketName)
	}

	reply.UIVersion = Version

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.GetAccessKey()

	sendEvent(eventArgs{
		EventName:  event.BucketCreated,
		BucketName: args.BucketName,
		ReqParams:  reqParams,
		UserAgent:  r.UserAgent(),
		Host:       handlers.GetSourceIP(r),
	})

	return nil
}

// RemoveBucketArgs - remove bucket args.
type RemoveBucketArgs struct {
	BucketName string `json:"bucketName"`
}

// DeleteBucket - removes a bucket, must be empty.
func (web *webAPIHandlers) DeleteBucket(r *http.Request, args *RemoveBucketArgs, reply *WebGenericRep) error {
	ctx := newWebContext(r, args, "WebDeleteBucket")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// For authenticated users apply IAM policy.
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.DeleteBucketAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errAccessDenied)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	reply.UIVersion = Version

	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		core, err := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
		if err = core.RemoveBucket(ctx, args.BucketName); err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
		return nil
	}

	deleteBucket := objectAPI.DeleteBucket

	if err := deleteBucket(ctx, args.BucketName, false); err != nil {
		return toJSONError(ctx, err, args.BucketName)
	}

	globalNotificationSys.DeleteBucketMetadata(ctx, args.BucketName)

	if globalDNSConfig != nil {
		if err := globalDNSConfig.Delete(args.BucketName); err != nil {
			logger.LogIf(ctx, fmt.Errorf("Unable to delete bucket DNS entry %w, please delete it manually", err))
			return toJSONError(ctx, err)
		}
	}

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.AccessKey

	sendEvent(eventArgs{
		EventName:  event.BucketRemoved,
		BucketName: args.BucketName,
		ReqParams:  reqParams,
		UserAgent:  r.UserAgent(),
		Host:       handlers.GetSourceIP(r),
	})

	return nil
}

// ListBucketsRep - list buckets response
type ListBucketsRep struct {
	Buckets   []WebBucketInfo `json:"buckets"`
	UIVersion string          `json:"uiVersion"`
}

// WebBucketInfo container for list buckets metadata.
type WebBucketInfo struct {
	// The name of the bucket.
	Name string `json:"name"`
	// Date the bucket was created.
	CreationDate time.Time `json:"creationDate"`
}

// ListBuckets - list buckets api.
func (web *webAPIHandlers) ListBuckets(r *http.Request, args *WebGenericArgs, reply *ListBucketsRep) error {
	ctx := newWebContext(r, args, "WebListBuckets")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}
	listBuckets := objectAPI.ListBuckets

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// Set prefix value for "s3:prefix" policy conditionals.
	r.Header.Set("prefix", "")

	// Set delimiter value for "s3:delimiter" policy conditionals.
	r.Header.Set("delimiter", SlashSeparator)

	// If etcd, dns federation configured list buckets from etcd.
	if globalDNSConfig != nil && globalBucketFederation {
		dnsBuckets, err := globalDNSConfig.List()
		if err != nil && !IsErrIgnored(err,
			dns.ErrNoEntriesFound,
			dns.ErrDomainMissing) {
			return toJSONError(ctx, err)
		}
		for _, dnsRecords := range dnsBuckets {
			if globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.ListBucketAction,
				BucketName:      dnsRecords[0].Key,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      "",
				Claims:          claims.Map(),
			}) {
				reply.Buckets = append(reply.Buckets, WebBucketInfo{
					Name:         dnsRecords[0].Key,
					CreationDate: dnsRecords[0].CreationDate,
				})
			}
		}
	} else {
		buckets, err := listBuckets(ctx)
		if err != nil {
			return toJSONError(ctx, err)
		}
		for _, bucket := range buckets {
			if globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.ListBucketAction,
				BucketName:      bucket.Name,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      "",
				Claims:          claims.Map(),
			}) {
				reply.Buckets = append(reply.Buckets, WebBucketInfo{
					Name:         bucket.Name,
					CreationDate: bucket.Created,
				})
			}
		}
	}

	reply.UIVersion = Version
	return nil
}

// ListObjectsArgs - list object args.
type ListObjectsArgs struct {
	BucketName string `json:"bucketName"`
	Prefix     string `json:"prefix"`
	Marker     string `json:"marker"`
}

// ListObjectsRep - list objects response.
type ListObjectsRep struct {
	Objects   []WebObjectInfo `json:"objects"`
	Writable  bool            `json:"writable"` // Used by client to show "upload file" button.
	UIVersion string          `json:"uiVersion"`
}

// WebObjectInfo container for list objects metadata.
type WebObjectInfo struct {
	// Name of the object
	Key string `json:"name"`
	// Date and time the object was last modified.
	LastModified time.Time `json:"lastModified"`
	// Size in bytes of the object.
	Size int64 `json:"size"`
	// ContentType is mime type of the object.
	ContentType string `json:"contentType"`
}

// ListObjects - list objects api.
func (web *webAPIHandlers) ListObjects(r *http.Request, args *ListObjectsArgs, reply *ListObjectsRep) error {
	ctx := newWebContext(r, args, "WebListObjects")
	reply.UIVersion = Version
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}

	listObjects := objectAPI.ListObjects

	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		core, err := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}

		nextMarker := ""
		// Fetch all the objects
		for {
			// Let listObjects reply back the maximum from server implementation
			result, err := core.ListObjects(args.BucketName, args.Prefix, nextMarker, SlashSeparator, 1000)
			if err != nil {
				return toJSONError(ctx, err, args.BucketName)
			}

			for _, obj := range result.Contents {
				reply.Objects = append(reply.Objects, WebObjectInfo{
					Key:          obj.Key,
					LastModified: obj.LastModified,
					Size:         obj.Size,
					ContentType:  obj.ContentType,
				})
			}
			for _, p := range result.CommonPrefixes {
				reply.Objects = append(reply.Objects, WebObjectInfo{
					Key: p.Prefix,
				})
			}

			nextMarker = result.NextMarker

			// Return when there are no more objects
			if !result.IsTruncated {
				return nil
			}
		}
	}

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		if authErr == errNoAuthToken {
			// Set prefix value for "s3:prefix" policy conditionals.
			r.Header.Set("prefix", args.Prefix)

			// Set delimiter value for "s3:delimiter" policy conditionals.
			r.Header.Set("delimiter", SlashSeparator)

			// Check if anonymous (non-owner) has access to download objects.
			readable := globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.ListBucketAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
			})

			// Check if anonymous (non-owner) has access to upload objects.
			writable := globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.PutObjectAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      args.Prefix + SlashSeparator,
			})

			reply.Writable = writable
			if !readable {
				// Error out if anonymous user (non-owner) has no access to download or upload objects
				if !writable {
					return errAccessDenied
				}
				// return empty object list if access is write only
				return nil
			}
		} else {
			return toJSONError(ctx, authErr)
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		// Set prefix value for "s3:prefix" policy conditionals.
		r.Header.Set("prefix", args.Prefix)

		// Set delimiter value for "s3:delimiter" policy conditionals.
		r.Header.Set("delimiter", SlashSeparator)

		readable := globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.ListBucketAction,
			BucketName:      args.BucketName,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			Claims:          claims.Map(),
		})

		writable := globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.PutObjectAction,
			BucketName:      args.BucketName,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      args.Prefix + SlashSeparator,
			Claims:          claims.Map(),
		})

		reply.Writable = writable
		if !readable {
			// Error out if anonymous user (non-owner) has no access to download or upload objects
			if !writable {
				return errAccessDenied
			}
			// return empty object list if access is write only
			return nil
		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	nextMarker := ""
	// Fetch all the objects
	for {
		// Limit browser to '1000' batches to be more responsive, scrolling friendly.
		// Also don't change the maxKeys value silly GCS SDKs do not honor maxKeys
		// values to be '-1'
		lo, err := listObjects(ctx, args.BucketName, args.Prefix, nextMarker, SlashSeparator, 1000)
		if err != nil {
			return &json2.Error{Message: err.Error()}
		}

		nextMarker = lo.NextMarker
		for i := range lo.Objects {
			lo.Objects[i].Size, err = lo.Objects[i].GetActualSize()
			if err != nil {
				return toJSONError(ctx, err)
			}
		}

		for _, obj := range lo.Objects {
			reply.Objects = append(reply.Objects, WebObjectInfo{
				Key:          obj.Name,
				LastModified: obj.ModTime,
				Size:         obj.Size,
				ContentType:  obj.ContentType,
			})
		}
		for _, prefix := range lo.Prefixes {
			reply.Objects = append(reply.Objects, WebObjectInfo{
				Key: prefix,
			})
		}

		// Return when there are no more objects
		if !lo.IsTruncated {
			return nil
		}
	}
}

// RemoveObjectArgs - args to remove an object, JSON will look like.
//
// {
//     "bucketname": "testbucket",
//     "objects": [
//         "photos/hawaii/",
//         "photos/maldives/",
//         "photos/sanjose.jpg"
//     ]
// }
type RemoveObjectArgs struct {
	Objects    []string `json:"objects"`    // Contains objects, prefixes.
	BucketName string   `json:"bucketname"` // Contains bucket name.
}

// RemoveObject - removes an object, or all the objects at a given prefix.
func (web *webAPIHandlers) RemoveObject(r *http.Request, args *RemoveObjectArgs, reply *WebGenericRep) error {
	ctx := newWebContext(r, args, "WebRemoveObject")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}

	deleteObjects := objectAPI.DeleteObjects
	if web.CacheAPI() != nil {
		deleteObjects = web.CacheAPI().DeleteObjects
	}
	getObjectInfoFn := objectAPI.GetObjectInfo
	if web.CacheAPI() != nil {
		getObjectInfoFn = web.CacheAPI().GetObjectInfo
	}

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if all objects are allowed to be deleted anonymously
			for _, object := range args.Objects {
				if !globalPolicySys.IsAllowed(policy.Args{
					Action:          policy.DeleteObjectAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", "", nil),
					IsOwner:         false,
					ObjectName:      object,
				}) {
					return toJSONError(ctx, errAuthentication)
				}
			}
		} else {
			return toJSONError(ctx, authErr)
		}
	}

	if args.BucketName == "" || len(args.Objects) == 0 {
		return toJSONError(ctx, errInvalidArgument)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	reply.UIVersion = Version
	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		core, err := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
		objectsCh := make(chan miniogo.ObjectInfo)

		// Send object names that are needed to be removed to objectsCh
		go func() {
			defer close(objectsCh)

			for _, objectName := range args.Objects {
				objectsCh <- miniogo.ObjectInfo{
					Key: objectName,
				}
			}
		}()

		for resp := range core.RemoveObjects(ctx, args.BucketName, objectsCh, minio.RemoveObjectsOptions{}) {
			if resp.Err != nil {
				return toJSONError(ctx, resp.Err, args.BucketName, resp.ObjectName)
			}
		}
		return nil
	}

	opts := ObjectOptions{
		Versioned:        globalBucketVersioningSys.Enabled(args.BucketName),
		VersionSuspended: globalBucketVersioningSys.Suspended(args.BucketName),
	}
	var (
		err           error
		replicateSync bool
	)

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.GetAccessKey()
	sourceIP := handlers.GetSourceIP(r)

next:
	for _, objectName := range args.Objects {
		// If not a directory, remove the object.
		if !HasSuffix(objectName, SlashSeparator) && objectName != "" {
			// Check permissions for non-anonymous user.
			if authErr != errNoAuthToken {
				if !globalIAMSys.IsAllowed(iampolicy.Args{
					AccountName:     claims.AccessKey,
					Action:          iampolicy.DeleteObjectAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
					IsOwner:         owner,
					ObjectName:      objectName,
					Claims:          claims.Map(),
				}) {
					return toJSONError(ctx, errAccessDenied)
				}
			}

			if authErr == errNoAuthToken {
				// Check if object is allowed to be deleted anonymously.
				if !globalPolicySys.IsAllowed(policy.Args{
					Action:          policy.DeleteObjectAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", "", nil),
					IsOwner:         false,
					ObjectName:      objectName,
				}) {
					return toJSONError(ctx, errAccessDenied)
				}
			}
			var (
				replicateDel, hasLifecycleConfig bool
				goi                              ObjectInfo
				gerr                             error
			)
			if _, err := globalBucketMetadataSys.GetLifecycleConfig(args.BucketName); err == nil {
				hasLifecycleConfig = true
			}
			os := newObjSweeper(args.BucketName, objectName)
			opts = os.GetOpts()
			if hasReplicationRules(ctx, args.BucketName, []ObjectToDelete{{ObjectName: objectName}}) || hasLifecycleConfig {
				goi, gerr = getObjectInfoFn(ctx, args.BucketName, objectName, opts)
				if gerr == nil {
					os.SetTransitionState(goi)
				}
				if replicateDel, replicateSync = checkReplicateDelete(ctx, args.BucketName, ObjectToDelete{
					ObjectName: objectName,
					VersionID:  goi.VersionID,
				}, goi, gerr); replicateDel {
					opts.DeleteMarkerReplicationStatus = string(replication.Pending)
					opts.DeleteMarker = true
				}
			}

			deleteObject := objectAPI.DeleteObject
			if web.CacheAPI() != nil {
				deleteObject = web.CacheAPI().DeleteObject
			}

			oi, err := deleteObject(ctx, args.BucketName, objectName, opts)
			if err != nil {
				switch err.(type) {
				case BucketNotFound:
					return toJSONError(ctx, err)
				}
			}
			if oi.Name == "" {
				logger.LogIf(ctx, err)
				continue
			}

			eventName := event.ObjectRemovedDelete
			if oi.DeleteMarker {
				eventName = event.ObjectRemovedDeleteMarkerCreated
			}

			// Notify object deleted event.
			sendEvent(eventArgs{
				EventName:  eventName,
				BucketName: args.BucketName,
				Object:     oi,
				ReqParams:  reqParams,
				UserAgent:  r.UserAgent(),
				Host:       sourceIP,
			})

			if replicateDel {
				dobj := DeletedObjectReplicationInfo{
					DeletedObject: DeletedObject{
						ObjectName:                    objectName,
						DeleteMarkerVersionID:         oi.VersionID,
						DeleteMarkerReplicationStatus: string(oi.ReplicationStatus),
						DeleteMarkerMTime:             DeleteMarkerMTime{oi.ModTime},
						DeleteMarker:                  oi.DeleteMarker,
						VersionPurgeStatus:            oi.VersionPurgeStatus,
					},
					Bucket: args.BucketName,
				}
				scheduleReplicationDelete(ctx, dobj, objectAPI, replicateSync)
			}

			logger.LogIf(ctx, err)
			logger.LogIf(ctx, os.Sweep())
			continue
		}

		if authErr == errNoAuthToken {
			// Check if object is allowed to be deleted anonymously
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          iampolicy.DeleteObjectAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      objectName,
			}) {
				return toJSONError(ctx, errAccessDenied)
			}
		} else {
			if !globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.DeleteObjectAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      objectName,
				Claims:          claims.Map(),
			}) {
				return toJSONError(ctx, errAccessDenied)
			}
		}

		// Allocate new results channel to receive ObjectInfo.
		objInfoCh := make(chan ObjectInfo)

		// Walk through all objects
		if err = objectAPI.Walk(ctx, args.BucketName, objectName, objInfoCh, ObjectOptions{}); err != nil {
			break next
		}

		for {
			var objects []ObjectToDelete
			for obj := range objInfoCh {
				if len(objects) == maxDeleteList {
					// Reached maximum delete requests, attempt a delete for now.
					break
				}
				if obj.ReplicationStatus == replication.Replica {
					if authErr == errNoAuthToken {
						// Check if object is allowed to be deleted anonymously
						if !globalPolicySys.IsAllowed(policy.Args{
							Action:          iampolicy.ReplicateDeleteAction,
							BucketName:      args.BucketName,
							ConditionValues: getConditionValues(r, "", "", nil),
							IsOwner:         false,
							ObjectName:      objectName,
						}) {
							return toJSONError(ctx, errAccessDenied)
						}
					} else {
						if !globalIAMSys.IsAllowed(iampolicy.Args{
							AccountName:     claims.AccessKey,
							Action:          iampolicy.ReplicateDeleteAction,
							BucketName:      args.BucketName,
							ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
							IsOwner:         owner,
							ObjectName:      objectName,
							Claims:          claims.Map(),
						}) {
							return toJSONError(ctx, errAccessDenied)
						}
					}
				}
				replicateDel, _ := checkReplicateDelete(ctx, args.BucketName, ObjectToDelete{ObjectName: obj.Name, VersionID: obj.VersionID}, obj, nil)
				// since versioned delete is not available on web browser, yet - this is a simple DeleteMarker replication
				objToDel := ObjectToDelete{ObjectName: obj.Name}
				if replicateDel {
					objToDel.DeleteMarkerReplicationStatus = string(replication.Pending)
				}

				objects = append(objects, objToDel)
			}

			// Nothing to do.
			if len(objects) == 0 {
				break next
			}

			// Deletes a list of objects.
			deletedObjects, errs := deleteObjects(ctx, args.BucketName, objects, opts)
			for i, err := range errs {
				if err != nil && !isErrObjectNotFound(err) {
					deletedObjects[i].DeleteMarkerReplicationStatus = objects[i].DeleteMarkerReplicationStatus
					deletedObjects[i].VersionPurgeStatus = objects[i].VersionPurgeStatus
				}
				if err != nil {
					logger.LogIf(ctx, err)
					break next
				}
			}
			// Notify deleted event for objects.
			for _, dobj := range deletedObjects {
				objInfo := ObjectInfo{
					Name:      dobj.ObjectName,
					VersionID: dobj.VersionID,
				}
				if dobj.DeleteMarker {
					objInfo = ObjectInfo{
						Name:         dobj.ObjectName,
						DeleteMarker: dobj.DeleteMarker,
						VersionID:    dobj.DeleteMarkerVersionID,
					}
				}
				sendEvent(eventArgs{
					EventName:  event.ObjectRemovedDelete,
					BucketName: args.BucketName,
					Object:     objInfo,
					ReqParams:  reqParams,
					UserAgent:  r.UserAgent(),
					Host:       sourceIP,
				})
				if dobj.DeleteMarkerReplicationStatus == string(replication.Pending) || dobj.VersionPurgeStatus == Pending {
					dv := DeletedObjectReplicationInfo{
						DeletedObject: dobj,
						Bucket:        args.BucketName,
					}
					scheduleReplicationDelete(ctx, dv, objectAPI, replicateSync)
				}
			}
		}
	}

	if err != nil && !isErrObjectNotFound(err) && !isErrVersionNotFound(err) {
		// Ignore object not found error.
		return toJSONError(ctx, err, args.BucketName, "")
	}

	return nil
}

// LoginArgs - login arguments.
type LoginArgs struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

// LoginRep - login reply.
type LoginRep struct {
	Token     string `json:"token"`
	UIVersion string `json:"uiVersion"`
}

// Login - user login handler.
func (web *webAPIHandlers) Login(r *http.Request, args *LoginArgs, reply *LoginRep) error {
	ctx := newWebContext(r, args, "WebLogin")
	token, err := authenticateWeb(args.Username, args.Password)
	if err != nil {
		return toJSONError(ctx, err)
	}

	reply.Token = token
	reply.UIVersion = Version
	return nil
}

// SetAuthArgs - argument for SetAuth
type SetAuthArgs struct {
	CurrentAccessKey string `json:"currentAccessKey"`
	CurrentSecretKey string `json:"currentSecretKey"`
	NewAccessKey     string `json:"newAccessKey"`
	NewSecretKey     string `json:"newSecretKey"`
}

// SetAuthReply - reply for SetAuth
type SetAuthReply struct {
	Token       string            `json:"token"`
	UIVersion   string            `json:"uiVersion"`
	PeerErrMsgs map[string]string `json:"peerErrMsgs"`
}

// SetAuth - Set accessKey and secretKey credentials.
func (web *webAPIHandlers) SetAuth(r *http.Request, args *SetAuthArgs, reply *SetAuthReply) error {
	ctx := newWebContext(r, args, "WebSetAuth")
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	if owner {
		// Owner is not allowed to change credentials through browser.
		return toJSONError(ctx, errChangeCredNotAllowed)
	}

	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.CreateUserAdminAction,
		IsOwner:         false,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		Claims:          claims.Map(),
		DenyOnly:        true,
	}) {
		return toJSONError(ctx, errChangeCredNotAllowed)
	}

	// for IAM users, access key cannot be updated
	// claims.AccessKey is used instead of accesskey from args
	prevCred, ok := globalIAMSys.GetUser(claims.AccessKey)
	if !ok {
		return errInvalidAccessKeyID
	}

	// Throw error when wrong secret key is provided
	if subtle.ConstantTimeCompare([]byte(prevCred.SecretKey), []byte(args.CurrentSecretKey)) != 1 {
		return errIncorrectCreds
	}

	creds, err := auth.CreateCredentials(claims.AccessKey, args.NewSecretKey)
	if err != nil {
		return toJSONError(ctx, err)
	}

	err = globalIAMSys.SetUserSecretKey(creds.AccessKey, creds.SecretKey)
	if err != nil {
		return toJSONError(ctx, err)
	}

	reply.Token, err = authenticateWeb(creds.AccessKey, creds.SecretKey)
	if err != nil {
		return toJSONError(ctx, err)
	}

	reply.UIVersion = Version

	return nil
}

// URLTokenReply contains the reply for CreateURLToken.
type URLTokenReply struct {
	Token     string `json:"token"`
	UIVersion string `json:"uiVersion"`
}

// CreateURLToken creates a URL token (short-lived) for GET requests.
func (web *webAPIHandlers) CreateURLToken(r *http.Request, args *WebGenericArgs, reply *URLTokenReply) error {
	ctx := newWebContext(r, args, "WebCreateURLToken")
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	creds := globalActiveCred
	if !owner {
		var ok bool
		creds, ok = globalIAMSys.GetUser(claims.AccessKey)
		if !ok {
			return toJSONError(ctx, errInvalidAccessKeyID)
		}
	}

	if creds.SessionToken != "" {
		// Use the same session token for URL token.
		reply.Token = creds.SessionToken
	} else {
		token, err := authenticateURL(creds.AccessKey, creds.SecretKey)
		if err != nil {
			return toJSONError(ctx, err)
		}
		reply.Token = token
	}

	reply.UIVersion = Version
	return nil
}

// Upload - file upload handler.
func (web *webAPIHandlers) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebUpload")

	// obtain the claims here if possible, for audit logging.
	claims, owner, authErr := webRequestAuthenticate(r)

	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	retPerms := ErrAccessDenied
	holdPerms := ErrAccessDenied
	replPerms := ErrAccessDenied
	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to upload objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.PutObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				writeWebErrorResponse(w, errAuthentication)
				return
			}
		} else {
			writeWebErrorResponse(w, authErr)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.PutObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			writeWebErrorResponse(w, errAuthentication)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.PutObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			retPerms = ErrNone
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.PutObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			holdPerms = ErrNone
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetReplicationConfigurationAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {
			replPerms = ErrNone
		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	// Check if bucket encryption is enabled
	sseConfig, _ := globalBucketSSEConfigSys.Get(bucket)
	sseConfig.Apply(r.Header, globalAutoEncryption)

	// Require Content-Length to be set in the request
	size := r.ContentLength
	if size < 0 {
		writeWebErrorResponse(w, errSizeUnspecified)
		return
	}

	if err := enforceBucketQuota(ctx, bucket, size); err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	// Extract incoming metadata if any.
	metadata, err := extractMetadata(ctx, r)
	if err != nil {
		writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL, guessIsBrowserReq(r))
		return
	}

	var pReader *PutObjReader
	var reader io.Reader = r.Body
	actualSize := size

	hashReader, err := hash.NewReader(reader, size, "", "", actualSize)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	if objectAPI.IsCompressionSupported() && isCompressible(r.Header, object) && size > 0 {
		// Storing the compression metadata.
		metadata[ReservedMetadataPrefix+"compression"] = compressionAlgorithmV2
		metadata[ReservedMetadataPrefix+"actual-size"] = strconv.FormatInt(actualSize, 10)

		actualReader, err := hash.NewReader(reader, actualSize, "", "", actualSize)
		if err != nil {
			writeWebErrorResponse(w, err)
			return
		}

		// Set compression metrics.
		size = -1 // Since compressed size is un-predictable.
		s2c := newS2CompressReader(actualReader, actualSize)
		defer s2c.Close()
		reader = etag.Wrap(s2c, actualReader)
		hashReader, err = hash.NewReader(reader, size, "", "", actualSize)
		if err != nil {
			writeWebErrorResponse(w, err)
			return
		}
	}

	mustReplicate, sync := mustReplicateWeb(ctx, r, bucket, object, metadata, replication.StatusType(""), replPerms)
	if mustReplicate {
		metadata[xhttp.AmzBucketReplicationStatus] = string(replication.Pending)
	}
	pReader = NewPutObjReader(hashReader)
	// get gateway encryption options
	opts, err := putOpts(ctx, r, bucket, object, metadata)
	if err != nil {
		writeErrorResponseHeadersOnly(w, toAPIError(ctx, err))
		return
	}

	if objectAPI.IsEncryptionSupported() {
		if _, ok := crypto.IsRequested(r.Header); ok && !HasSuffix(object, SlashSeparator) { // handle SSE requests
			var (
				objectEncryptionKey crypto.ObjectKey
				encReader           io.Reader
			)
			encReader, objectEncryptionKey, err = EncryptRequest(hashReader, r, bucket, object, metadata)
			if err != nil {
				writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL, guessIsBrowserReq(r))
				return
			}
			info := ObjectInfo{Size: size}
			// do not try to verify encrypted content
			hashReader, err = hash.NewReader(etag.Wrap(encReader, hashReader), info.EncryptedSize(), "", "", size)
			if err != nil {
				writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL, guessIsBrowserReq(r))
				return
			}
			pReader, err = pReader.WithEncryption(hashReader, &objectEncryptionKey)
			if err != nil {
				writeErrorResponse(ctx, w, toAPIError(ctx, err), r.URL, guessIsBrowserReq(r))
				return
			}
		}
	}

	// Ensure that metadata does not contain sensitive information
	crypto.RemoveSensitiveEntries(metadata)

	putObject := objectAPI.PutObject
	getObjectInfo := objectAPI.GetObjectInfo
	if web.CacheAPI() != nil {
		putObject = web.CacheAPI().PutObject
		getObjectInfo = web.CacheAPI().GetObjectInfo
	}

	// enforce object retention rules
	retentionMode, retentionDate, _, s3Err := checkPutObjectLockAllowed(ctx, r, bucket, object, getObjectInfo, retPerms, holdPerms)
	if s3Err != ErrNone {
		writeErrorResponse(ctx, w, errorCodes.ToAPIErr(s3Err), r.URL, guessIsBrowserReq(r))
		return
	}
	if retentionMode != "" {
		opts.UserDefined[strings.ToLower(xhttp.AmzObjectLockMode)] = string(retentionMode)
		opts.UserDefined[strings.ToLower(xhttp.AmzObjectLockRetainUntilDate)] = retentionDate.UTC().Format(iso8601TimeFormat)
	}

	os := newObjSweeper(bucket, object)
	// Get appropriate object info to identify the remote object to delete
	goiOpts := os.GetOpts()
	if goi, gerr := getObjectInfo(ctx, bucket, object, goiOpts); gerr == nil {
		os.SetTransitionState(goi)
	}

	objInfo, err := putObject(GlobalContext, bucket, object, pReader, opts)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}
	if objectAPI.IsEncryptionSupported() {
		switch kind, _ := crypto.IsEncrypted(objInfo.UserDefined); kind {
		case crypto.S3:
			w.Header().Set(xhttp.AmzServerSideEncryption, xhttp.AmzEncryptionAES)
		case crypto.S3KMS:
			w.Header().Set(xhttp.AmzServerSideEncryption, xhttp.AmzEncryptionKMS)
			if kmsCtx, ok := objInfo.UserDefined[crypto.MetaContext]; ok {
				w.Header().Set(xhttp.AmzServerSideEncryptionKmsContext, kmsCtx)
			}
		case crypto.SSEC:
			w.Header().Set(xhttp.AmzServerSideEncryptionCustomerAlgorithm, r.Header.Get(xhttp.AmzServerSideEncryptionCustomerAlgorithm))
			w.Header().Set(xhttp.AmzServerSideEncryptionCustomerKeyMD5, r.Header.Get(xhttp.AmzServerSideEncryptionCustomerKeyMD5))
		}
	}
	if mustReplicate {
		scheduleReplication(ctx, objInfo.Clone(), objectAPI, sync, replication.ObjectReplicationType)
	}
	logger.LogIf(ctx, os.Sweep())

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.GetAccessKey()

	// Notify object created event.
	sendEvent(eventArgs{
		EventName:    event.ObjectCreatedPut,
		BucketName:   bucket,
		Object:       objInfo,
		ReqParams:    reqParams,
		RespElements: extractRespElements(w),
		UserAgent:    r.UserAgent(),
		Host:         handlers.GetSourceIP(r),
	})
}

// Download - file download handler.
func (web *webAPIHandlers) Download(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebDownload")

	claims, owner, authErr := webTokenAuthenticate(r.URL.Query().Get("token"))
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	getRetPerms := ErrAccessDenied
	legalHoldPerms := ErrAccessDenied

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				writeWebErrorResponse(w, errAuthentication)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				getRetPerms = ErrNone
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				legalHoldPerms = ErrNone
			}
		} else {
			writeWebErrorResponse(w, authErr)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			writeWebErrorResponse(w, errAuthentication)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			getRetPerms = ErrNone
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			legalHoldPerms = ErrNone
		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}

	var opts ObjectOptions
	gr, err := getObjectNInfo(ctx, bucket, object, nil, r.Header, readLock, opts)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}
	defer gr.Close()

	objInfo := gr.ObjInfo

	// filter object lock metadata if permission does not permit
	objInfo.UserDefined = objectlock.FilterObjectLockMetadata(objInfo.UserDefined, getRetPerms != ErrNone, legalHoldPerms != ErrNone)

	if objectAPI.IsEncryptionSupported() {
		if _, err = DecryptObjectInfo(&objInfo, r); err != nil {
			writeWebErrorResponse(w, err)
			return
		}
	}

	// Set encryption response headers
	if objectAPI.IsEncryptionSupported() {
		switch kind, _ := crypto.IsEncrypted(objInfo.UserDefined); kind {
		case crypto.S3:
			w.Header().Set(xhttp.AmzServerSideEncryption, xhttp.AmzEncryptionAES)
		case crypto.S3KMS:
			w.Header().Set(xhttp.AmzServerSideEncryption, xhttp.AmzEncryptionKMS)
			w.Header().Set(xhttp.AmzServerSideEncryptionKmsID, objInfo.UserDefined[crypto.MetaKeyID])
			if kmsCtx, ok := objInfo.UserDefined[crypto.MetaContext]; ok {
				w.Header().Set(xhttp.AmzServerSideEncryptionKmsContext, kmsCtx)
			}
		case crypto.SSEC:
			w.Header().Set(xhttp.AmzServerSideEncryptionCustomerAlgorithm, r.Header.Get(xhttp.AmzServerSideEncryptionCustomerAlgorithm))
			w.Header().Set(xhttp.AmzServerSideEncryptionCustomerKeyMD5, r.Header.Get(xhttp.AmzServerSideEncryptionCustomerKeyMD5))
		}
	}

	// Set Parts Count Header
	if opts.PartNumber > 0 && len(objInfo.Parts) > 0 {
		setPartsCountHeaders(w, objInfo)
	}

	if err = setObjectHeaders(w, objInfo, nil, opts); err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	// Add content disposition.
	w.Header().Set(xhttp.ContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", path.Base(objInfo.Name)))

	setHeadGetRespHeaders(w, r.URL.Query())

	httpWriter := ioutil.WriteOnClose(w)

	// Write object content to response body
	if _, err = io.Copy(httpWriter, gr); err != nil {
		if !httpWriter.HasWritten() { // write error response only if no data or headers has been written to client yet
			writeWebErrorResponse(w, err)
		}
		return
	}

	if err = httpWriter.Close(); err != nil {
		if !httpWriter.HasWritten() { // write error response only if no data or headers has been written to client yet
			writeWebErrorResponse(w, err)
			return
		}
	}

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.GetAccessKey()

	// Notify object accessed via a GET request.
	sendEvent(eventArgs{
		EventName:    event.ObjectAccessedGet,
		BucketName:   bucket,
		Object:       objInfo,
		ReqParams:    reqParams,
		RespElements: extractRespElements(w),
		UserAgent:    r.UserAgent(),
		Host:         handlers.GetSourceIP(r),
	})
}

// DownloadZipArgs - Argument for downloading a bunch of files as a zip file.
// JSON will look like:
// '{"bucketname":"testbucket","prefix":"john/pics/","objects":["hawaii/","maldives/","sanjose.jpg"]}'
type DownloadZipArgs struct {
	Objects    []string `json:"objects"`    // can be files or sub-directories
	Prefix     string   `json:"prefix"`     // current directory in the browser-ui
	BucketName string   `json:"bucketname"` // bucket name.
}

// Takes a list of objects and creates a zip file that sent as the response body.
func (web *webAPIHandlers) DownloadZip(w http.ResponseWriter, r *http.Request) {
	host := handlers.GetSourceIP(r)

	claims, owner, authErr := webTokenAuthenticate(r.URL.Query().Get("token"))

	ctx := newContext(r, w, "WebDownloadZip")
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	// Auth is done after reading the body to accommodate for anonymous requests
	// when bucket policy is enabled.
	var args DownloadZipArgs
	tenKB := 10 * 1024 // To limit r.Body to take care of misbehaving anonymous client.
	decodeErr := json.NewDecoder(io.LimitReader(r.Body, int64(tenKB))).Decode(&args)
	if decodeErr != nil {
		writeWebErrorResponse(w, decodeErr)
		return
	}

	var getRetPerms []APIErrorCode
	var legalHoldPerms []APIErrorCode

	if authErr != nil {
		if authErr == errNoAuthToken {
			for _, object := range args.Objects {
				// Check if anonymous (non-owner) has access to download objects.
				if !globalPolicySys.IsAllowed(policy.Args{
					Action:          policy.GetObjectAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", "", nil),
					IsOwner:         false,
					ObjectName:      pathJoin(args.Prefix, object),
				}) {
					writeWebErrorResponse(w, errAuthentication)
					return
				}
				retentionPerm := ErrAccessDenied
				if globalPolicySys.IsAllowed(policy.Args{
					Action:          policy.GetObjectRetentionAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", "", nil),
					IsOwner:         false,
					ObjectName:      pathJoin(args.Prefix, object),
				}) {
					retentionPerm = ErrNone
				}
				getRetPerms = append(getRetPerms, retentionPerm)

				legalHoldPerm := ErrAccessDenied
				if globalPolicySys.IsAllowed(policy.Args{
					Action:          policy.GetObjectLegalHoldAction,
					BucketName:      args.BucketName,
					ConditionValues: getConditionValues(r, "", "", nil),
					IsOwner:         false,
					ObjectName:      pathJoin(args.Prefix, object),
				}) {
					legalHoldPerm = ErrNone
				}
				legalHoldPerms = append(legalHoldPerms, legalHoldPerm)
			}
		} else {
			writeWebErrorResponse(w, authErr)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		for _, object := range args.Objects {
			if !globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.GetObjectAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      pathJoin(args.Prefix, object),
				Claims:          claims.Map(),
			}) {
				writeWebErrorResponse(w, errAuthentication)
				return
			}
			retentionPerm := ErrAccessDenied
			if globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.GetObjectRetentionAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      pathJoin(args.Prefix, object),
				Claims:          claims.Map(),
			}) {
				retentionPerm = ErrNone
			}
			getRetPerms = append(getRetPerms, retentionPerm)

			legalHoldPerm := ErrAccessDenied
			if globalIAMSys.IsAllowed(iampolicy.Args{
				AccountName:     claims.AccessKey,
				Action:          iampolicy.GetObjectLegalHoldAction,
				BucketName:      args.BucketName,
				ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
				IsOwner:         owner,
				ObjectName:      pathJoin(args.Prefix, object),
				Claims:          claims.Map(),
			}) {
				legalHoldPerm = ErrNone
			}
			legalHoldPerms = append(legalHoldPerms, legalHoldPerm)
		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}

	archive := zip.NewWriter(w)
	defer archive.Close()

	reqParams := extractReqParams(r)
	reqParams["accessKey"] = claims.GetAccessKey()
	respElements := extractRespElements(w)

	for i, object := range args.Objects {
		if contextCanceled(ctx) {
			return
		}
		// Writes compressed object file to the response.
		zipit := func(objectName string) error {
			var opts ObjectOptions
			gr, err := getObjectNInfo(ctx, args.BucketName, objectName, nil, r.Header, readLock, opts)
			if err != nil {
				return err
			}
			defer gr.Close()

			info := gr.ObjInfo
			// filter object lock metadata if permission does not permit
			info.UserDefined = objectlock.FilterObjectLockMetadata(info.UserDefined, getRetPerms[i] != ErrNone, legalHoldPerms[i] != ErrNone)
			// For reporting, set the file size to the uncompressed size.
			info.Size, err = info.GetActualSize()
			if err != nil {
				return err
			}
			header := &zip.FileHeader{
				Name:               strings.TrimPrefix(objectName, args.Prefix),
				Method:             zip.Deflate,
				Flags:              1 << 11,
				Modified:           info.ModTime,
				UncompressedSize64: uint64(info.Size),
			}
			if info.Size < 20 || hasStringSuffixInSlice(info.Name, standardExcludeCompressExtensions) || hasPattern(standardExcludeCompressContentTypes, info.ContentType) {
				// We strictly disable compression for standard extensions/content-types.
				header.Method = zip.Store
			}
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			// Write object content to response body
			if _, err = io.Copy(writer, gr); err != nil {
				return err
			}

			// Notify object accessed via a GET request.
			sendEvent(eventArgs{
				EventName:    event.ObjectAccessedGet,
				BucketName:   args.BucketName,
				Object:       info,
				ReqParams:    reqParams,
				RespElements: respElements,
				UserAgent:    r.UserAgent(),
				Host:         host,
			})

			return nil
		}

		if !HasSuffix(object, SlashSeparator) {
			// If not a directory, compress the file and write it to response.
			err := zipit(pathJoin(args.Prefix, object))
			if err != nil {
				logger.LogIf(ctx, err)
				return
			}
			continue
		}

		objInfoCh := make(chan ObjectInfo)

		// Walk through all objects
		if err := objectAPI.Walk(ctx, args.BucketName, pathJoin(args.Prefix, object), objInfoCh, ObjectOptions{}); err != nil {
			logger.LogIf(ctx, err)
			continue
		}

		for obj := range objInfoCh {
			if err := zipit(obj.Name); err != nil {
				logger.LogIf(ctx, err)
				continue
			}
		}
	}
}

// GetBucketPolicyArgs - get bucket policy args.
type GetBucketPolicyArgs struct {
	BucketName string `json:"bucketName"`
	Prefix     string `json:"prefix"`
}

// GetBucketPolicyRep - get bucket policy reply.
type GetBucketPolicyRep struct {
	UIVersion string                     `json:"uiVersion"`
	Policy    miniogopolicy.BucketPolicy `json:"policy"`
}

// GetBucketPolicy - get bucket policy for the requested prefix.
func (web *webAPIHandlers) GetBucketPolicy(r *http.Request, args *GetBucketPolicyArgs, reply *GetBucketPolicyRep) error {
	ctx := newWebContext(r, args, "WebGetBucketPolicy")

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// For authenticated users apply IAM policy.
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.GetBucketPolicyAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errAccessDenied)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	var policyInfo = &miniogopolicy.BucketAccessPolicy{Version: "2012-10-17"}
	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		client, rerr := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if rerr != nil {
			return toJSONError(ctx, rerr, args.BucketName)
		}
		policyStr, err := client.GetBucketPolicy(ctx, args.BucketName)
		if err != nil {
			return toJSONError(ctx, rerr, args.BucketName)
		}
		bucketPolicy, err := policy.ParseConfig(strings.NewReader(policyStr), args.BucketName)
		if err != nil {
			return toJSONError(ctx, rerr, args.BucketName)
		}
		policyInfo, err = PolicyToBucketAccessPolicy(bucketPolicy)
		if err != nil {
			// This should not happen.
			return toJSONError(ctx, err, args.BucketName)
		}
	} else {
		bucketPolicy, err := globalPolicySys.Get(args.BucketName)
		if err != nil {
			if _, ok := err.(BucketPolicyNotFound); !ok {
				return toJSONError(ctx, err, args.BucketName)
			}
		}

		policyInfo, err = PolicyToBucketAccessPolicy(bucketPolicy)
		if err != nil {
			// This should not happen.
			return toJSONError(ctx, err, args.BucketName)
		}
	}

	reply.UIVersion = Version
	reply.Policy = miniogopolicy.GetPolicy(policyInfo.Statements, args.BucketName, args.Prefix)

	return nil
}

// ListAllBucketPoliciesArgs - get all bucket policies.
type ListAllBucketPoliciesArgs struct {
	BucketName string `json:"bucketName"`
}

// BucketAccessPolicy - Collection of canned bucket policy at a given prefix.
type BucketAccessPolicy struct {
	Bucket string                     `json:"bucket"`
	Prefix string                     `json:"prefix"`
	Policy miniogopolicy.BucketPolicy `json:"policy"`
}

// ListAllBucketPoliciesRep - get all bucket policy reply.
type ListAllBucketPoliciesRep struct {
	UIVersion string               `json:"uiVersion"`
	Policies  []BucketAccessPolicy `json:"policies"`
}

// ListAllBucketPolicies - get all bucket policy.
func (web *webAPIHandlers) ListAllBucketPolicies(r *http.Request, args *ListAllBucketPoliciesArgs, reply *ListAllBucketPoliciesRep) error {
	ctx := newWebContext(r, args, "WebListAllBucketPolicies")
	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// For authenticated users apply IAM policy.
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.GetBucketPolicyAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errAccessDenied)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	var policyInfo = new(miniogopolicy.BucketAccessPolicy)
	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		core, rerr := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if rerr != nil {
			return toJSONError(ctx, rerr, args.BucketName)
		}
		var policyStr string
		policyStr, err = core.Client.GetBucketPolicy(ctx, args.BucketName)
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
		if policyStr != "" {
			if err = json.Unmarshal([]byte(policyStr), policyInfo); err != nil {
				return toJSONError(ctx, err, args.BucketName)
			}
		}
	} else {
		bucketPolicy, err := globalPolicySys.Get(args.BucketName)
		if err != nil {
			if _, ok := err.(BucketPolicyNotFound); !ok {
				return toJSONError(ctx, err, args.BucketName)
			}
		}
		policyInfo, err = PolicyToBucketAccessPolicy(bucketPolicy)
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
	}

	reply.UIVersion = Version
	for prefix, policy := range miniogopolicy.GetPolicies(policyInfo.Statements, args.BucketName, "") {
		bucketName, objectPrefix := path2BucketObject(prefix)
		objectPrefix = strings.TrimSuffix(objectPrefix, "*")
		reply.Policies = append(reply.Policies, BucketAccessPolicy{
			Bucket: bucketName,
			Prefix: objectPrefix,
			Policy: policy,
		})
	}

	return nil
}

// SetBucketPolicyWebArgs - set bucket policy args.
type SetBucketPolicyWebArgs struct {
	BucketName string `json:"bucketName"`
	Prefix     string `json:"prefix"`
	Policy     string `json:"policy"`
}

// SetBucketPolicy - set bucket policy.
func (web *webAPIHandlers) SetBucketPolicy(r *http.Request, args *SetBucketPolicyWebArgs, reply *WebGenericRep) error {
	ctx := newWebContext(r, args, "WebSetBucketPolicy")
	objectAPI := web.ObjectAPI()
	reply.UIVersion = Version

	if objectAPI == nil {
		return toJSONError(ctx, errServerNotInitialized)
	}

	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}

	// For authenticated users apply IAM policy.
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.PutBucketPolicyAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errAccessDenied)
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	policyType := miniogopolicy.BucketPolicy(args.Policy)
	if !policyType.IsValidBucketPolicy() {
		return &json2.Error{
			Message: "Invalid policy type " + args.Policy,
		}
	}

	if isRemoteCallRequired(ctx, args.BucketName, objectAPI) {
		sr, err := globalDNSConfig.Get(args.BucketName)
		if err != nil {
			if err == dns.ErrNoEntriesFound {
				return toJSONError(ctx, BucketNotFound{
					Bucket: args.BucketName,
				}, args.BucketName)
			}
			return toJSONError(ctx, err, args.BucketName)
		}
		core, rerr := getRemoteInstanceClient(r, getHostFromSrv(sr))
		if rerr != nil {
			return toJSONError(ctx, rerr, args.BucketName)
		}
		var policyStr string
		// Use the abstracted API instead of core, such that
		// NoSuchBucketPolicy errors are automatically handled.
		policyStr, err = core.Client.GetBucketPolicy(ctx, args.BucketName)
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
		var policyInfo = &miniogopolicy.BucketAccessPolicy{Version: "2012-10-17"}
		if policyStr != "" {
			if err = json.Unmarshal([]byte(policyStr), policyInfo); err != nil {
				return toJSONError(ctx, err, args.BucketName)
			}
		}

		policyInfo.Statements = miniogopolicy.SetPolicy(policyInfo.Statements, policyType, args.BucketName, args.Prefix)
		if len(policyInfo.Statements) == 0 {
			if err = core.SetBucketPolicy(ctx, args.BucketName, ""); err != nil {
				return toJSONError(ctx, err, args.BucketName)
			}
			return nil
		}

		bucketPolicy, err := BucketAccessPolicyToPolicy(policyInfo)
		if err != nil {
			// This should not happen.
			return toJSONError(ctx, err, args.BucketName)
		}

		policyData, err := json.Marshal(bucketPolicy)
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}

		if err = core.SetBucketPolicy(ctx, args.BucketName, string(policyData)); err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}

	} else {
		bucketPolicy, err := globalPolicySys.Get(args.BucketName)
		if err != nil {
			if _, ok := err.(BucketPolicyNotFound); !ok {
				return toJSONError(ctx, err, args.BucketName)
			}
		}
		policyInfo, err := PolicyToBucketAccessPolicy(bucketPolicy)
		if err != nil {
			// This should not happen.
			return toJSONError(ctx, err, args.BucketName)
		}

		policyInfo.Statements = miniogopolicy.SetPolicy(policyInfo.Statements, policyType, args.BucketName, args.Prefix)
		if len(policyInfo.Statements) == 0 {
			if err = globalBucketMetadataSys.Update(args.BucketName, bucketPolicyConfig, nil); err != nil {
				return toJSONError(ctx, err, args.BucketName)
			}

			return nil
		}

		bucketPolicy, err = BucketAccessPolicyToPolicy(policyInfo)
		if err != nil {
			// This should not happen.
			return toJSONError(ctx, err, args.BucketName)
		}

		configData, err := json.Marshal(bucketPolicy)
		if err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}

		// Parse validate and save bucket policy.
		if err = globalBucketMetadataSys.Update(args.BucketName, bucketPolicyConfig, configData); err != nil {
			return toJSONError(ctx, err, args.BucketName)
		}
	}

	return nil
}

// PresignedGetArgs - presigned-get API args.
type PresignedGetArgs struct {
	// Host header required for signed headers.
	HostName string `json:"host"`

	// Bucket name of the object to be presigned.
	BucketName string `json:"bucket"`

	// Object name to be presigned.
	ObjectName string `json:"object"`

	// Expiry in seconds.
	Expiry int64 `json:"expiry"`
}

// PresignedGetRep - presigned-get URL reply.
type PresignedGetRep struct {
	UIVersion string `json:"uiVersion"`
	// Presigned URL of the object.
	URL string `json:"url"`
}

// PresignedGET - returns presigned-Get url.
func (web *webAPIHandlers) PresignedGet(r *http.Request, args *PresignedGetArgs, reply *PresignedGetRep) error {
	ctx := newWebContext(r, args, "WebPresignedGet")
	claims, owner, authErr := webRequestAuthenticate(r)
	if authErr != nil {
		return toJSONError(ctx, authErr)
	}
	var creds auth.Credentials
	if !owner {
		var ok bool
		creds, ok = globalIAMSys.GetUser(claims.AccessKey)
		if !ok {
			return toJSONError(ctx, errInvalidAccessKeyID)
		}
	} else {
		creds = globalActiveCred
	}

	region := globalServerRegion
	if args.BucketName == "" || args.ObjectName == "" {
		return &json2.Error{
			Message: "Bucket and Object are mandatory arguments.",
		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(args.BucketName, false) {
		return toJSONError(ctx, errInvalidBucketName, args.BucketName)
	}

	// Check if the user indeed has GetObject access,
	// if not we do not need to generate presigned URLs
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName:     claims.AccessKey,
		Action:          iampolicy.GetObjectAction,
		BucketName:      args.BucketName,
		ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
		IsOwner:         owner,
		ObjectName:      args.ObjectName,
		Claims:          claims.Map(),
	}) {
		return toJSONError(ctx, errPresignedNotAllowed)
	}

	reply.UIVersion = Version
	reply.URL = presignedGet(args.HostName, args.BucketName, args.ObjectName, args.Expiry, creds, region)
	return nil
}

// Returns presigned url for GET method.
func presignedGet(host, bucket, object string, expiry int64, creds auth.Credentials, region string) string {
	accessKey := creds.AccessKey
	secretKey := creds.SecretKey
	sessionToken := creds.SessionToken

	date := UTCNow()
	dateStr := date.Format(iso8601Format)
	credential := fmt.Sprintf("%s/%s", accessKey, getScope(date, region))

	var expiryStr = "604800" // Default set to be expire in 7days.
	if expiry < 604800 && expiry > 0 {
		expiryStr = strconv.FormatInt(expiry, 10)
	}

	query := url.Values{}
	query.Set(xhttp.AmzAlgorithm, signV4Algorithm)
	query.Set(xhttp.AmzCredential, credential)
	query.Set(xhttp.AmzDate, dateStr)
	query.Set(xhttp.AmzExpires, expiryStr)
	query.Set(xhttp.ContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", object))
	// Set session token if available.
	if sessionToken != "" {
		query.Set(xhttp.AmzSecurityToken, sessionToken)
	}
	query.Set(xhttp.AmzSignedHeaders, "host")
	queryStr := s3utils.QueryEncode(query)

	path := SlashSeparator + path.Join(bucket, object)

	// "host" is the only header required to be signed for Presigned URLs.
	extractedSignedHeaders := make(http.Header)
	extractedSignedHeaders.Set("host", host)
	canonicalRequest := getCanonicalRequest(extractedSignedHeaders, unsignedPayload, queryStr, path, http.MethodGet)
	stringToSign := getStringToSign(canonicalRequest, date, getScope(date, region))
	signingKey := getSigningKey(secretKey, date, region, serviceS3)
	signature := getSignature(signingKey, stringToSign)

	return host + s3utils.EncodePath(path) + "?" + queryStr + "&" + xhttp.AmzSignature + "=" + signature
}

// DiscoveryDocResp - OpenID discovery document reply.
type DiscoveryDocResp struct {
	DiscoveryDoc openid.DiscoveryDoc
	UIVersion    string `json:"uiVersion"`
	ClientID     string `json:"clientId"`
}

// GetDiscoveryDoc - returns parsed value of OpenID discovery document
func (web *webAPIHandlers) GetDiscoveryDoc(r *http.Request, args *WebGenericArgs, reply *DiscoveryDocResp) error {
	if globalOpenIDConfig.DiscoveryDoc.AuthEndpoint != "" {
		reply.DiscoveryDoc = globalOpenIDConfig.DiscoveryDoc
		reply.ClientID = globalOpenIDConfig.ClientID
	}
	reply.UIVersion = Version
	return nil
}

// LoginSTSArgs - login arguments.
type LoginSTSArgs struct {
	Token string `json:"token" form:"token"`
}

var (
	errSTSNotInitialized        = errors.New("STS API not initialized, please configure STS support")
	errSTSInvalidParameterValue = errors.New("An invalid or out-of-range value was supplied for the input parameter")
)

// LoginSTS - STS user login handler.
func (web *webAPIHandlers) LoginSTS(r *http.Request, args *LoginSTSArgs, reply *LoginRep) error {
	ctx := newWebContext(r, args, "WebLoginSTS")

	if globalOpenIDValidators == nil {
		return toJSONError(ctx, errSTSNotInitialized)
	}

	v, err := globalOpenIDValidators.Get("jwt")
	if err != nil {
		logger.LogIf(ctx, err)
		return toJSONError(ctx, errSTSNotInitialized)
	}

	m, err := v.Validate(args.Token, "")
	if err != nil {
		return toJSONError(ctx, err)
	}

	var subFromToken string
	if v, ok := m[subClaim]; ok {
		subFromToken, _ = v.(string)
	}

	if subFromToken == "" {
		logger.LogIf(ctx, errors.New("STS JWT Token has `sub` claim missing, `sub` claim is mandatory"))
		return toJSONError(ctx, errSTSInvalidParameterValue)
	}

	var issFromToken string
	if v, ok := m[issClaim]; ok {
		issFromToken, _ = v.(string)
	}

	// JWT has requested a custom claim with policy value set.
	// This is a MinIO STS API specific value, this value should
	// be set and configured on your identity provider as part of
	// JWT custom claims.
	var policyName string
	policySet, ok := iampolicy.GetPoliciesFromClaims(m, iamPolicyClaimNameOpenID())
	if ok {
		policyName = globalIAMSys.CurrentPolicies(strings.Join(policySet.ToSlice(), ","))
	}
	if policyName == "" && globalPolicyOPA == nil {
		return toJSONError(ctx, fmt.Errorf("%s claim missing from the JWT token, credentials will not be generated", iamPolicyClaimNameOpenID()))
	}
	m[iamPolicyClaimNameOpenID()] = policyName

	secret := globalActiveCred.SecretKey
	cred, err := auth.GetNewCredentialsWithMetadata(m, secret)
	if err != nil {
		return toJSONError(ctx, err)
	}

	// https://openid.net/specs/openid-connect-core-1_0.html#ClaimStability
	// claim is only considered stable when subject and iss are used together
	// this is to ensure that ParentUser doesn't change and we get to use
	// parentUser as per the requirements for service accounts for OpenID
	// based logins.
	cred.ParentUser = "jwt:" + subFromToken + ":" + issFromToken

	// Set the newly generated credentials.
	if err = globalIAMSys.SetTempUser(cred.AccessKey, cred, policyName); err != nil {
		return toJSONError(ctx, err)
	}

	// Notify all other MinIO peers to reload temp users
	for _, nerr := range globalNotificationSys.LoadUser(cred.AccessKey, true) {
		if nerr.Err != nil {
			logger.GetReqInfo(ctx).SetTags("peerAddress", nerr.Host.String())
			logger.LogIf(ctx, nerr.Err)
		}
	}

	reply.Token = cred.SessionToken
	reply.UIVersion = Version
	return nil
}

// toJSONError converts regular errors into more user friendly
// and consumable error message for the browser UI.
func toJSONError(ctx context.Context, err error, params ...string) (jerr *json2.Error) {
	apiErr := toWebAPIError(ctx, err)
	jerr = &json2.Error{
		Message: apiErr.Description,
	}
	switch apiErr.Code {
	// Reserved bucket name provided.
	case "AllAccessDisabled":
		if len(params) > 0 {
			jerr = &json2.Error{
				Message: fmt.Sprintf("All access to this bucket %s has been disabled.", params[0]),
			}
		}
	// Bucket name invalid with custom error message.
	case "InvalidBucketName":
		if len(params) > 0 {
			jerr = &json2.Error{
				Message: fmt.Sprintf("Bucket Name %s is invalid. Lowercase letters, period, hyphen, numerals are the only allowed characters and should be minimum 3 characters in length.", params[0]),
			}
		}
	// Bucket not found custom error message.
	case "NoSuchBucket":
		if len(params) > 0 {
			jerr = &json2.Error{
				Message: fmt.Sprintf("The specified bucket %s does not exist.", params[0]),
			}
		}
	// Object not found custom error message.
	case "NoSuchKey":
		if len(params) > 1 {
			jerr = &json2.Error{
				Message: fmt.Sprintf("The specified key %s does not exist", params[1]),
			}
		}
		// Add more custom error messages here with more context.
	}
	return jerr
}

// toWebAPIError - convert into error into APIError.
func toWebAPIError(ctx context.Context, err error) APIError {
	switch err {
	case errNoAuthToken:
		return APIError{
			Code:           "WebTokenMissing",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    err.Error(),
		}
	case errSTSNotInitialized:
		return APIError(stsErrCodes.ToSTSErr(ErrSTSNotInitialized))
	case errServerNotInitialized:
		return APIError{
			Code:           "XMinioServerNotInitialized",
			HTTPStatusCode: http.StatusServiceUnavailable,
			Description:    err.Error(),
		}
	case errAuthentication, auth.ErrInvalidAccessKeyLength,
		auth.ErrInvalidSecretKeyLength, errInvalidAccessKeyID, errAccessDenied, errLockedObject:
		return APIError{
			Code:           "AccessDenied",
			HTTPStatusCode: http.StatusForbidden,
			Description:    err.Error(),
		}
	case errSizeUnspecified:
		return APIError{
			Code:           "InvalidRequest",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    err.Error(),
		}
	case errChangeCredNotAllowed:
		return APIError{
			Code:           "MethodNotAllowed",
			HTTPStatusCode: http.StatusMethodNotAllowed,
			Description:    err.Error(),
		}
	case errInvalidBucketName:
		return APIError{
			Code:           "InvalidBucketName",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    err.Error(),
		}
	case errInvalidArgument:
		return APIError{
			Code:           "InvalidArgument",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    err.Error(),
		}
	case errEncryptedObject:
		return getAPIError(ErrSSEEncryptedObject)
	case errInvalidEncryptionParameters:
		return getAPIError(ErrInvalidEncryptionParameters)
	case errObjectTampered:
		return getAPIError(ErrObjectTampered)
	case errMethodNotAllowed:
		return getAPIError(ErrMethodNotAllowed)
	}

	// Convert error type to api error code.
	switch err.(type) {
	case StorageFull:
		return getAPIError(ErrStorageFull)
	case BucketQuotaExceeded:
		return getAPIError(ErrAdminBucketQuotaExceeded)
	case BucketNotFound:
		return getAPIError(ErrNoSuchBucket)
	case BucketNotEmpty:
		return getAPIError(ErrBucketNotEmpty)
	case BucketExists:
		return getAPIError(ErrBucketAlreadyOwnedByYou)
	case BucketNameInvalid:
		return getAPIError(ErrInvalidBucketName)
	case hash.BadDigest:
		return getAPIError(ErrBadDigest)
	case IncompleteBody:
		return getAPIError(ErrIncompleteBody)
	case ObjectExistsAsDirectory:
		return getAPIError(ErrObjectExistsAsDirectory)
	case ObjectNotFound:
		return getAPIError(ErrNoSuchKey)
	case ObjectNameInvalid:
		return getAPIError(ErrNoSuchKey)
	case InsufficientWriteQuorum:
		return getAPIError(ErrWriteQuorum)
	case InsufficientReadQuorum:
		return getAPIError(ErrReadQuorum)
	case NotImplemented:
		return APIError{
			Code:           "NotImplemented",
			HTTPStatusCode: http.StatusBadRequest,
			Description:    "Functionality not implemented",
		}
	}

	// Log unexpected and unhandled errors.
	logger.LogIf(ctx, err)
	return toAPIError(ctx, err)
}

// writeWebErrorResponse - set HTTP status code and write error description to the body.
func writeWebErrorResponse(w http.ResponseWriter, err error) {
	reqInfo := &logger.ReqInfo{
		DeploymentID: globalDeploymentID,
	}
	ctx := logger.SetReqInfo(GlobalContext, reqInfo)
	apiErr := toWebAPIError(ctx, err)
	w.WriteHeader(apiErr.HTTPStatusCode)
	sendResponse := SendResponse{Status: FailResponseStatus, Message: apiErr.Description}
	errJson, error := json.Marshal(sendResponse)
	if error != nil {
		logs.GetLogger().Error(error)
	}
	w.Write(errJson)
}

func writeOfflineDealsErrorResponse(w http.ResponseWriter, err error) {
	reqInfo := &logger.ReqInfo{
		DeploymentID: globalDeploymentID,
	}
	ctx := logger.SetReqInfo(GlobalContext, reqInfo)
	apiErr := toWebAPIError(ctx, err)
	w.WriteHeader(apiErr.HTTPStatusCode)
	sendResponse := BucketOfflineDealResponse{Status: FailResponseStatus, Message: apiErr.Description}
	errJson, error := json.Marshal(sendResponse)
	if error != nil {
		logs.GetLogger().Error(error)
	}
	w.Write(errJson)
}

type DealVo struct {
	SwanEndpoint string `json:"swan_endpoint,omitempty"`
	CarSliceSize int64  `json:"car_slice_size,omitempty"`
	Start        uint   `json:"start,omitempty"`
	Duration     uint   `json:"duration,omitempty"`
	Price        string `json:"price,omitempty"`
	SwanApiToken string `json:"swan_api_token,omitempty"`
	MinerId      string `json:"miner_id,omitempty"`
}

type OfflineDealRequest struct {
	TaskName string `json:"task_name"`
	Start    uint   `json:"start"`
	Duration uint   `json:"duration"`
	Price    string `json:"price"`
	MinerId  string `json:"miner_id"`
}

type OnlineDealRequest struct {
	VerifiedDeal  string `json:"verifiedDeal"`
	FastRetrieval string `json:"fastRetrieval"`
	MinerId       string `json:"minerId"`
	Price         string `json:"price"`
	Duration      string `json:"duration"`
}

type OnlineDealResponse struct {
	Filename      string `json:"filename"`
	WalletAddress string `json:"walletAddress"`
	VerifiedDeal  string `json:"verifiedDeal"`
	FastRetrieval string `json:"fastRetrieval"`
	DataCid       string `json:"dataCid"`
	MinerId       string `json:"minerId"`
	Price         string `json:"price"`
	Duration      string `json:"duration"`
	DealCid       string `json:"dealCid"`
	TimeStamp     string `json:"timeStamp"`
}

func (d *DealVo) setDefault() {
	if d.CarSliceSize == 0 {
		d.CarSliceSize = 34091302912 // 32GB = 1024 *1024 *1024 *254 /256 *32
	}
	if d.Start == 0 {
		d.Start = 7 // 1 week = 7 days
	}
	if d.Duration == 0 {
		d.Duration = 365 // 1 year = 365 days
	}
	if len(d.Price) == 0 {
		d.Price = "0"
	}
	//if len(d.SwanEndpoint) == 0 {
	//d.SwanEndpoint = os.Getenv("SWAN_API")
	//if len(strings.TrimSpace(d.SwanEndpoint)) == 0 {
	//d.SwanEndpoint = "https://api.filswan.com"
	//}
	//os.Setenv("SWAN_API", d.SwanEndpoint)
	//} else {
	//os.Setenv("SWAN_API", d.SwanEndpoint)
	//}
	//if len(d.SwanApiToken) == 0 {
	//d.SwanApiToken = os.Getenv("SWAN_TOKEN")
	//} else {
	//os.Setenv("SWAN_TOKEN", d.SwanApiToken)
	//}
	if len(d.MinerId) == 0 {
		d.MinerId = "f0447183"
	}
}

func ExecCommand(strCommand string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", strCommand)
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		logs.GetLogger().Error("Execute failed when Start:" + err.Error())
		return "", err
	}
	out_bytes, _ := ioioutil.ReadAll(stdout)
	if err := stdout.Close(); err != nil {
		logs.GetLogger().Error("Execute failed when close stdout:" + err.Error())
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		logs.GetLogger().Error("Execute failed when Wait:" + err.Error())
		return "", err
	}
	return string(out_bytes), nil
}

type SendRequest struct {
	Data    OnlineDealRequest `json:"data"`
	Status  string            `json:"status"`
	Message string            `json:"message"`
}

type AuthToken struct {
	Data    string `json:"data"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type SendResponse struct {
	Data    OnlineDealResponse `json:"data"`
	Status  string             `json:"status"`
	Message string             `json:"message"`
}

func (web *webAPIHandlers) JsonRetrieveDeal(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebJsonRetrieveDeal")
	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}

	var opts ObjectOptions
	gr, err := getObjectNInfo(ctx, bucket, object, nil, r.Header, readLock, opts)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	defer gr.Close()

	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	expandedDir, err := JsonPath(bucket, object)
	file, err := ioioutil.ReadFile(expandedDir)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
	}

	data := ManifestJson{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
	}

	fileIndex := -1
	for i, v := range data.FileList {
		if v.FileName == object {
			fileIndex = i
		}
	}
	if fileIndex != -1 {
		fileDeals := data.FileList[fileIndex]
		retrieveResponse := RetrieveResponse{
			Data:    fileDeals,
			Status:  SuccessResponseStatus,
			Message: SuccessResponseStatus,
		}

		dataBytes, err := json.Marshal(retrieveResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
		}
		w.Write(dataBytes)
		return
	} else {
		blankDeals := BucketFileList{
			FileName: object,
			Deals:    []SendResponse{},
		}
		retrieveResponse := RetrieveResponse{Data: blankDeals, Status: SuccessResponseStatus, Message: "The specified object does not have deals"}
		dataBytes, err := json.Marshal(retrieveResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(dataBytes)
		return
	}
}

func (web *webAPIHandlers) RetrieveDeal(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebRetrieveDeal")
	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}

	var opts ObjectOptions
	gr, err := getObjectNInfo(ctx, bucket, object, nil, r.Header, readLock, opts)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}
	defer gr.Close()

	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	expandedDir, err := LevelDbPath()
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}
	defer db.Close()
	fileDealKey := bucket + "_" + object
	fileDeals, err := db.Get([]byte(fileDealKey), nil)
	if err != nil || fileDeals == nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	data := BucketFileList{}
	err = json.Unmarshal(fileDeals, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}

	retrieveResponse := RetrieveResponse{Data: data, Status: SuccessResponseStatus, Message: SuccessResponseStatus}
	dataBytes, err := json.Marshal(retrieveResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
	}
	w.Write(dataBytes)
	return
}

func (web *webAPIHandlers) JsonRetrieveBucketDeal(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebJsonRetrieveBucketDeal")
	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
					return
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
	}

	_, err := objectAPI.GetBucketInfo(ctx, bucket)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	expandedDir, err := BucketJsonPath()
	file, err := ioioutil.ReadFile(expandedDir)
	if err != nil {
		logs.GetLogger().Error(err)
	}

	data := BucketManifestJson{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}

	fileIndex := -1
	for i, v := range data.BucketDealsList {
		if v.BucketName == bucket {
			fileIndex = i
		}
	}
	if fileIndex != -1 {
		bucketDeals := data.BucketDealsList[fileIndex]
		retrieveResponse := RetrieveBucketResponse{
			Data:    bucketDeals,
			Status:  SuccessResponseStatus,
			Message: SuccessResponseStatus,
		}

		dataBytes, err := json.Marshal(retrieveResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(dataBytes)
		return
	} else {
		blankDeals := BucketDealList{
			BucketName: bucket,
			Deals:      []SendResponse{},
		}
		retrieveResponse := RetrieveBucketResponse{Data: blankDeals, Status: SuccessResponseStatus, Message: "The specified bucket does not have deals"}
		dataBytes, err := json.Marshal(retrieveResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(dataBytes)
		return
	}
}

func (web *webAPIHandlers) RetrieveDeals(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebRetrieveDeals")
	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
					return
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
	}

	_, err := objectAPI.GetBucketInfo(ctx, bucket)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	expandedDir, err := LevelDbPath()
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}
	defer db.Close()

	bucketDeals, err := db.Get([]byte(bucket), nil)
	if err != nil || bucketDeals == nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}
	data := BucketDealList{}
	err = json.Unmarshal(bucketDeals, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}

	retrieveResponse := RetrieveBucketResponse{Data: data, Status: SuccessResponseStatus, Message: SuccessResponseStatus}
	dataBytes, err := json.Marshal(retrieveResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	w.Write(dataBytes)
	return
}

// SendDeal - send deal to filecoin network.
func (web *webAPIHandlers) SendDeal(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebSendDeal")

	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
					return
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}
	var opts ObjectOptions
	gr, err := getObjectNInfo(ctx, bucket, object, nil, r.Header, readLock, opts)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	defer gr.Close()

	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	decoder := json.NewDecoder(r.Body)
	var onlineDealRequest OnlineDealRequest
	err = decoder.Decode(&onlineDealRequest)
	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	_ = func(address string) (string, error) {
		addr := net.ParseIP(address)
		if addr != nil {
			// Host is an ip address
			err := errors.New("for dev env, please provide a header with valid Host, exp: a5a84b78-dd4a-45f4-bd90-31428fc23a21.cygnus.nbai.io")
			return "", err
		} else {
			// Host is a host name
			domainSegments := strings.Split(address, ".")
			if len(domainSegments) > 1 {
				return domainSegments[0], nil
			} else {
				err := errors.New(fmt.Sprintf("invalid Host in header %s", address))
				return "", err
			}
		}
	}

	fs3VolumeAddress := config.Fs3VolumeAddress

	//sourceBucketPath := filepath.Join(fs3VolumeAddress, bucket)
	sourceFilePath := filepath.Join(fs3VolumeAddress, bucket, object)
	// send online deal to lotus
	filWallet := config.Fs3WalletAddress
	if filWallet == "" {
		noWalletResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noWalletResponse,
			Status:  FailResponseStatus,
			Message: "Please provide a wallet address for sending deals",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(bodyByte)
		return
	}

	verifiedDeal := "--verified-deal=" + onlineDealRequest.VerifiedDeal
	fastRetrieval := "--fast-retrieval=" + onlineDealRequest.FastRetrieval
	commandLine := "lotus " + "client " + "import " + sourceFilePath
	dataCID, err := ExecCommand(commandLine)
	if err != nil {
		noDataCidResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noDataCidResponse,
			Status:  FailResponseStatus,
			Message: "Sending deal failed during lotus importing",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(bodyByte)
		return
	}
	outStr := strings.Fields(string(dataCID))
	dataCIDStr := outStr[len(outStr)-1]
	dealCID, err := exec.Command("lotus", "client", "deal", "--from", filWallet, verifiedDeal, fastRetrieval, dataCIDStr, onlineDealRequest.MinerId, onlineDealRequest.Price, onlineDealRequest.Duration).Output()
	if err != nil {
		noDealCidResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noDealCidResponse,
			Status:  FailResponseStatus,
			Message: "Sending deal failed during lotus sending deal",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			writeWebErrorResponse(w, err)
			logs.GetLogger().Error(err)
			return
		}
		w.Write(bodyByte)
		return
	}
	dealCIDStr := string(dealCID)
	dealCIDStr = strings.TrimSuffix(dealCIDStr, "\n")

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	onlineDealResponse := OnlineDealResponse{
		Filename:      sourceFilePath,
		WalletAddress: filWallet,
		VerifiedDeal:  onlineDealRequest.VerifiedDeal,
		FastRetrieval: onlineDealRequest.FastRetrieval,
		DataCid:       dataCIDStr,
		MinerId:       onlineDealRequest.MinerId,
		Price:         onlineDealRequest.Price,
		Duration:      onlineDealRequest.Duration,
		DealCid:       dealCIDStr,
		TimeStamp:     timestamp,
	}
	sendResponse := SendResponse{
		Data:    onlineDealResponse,
		Status:  SuccessResponseStatus,
		Message: SuccessResponseStatus,
	}
	bodyByte, err := json.Marshal(sendResponse)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}
	w.Write(bodyByte)
	//SaveToJson(bucket, object, sendResponse)
	err = SaveToDb(bucket, object, sendResponse)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
	}
	return
}

func (web *webAPIHandlers) SendDeals(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebSendBucketDeal")

	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, err := json.Marshal(sendResponse)
				if err != nil {
					logs.GetLogger().Error(err)
					writeWebErrorResponse(w, err)
					return
				}
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, err := json.Marshal(sendResponse)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, err := json.Marshal(sendResponseIam)
			if err != nil {
				logs.GetLogger().Error(err)
				writeWebErrorResponse(w, err)
				return
			}
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
	}

	_, err := objectAPI.GetBucketInfo(ctx, bucket)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var onlineDealRequest OnlineDealRequest
	err = decoder.Decode(&onlineDealRequest)
	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	_ = func(address string) (string, error) {
		addr := net.ParseIP(address)
		if addr != nil {
			// Host is an ip address
			err := errors.New("for dev env, please provide a header with valid Host, exp: a5a84b78-dd4a-45f4-bd90-31428fc23a21.cygnus.nbai.io")
			return "", err
		} else {
			// Host is a host name
			domainSegments := strings.Split(address, ".")
			if len(domainSegments) > 1 {
				return domainSegments[0], nil
			} else {
				err := errors.New(fmt.Sprintf("invalid Host in header %s", address))
				return "", err
			}
		}
	}

	// send online deal to lotus
	filWallet := config.Fs3WalletAddress
	if filWallet == "" {
		noWalletResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noWalletResponse,
			Status:  FailResponseStatus,
			Message: "Please provide a wallet address for sending deals",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			writeWebErrorResponse(w, err)
			logs.GetLogger().Error(err)
			return
		}
		w.Write(bodyByte)
		return
	}
	fs3VolumeAddress := config.Fs3VolumeAddress
	sourceBucketPath := filepath.Join(fs3VolumeAddress, bucket)
	outputBucketZipPath := filepath.Join(fs3VolumeAddress, bucket+"_deals.zip")
	sourceBucketZipPath, err := ZipBucket(sourceBucketPath, outputBucketZipPath)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
		return
	}

	verifiedDeal := "--verified-deal=" + onlineDealRequest.VerifiedDeal
	fastRetrieval := "--fast-retrieval=" + onlineDealRequest.FastRetrieval
	commandLine := "lotus " + "client " + "import " + sourceBucketZipPath
	dataCID, err := ExecCommand(commandLine)
	if err != nil {
		noDataCidResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noDataCidResponse,
			Status:  FailResponseStatus,
			Message: "Sending bucket deal failed during lotus importing",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(bodyByte)
		return
	}
	outStr := strings.Fields(string(dataCID))
	dataCIDStr := outStr[len(outStr)-1]
	dealCID, err := exec.Command("lotus", "client", "deal", "--from", filWallet, verifiedDeal, fastRetrieval, dataCIDStr, onlineDealRequest.MinerId, onlineDealRequest.Price, onlineDealRequest.Duration).Output()
	if err != nil {
		noDealCidResponse := OnlineDealResponse{}
		sendResponse := SendResponse{
			Data:    noDealCidResponse,
			Status:  FailResponseStatus,
			Message: "Sending bucket deal failed during lotus sending deal",
		}
		bodyByte, err := json.Marshal(sendResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		w.Write(bodyByte)
		return
	}
	dealCIDStr := string(dealCID)
	dealCIDStr = strings.TrimSuffix(dealCIDStr, "\n")

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()/1000, 10)

	onlineDealResponse := OnlineDealResponse{
		Filename:      sourceBucketZipPath,
		WalletAddress: filWallet,
		VerifiedDeal:  onlineDealRequest.VerifiedDeal,
		FastRetrieval: onlineDealRequest.FastRetrieval,
		DataCid:       dataCIDStr,
		MinerId:       onlineDealRequest.MinerId,
		Price:         onlineDealRequest.Price,
		Duration:      onlineDealRequest.Duration,
		DealCid:       dealCIDStr,
		TimeStamp:     timestamp,
	}
	sendResponse := SendResponse{
		Data:    onlineDealResponse,
		Status:  SuccessResponseStatus,
		Message: SuccessResponseStatus,
	}
	bodyByte, err := json.Marshal(sendResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	w.Write(bodyByte)
	//BucketSaveToJson(bucket, sendResponse)
	err = BucketSaveToDb(bucket, sendResponse)
	if err != nil {
		writeWebErrorResponse(w, err)
		logs.GetLogger().Error(err)
	}
	return
}

func JsonPath(bucket string, object string) (string, error) {

	fs3VolumeAddress := config.Fs3VolumeAddress
	bucketJson := "." + bucket + ".json"
	bucketJsonPath := filepath.Join(fs3VolumeAddress, bucketJson)
	expandedDir, err := oshomedir.Expand(bucketJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedDir, nil
}

func BucketJsonPath() (string, error) {

	fs3VolumeAddress := config.Fs3VolumeAddress
	bucketJson := "." + "bucketdeals" + ".json"
	bucketJsonPath := filepath.Join(fs3VolumeAddress, bucketJson)
	expandedDir, err := oshomedir.Expand(bucketJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedDir, nil
}

func LevelDbPath() (string, error) {
	fs3VolumeAddress := config.Fs3VolumeAddress
	levelDbName := ".leveldb.db"
	levelDbPath := filepath.Join(fs3VolumeAddress, levelDbName)
	expandedDir, err := oshomedir.Expand(levelDbPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedDir, nil
}

func BucketZipPath(outputBucketZipPath string) (string, error) {
	expandedDir, err := oshomedir.Expand(outputBucketZipPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return expandedDir, nil
}

type RetrieveResponse struct {
	Data    BucketFileList `json:"data"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
}

type RetrieveBucketResponse struct {
	Data    BucketDealList `json:"data"`
	Status  string         `json:"status"`
	Message string         `json:"message"`
}

type ManifestJson struct {
	BucketName string           `json:"bucket_name"`
	FileList   []BucketFileList `json:"file_list"`
}

type BucketFileList struct {
	FileName string         `json:"file_name"`
	Deals    []SendResponse `json:"deals"`
}

type TaskResponse struct {
	FileName string `json:"filename"`
	Uuid     string `json:"uuid"`
}

type CreateTaskResponse struct {
	Data    TaskResponse `json:"data"`
	Status  string       `json:"status"`
	Message string       `json:"message"`
}

type BucketInfoResponse struct {
	BucketName string             `json:"bucket_name"`
	Deals      CreateTaskResponse `json:"deals"`
}

type BucketOfflineDealResponse struct {
	Data    BucketInfoResponse `json:"data"`
	Status  string             `json:"status"`
	Message string             `json:"message"`
}

func SaveToJson(bucket string, object string, response SendResponse) error {
	expandedDir, _ := JsonPath(bucket, object)
	_, err := os.OpenFile(expandedDir, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	file, err := ioutil.ReadFile(expandedDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	data := ManifestJson{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	dealBucketName := bucket
	dealFileName := object
	if data.BucketName == "" {
		newDeals := []SendResponse{}
		newDeals = append(newDeals, response)
		newBucketFileList := BucketFileList{
			FileName: dealFileName,
			Deals:    newDeals,
		}
		newFileList := []BucketFileList{}
		newFileList = append(newFileList, newBucketFileList)
		newManifestJson := ManifestJson{
			BucketName: dealBucketName,
			FileList:   newFileList,
		}
		dataBytes, err := json.Marshal(newManifestJson)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	} else {
		fileIndex := -1
		for i, v := range data.FileList {
			if v.FileName == object {
				fileIndex = i
			}
		}
		if fileIndex != -1 {
			data.FileList[fileIndex].Deals = append(data.FileList[fileIndex].Deals, response)
			dataBytes, err := json.Marshal(data)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
		} else {
			newDeal := []SendResponse{}
			newDeal = append(newDeal, response)
			newBucketFileList := BucketFileList{
				FileName: dealFileName,
				Deals:    newDeal,
			}
			data.FileList = append(data.FileList, newBucketFileList)
			dataBytes, err := json.Marshal(data)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
		}

	}

	return err
}

type DealResponseVo struct {
	Status string         `json:"status"`
	Data   *DealRequestVo `json:"data"`
}

type DealRequestVo struct {
	PieceCid     string `json:"piece_cid"`
	PieceSize    uint64 `json:"piece_size"`
	DataCid      string `json:"data_cid"`
	MinerId      string `json:"miner_id"`
	Duration     uint   `json:"duration"`
	Price        string `json:"price,omitempty"`
	Start        uint   `json:"start,omitempty"`
	SenderWallet string `json:"sender_wallet,omitempty"`
	StartEpoch   uint   `json:"start_epoch,omitempty"`
	VerifiedDeal bool   `json:"verified_deal,omitempty"`
	DealCost     string `json:"deal_cost,omitempty"`
	TotalCost    string `json:"total_cost,omitempty"`
	DealCid      string `json:"deal_cid,omitempty"`
}

func ZipBucket(sourceBucketPath string, outputBucketZipPath string) (string, error) {
	baseFolder := sourceBucketPath

	// Get a Buffer to Write To
	expandedDir, err := BucketZipPath(outputBucketZipPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	outFile, err := os.OpenFile(expandedDir, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0660)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	addFiles(w, baseFolder, "")

	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		logs.GetLogger().Error(err)
	}
	return outputBucketZipPath, err
}

func addFiles(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	expandedDir, _ := BucketZipPath(basePath)
	files, err := ioioutil.ReadDir(expandedDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			dat, err := ioioutil.ReadFile(expandedDir + "/" + file.Name())
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}

			// Add some files to the archive.
			f, err := w.Create(baseInZip + file.Name())
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			_, err = f.Write(dat)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
		} else if file.IsDir() {
			// Recurse
			newBase := basePath + file.Name() + "/"
			addFiles(w, newBase, baseInZip+file.Name()+"/")
		}
	}
}

type BucketManifestJson struct {
	VolumeAddress   string           `json:"volume_address"`
	BucketDealsList []BucketDealList `json:"bucket_deal_list"`
}

type BucketDealList struct {
	BucketName string         `json:"bucket_name"`
	Deals      []SendResponse `json:"deals"`
}

func BucketSaveToJson(bucket string, response SendResponse) error {
	expandedDir, _ := BucketJsonPath()
	_, err := os.OpenFile(expandedDir, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	file, err := ioioutil.ReadFile(expandedDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	data := BucketManifestJson{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	VolumeAddress := config.Fs3VolumeAddress
	if data.VolumeAddress == "" {
		newDeals := []SendResponse{}
		newDeals = append(newDeals, response)
		newBucketDealList := BucketDealList{
			BucketName: bucket,
			Deals:      newDeals,
		}
		newBucketDealsList := []BucketDealList{}
		newBucketDealsList = append(newBucketDealsList, newBucketDealList)
		newBucketManifestJson := BucketManifestJson{
			VolumeAddress:   VolumeAddress,
			BucketDealsList: newBucketDealsList,
		}
		dataBytes, err := json.Marshal(newBucketManifestJson)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	} else {
		fileIndex := -1
		for i, v := range data.BucketDealsList {
			if v.BucketName == bucket {
				fileIndex = i
			}
		}
		if fileIndex != -1 {
			data.BucketDealsList[fileIndex].Deals = append(data.BucketDealsList[fileIndex].Deals, response)
			dataBytes, err := json.Marshal(data)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
		} else {
			newDeal := []SendResponse{}
			newDeal = append(newDeal, response)
			newBucketDealList := BucketDealList{
				BucketName: bucket,
				Deals:      newDeal,
			}
			data.BucketDealsList = append(data.BucketDealsList, newBucketDealList)
			dataBytes, err := json.Marshal(data)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			err = ioioutil.WriteFile(expandedDir, dataBytes, 0644)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
		}

	}
	return err
}

func SaveToDb(bucket string, object string, response SendResponse) error {
	expandedDir, _ := LevelDbPath()
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	fileDealKey := bucket + "_" + object
	bucketDeals, err := db.Get([]byte(fileDealKey), nil)
	if err == nil {
		data := BucketDealList{}
		err = json.Unmarshal(bucketDeals, &data)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		data.Deals = append(data.Deals, response)
		dataBytes, err := json.Marshal(data)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = db.Put([]byte(fileDealKey), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return err
	} else {
		newDeals := []SendResponse{}
		newDeals = append(newDeals, response)
		newBucketDealList := BucketFileList{
			FileName: object,
			Deals:    newDeals,
		}
		dataBytes, err := json.Marshal(newBucketDealList)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = db.Put([]byte(fileDealKey), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return err
	}
}

func BucketSaveToDb(bucket string, response SendResponse) error {
	expandedDir, err := LevelDbPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	bucketDeals, err := db.Get([]byte(bucket), nil)
	if err == nil {
		data := BucketDealList{}
		err = json.Unmarshal(bucketDeals, &data)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		data.Deals = append(data.Deals, response)
		dataBytes, err := json.Marshal(data)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = db.Put([]byte(bucket), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return err
	} else {
		newDeals := []SendResponse{}
		newDeals = append(newDeals, response)
		newBucketDealList := BucketDealList{
			BucketName: bucket,
			Deals:      newDeals,
		}
		dataBytes, err := json.Marshal(newBucketDealList)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		err = db.Put([]byte(bucket), []byte(dataBytes), nil)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return err
	}
}

func (web *webAPIHandlers) SendOfflineDeal(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebSendOfflineDeal")

	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeWebErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	object, err := unescapePath(vars["object"])
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, _ := json.Marshal(sendResponse)
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      object,
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, _ := json.Marshal(sendResponse)
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, _ := json.Marshal(sendResponseIam)
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      object,
			Claims:          claims.Map(),
		}) {

		}
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeWebErrorResponse(w, errInvalidBucketName)
		return
	}

	getObjectNInfo := objectAPI.GetObjectNInfo
	if web.CacheAPI() != nil {
		getObjectNInfo = web.CacheAPI().GetObjectNInfo
	}
	var opts ObjectOptions
	gr, err := getObjectNInfo(ctx, bucket, object, nil, r.Header, readLock, opts)
	if err != nil {
		writeWebErrorResponse(w, err)
		return
	}
	defer gr.Close()

	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	decoder := json.NewDecoder(r.Body)
	fmt.Println(r.Body, decoder)

	var offlineDealRequest OfflineDealRequest
	err = decoder.Decode(&offlineDealRequest)

	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	_ = func(address string) (string, error) {
		addr := net.ParseIP(address)
		if addr != nil {
			// Host is an ip address
			err := errors.New("for dev env, please provide a header with valid Host, exp: a5a84b78-dd4a-45f4-bd90-31428fc23a21.cygnus.nbai.io")
			return "", err
		} else {
			// Host is a host name
			domainSegments := strings.Split(address, ".")
			if len(domainSegments) > 1 {
				return domainSegments[0], nil
			} else {
				err := errors.New(fmt.Sprintf("invalid Host in header %s", address))
				return "", err
			}
		}
	}

	// generate car
	VolumeAddress := config.Fs3VolumeAddress

	sourceFilePath := filepath.Join(VolumeAddress, bucket, object)
	carFileDir := "." + bucket
	carFileDirPath := filepath.Join(VolumeAddress, carFileDir)

	carDirExpand, err := oshomedir.Expand(carFileDirPath)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	if _, err := os.Stat(carDirExpand); os.IsNotExist(err) {
		err := os.Mkdir(carDirExpand, 0775)
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
	}

	sourceDirExpand, err := oshomedir.Expand(sourceFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}

	sliceSize, err := strconv.ParseInt(config.CarFileSize, 10, 64)
	carDir := carDirExpand
	parentPath := sourceDirExpand
	targetPath := sourceDirExpand
	graphName := object
	parallel := 4

	Emptyctx := context.Background()
	var cb graphsplit.GraphBuildCallback

	cb = graphsplit.CommPCallback(carDir)
	err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}
	// send deal request to swan
	offlineDeals, err := readCsv(filepath.Join(carDir, "car.csv"))
	logger.LogIf(Emptyctx, err)
	if err != nil {
		logs.GetLogger().Error(err)
		writeWebErrorResponse(w, err)
		return
	}

	// todo send multiple deals
	if len(offlineDeals) == 1 {
		if err != nil {
			logs.GetLogger().Error(err)
			writeWebErrorResponse(w, err)
			return
		}
		bodyByte, _ := json.Marshal(offlineDeals[0])
		w.Write(bodyByte)
	}
	return

}

func (web *webAPIHandlers) SendOfflineDeals(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "WebSendOfflineDeals")

	claims, owner, authErr := webRequestAuthenticate(r)
	defer logger.AuditLog(ctx, w, r, claims.Map())

	objectAPI := web.ObjectAPI()
	if objectAPI == nil {
		writeOfflineDealsErrorResponse(w, errServerNotInitialized)
		return
	}

	vars := mux.Vars(r)

	bucket := vars["bucket"]

	if authErr != nil {
		if authErr == errNoAuthToken {
			// Check if anonymous (non-owner) has access to download objects.
			if !globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {
				w.WriteHeader(http.StatusUnauthorized)
				sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, FS3 token missing"}
				errJson, _ := json.Marshal(sendResponse)
				w.Write(errJson)
				return
			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectRetentionAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
			if globalPolicySys.IsAllowed(policy.Args{
				Action:          policy.GetObjectLegalHoldAction,
				BucketName:      bucket,
				ConditionValues: getConditionValues(r, "", "", nil),
				IsOwner:         false,
				ObjectName:      "",
			}) {

			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponse := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJson, _ := json.Marshal(sendResponse)
			w.Write(errJson)
			return
		}
	}

	// For authenticated users apply IAM policy.
	if authErr == nil {
		if !globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {
			w.WriteHeader(http.StatusUnauthorized)
			sendResponseIam := AuthToken{Status: FailResponseStatus, Message: "Authentication failed, check your FS3 token"}
			errJsonIam, _ := json.Marshal(sendResponseIam)
			w.Write(errJsonIam)
			return
		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectRetentionAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
		if globalIAMSys.IsAllowed(iampolicy.Args{
			AccountName:     claims.AccessKey,
			Action:          iampolicy.GetObjectLegalHoldAction,
			BucketName:      bucket,
			ConditionValues: getConditionValues(r, "", claims.AccessKey, claims.Map()),
			IsOwner:         owner,
			ObjectName:      "",
			Claims:          claims.Map(),
		}) {

		}
	}

	_, err := objectAPI.GetBucketInfo(ctx, bucket)
	if err != nil {
		writeOfflineDealsErrorResponse(w, err)
		return
	}

	// Check if bucket is a reserved bucket name or invalid.
	if isReservedOrInvalidBucket(bucket, false) {
		writeOfflineDealsErrorResponse(w, errInvalidBucketName)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var offlineDealRequest TaskInfo
	err = decoder.Decode(&offlineDealRequest)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	if err != nil && err != io.EOF {
		w.Write([]byte(fmt.Sprintf("bad request: %s", err.Error())))
		return
	}

	_ = func(address string) (string, error) {
		addr := net.ParseIP(address)
		if addr != nil {
			// Host is an ip address
			err := errors.New("for dev env, please provide a header with valid Host, exp: a5a84b78-dd4a-45f4-bd90-31428fc23a21.cygnus.nbai.io")
			return "", err
		} else {
			// Host is a host name
			domainSegments := strings.Split(address, ".")
			if len(domainSegments) > 1 {
				return domainSegments[0], nil
			} else {
				err := errors.New(fmt.Sprintf("invalid Host in header %s", address))
				return "", err
			}
		}
	}

	// generate car
	VolumeAddress := config.Fs3VolumeAddress

	sourceDirPath := filepath.Join(VolumeAddress, bucket)
	carFileDir := "." + bucket + "_deals"
	carFileDirPath := filepath.Join(VolumeAddress, carFileDir)

	carDirExpand, err := oshomedir.Expand(carFileDirPath)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	if _, err := os.Stat(carDirExpand); os.IsNotExist(err) {
		err := os.Mkdir(carDirExpand, 0775)
		if err != nil {
			logs.GetLogger().Error(err)
			writeOfflineDealsErrorResponse(w, err)
			return
		}
	}

	sourceDirExpand, err := oshomedir.Expand(sourceDirPath)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}

	//check if sourceDir is empty
	dirEmpty, err := IsDirEmpty(sourceDirExpand)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	if dirEmpty {
		noFileResponse := BucketInfoResponse{}
		bucketOfflineDealResponse := BucketOfflineDealResponse{Data: noFileResponse, Status: FailResponseStatus, Message: NoFileInBucket}
		dataBytes, err := json.Marshal(bucketOfflineDealResponse)
		if err != nil {
			logs.GetLogger().Error(err)
			writeOfflineDealsErrorResponse(w, err)
			return
		}
		w.Write(dataBytes)
		return
	}

	sliceSize, err := strconv.ParseInt(config.CarFileSize, 10, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	carDir := carDirExpand
	parentPath := sourceDirExpand
	targetPath := sourceDirExpand
	graphName := bucket
	parallel := 4

	Emptyctx := context.Background()
	var cb graphsplit.GraphBuildCallback

	cb = graphsplit.CommPCallback(carDir)
	err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}

	// generate car.csv
	err = saveCarCsvToDb(carDir, parentPath, bucket)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}

	//Upload to ipfs
	err = uploadCarFileAndSaveToDb(carDir, graphName)
	logger.LogIf(Emptyctx, err)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}

	//Create task on swan
	offlineDeals, err := readCsvInDb(bucket)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	reply, err := createTask(bucket, offlineDeals, carDir, offlineDealRequest)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	var createTaskResponse CreateTaskResponse
	json.Unmarshal(reply, &createTaskResponse)

	bucketInfoResponse := BucketInfoResponse{BucketName: bucket, Deals: createTaskResponse}
	bucketOfflineDealResponse := BucketOfflineDealResponse{Data: bucketInfoResponse, Status: SuccessResponseStatus, Message: SuccessResponseStatus}
	dataBytes, err := json.Marshal(bucketOfflineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		writeOfflineDealsErrorResponse(w, err)
		return
	}
	w.Write(dataBytes)
	return

}

type OfflineDeal struct {
	MinerId       string
	PieceCid      string
	PieceSize     string
	DataCid       string
	Duration      string
	Start         string
	FastRetrieval bool
	DealCid       string
	Filename      string
	Price         string
}

func NewOfflineDeal() *OfflineDeal {
	return &OfflineDeal{FastRetrieval: true}
}

type CarRecord struct {
	CarFileName    string
	CarFilePath    string
	PieceCid       string
	DataCid        string
	CarFileSize    string
	CarFileMd5     string
	SourceFileName string
	SourceFilePath string
	SourceFileSize string
	SourceFileMd5  string
}

type SourceFiles struct {
	Bucket           string
	SourceFilesNames []SourceFile
}
type SourceFile struct {
	Name string
}

type SourceFilesPath struct {
	Bucket           string
	SourceFilesPaths []SourceFilePath
}
type SourceFilePath struct {
	Path string
}

type SourceFilesSize struct {
	Bucket           string
	SourceFilesSizes []SourceFileSize
}
type SourceFileSize struct {
	Size string
}

type SourceFilesMd5 struct {
	Bucket          string
	SourceFilesMd5s []SourceFileMd5
}
type SourceFileMd5 struct {
	Md5 string
}

func uploadCarFileAndSaveToDb(carDir string, graphName string) error {
	expandedDir, err := LevelDbPath()
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	carCsv, err := db.Get([]byte(graphName+"_deals_car_csv"), nil)
	if err != nil || carCsv == nil {
		logs.GetLogger().Error(err)
		return err
	}
	csvRecord := CarCsv{}
	err = json.Unmarshal(carCsv, &csvRecord)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	carHash, err := uploadCarFileIpfs(csvRecord.CarFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	carFileAddress := config.IpfsAddress + "/ipfs/" + carHash
	csvRecord.CarFileUrl = carFileAddress

	dataBytes, err := json.Marshal(csvRecord)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(graphName+"_deals_car_csv"), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func uploadCarFile(carDir string, graphName string) error {
	records, err := readCsv(filepath.Join(carDir, "car.csv"))
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	var newRecords [][]string
	newRecords = append(newRecords, []string{"car_file_name", "car_file_path", "piece_cid", "data_cid", "car_file_size", "car_file_md5", "source_file_name", "source_file_path", "source_file_size", "source_file_md5", "car_file_url"})
	for _, record := range records {
		carHash, err := uploadCarFileIpfs(record[1])
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		carFileAddress := config.IpfsAddress + "/ipfs/" + carHash
		record = append(record, carFileAddress)
		newRecords = append(newRecords, record)
	}
	err = writeCarCsv(filepath.Join(carDir, "car.csv"), newRecords)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

type UploadIpfsResponse struct {
	Name string
	Hash string
	Size string
}

func uploadCarFileIpfs(carFilePath string) (string, error) {
	file, err := os.Open(carFilePath)

	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("", filepath.Base(file.Name()))
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	io.Copy(part, file)
	writer.Close()

	url := config.IpfsAddress + "/api/v0/add"
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	defer response.Body.Close()

	content, err := ioioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	uploadResponse := UploadIpfsResponse{}
	err = json.Unmarshal(content, &uploadResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return uploadResponse.Hash, err
}

type CarCsv struct {
	CarFileName    string
	CarFilePath    string
	PieceCid       string
	DataCid        string
	CarFileSize    string
	CarFileMd5     string
	SourceFileName string
	SourceFilePath string
	SourceFileSize string
	SourceFileMd5  string
	CarFileUrl     string
}

func saveCarCsvToDb(carDir string, parentPath string, bucket string) error {
	manifestPath := filepath.Join(carDir, "manifest.csv")
	manifestCSV, err := os.Open(manifestPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer manifestCSV.Close()

	reader := csv.NewReader(manifestCSV)
	reader.LazyQuotes = true
	reader.Comma = ','

	//ignore values in detail
	reader.FieldsPerRecord = -1

	csvLines, err := reader.ReadAll()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	manifestRecord := csvLines[len(csvLines)-1]
	carFileName := manifestRecord[0] + ".car"
	carFilePath := filepath.Join(carDir, carFileName)
	carFile, err := os.Open(carFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer carFile.Close()

	stat, err := carFile.Stat()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	carFileSize := strconv.FormatInt(stat.Size(), 10)

	h := md5.New()
	if _, err := io.Copy(h, carFile); err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	carFileMd5 := hex.EncodeToString(h.Sum(nil))

	files, err := ioioutil.ReadDir(parentPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	var sourceFiles SourceFiles
	var sourceFilesPath SourceFilesPath
	var sourceFilesSize SourceFilesSize
	var sourceFilesMd5 SourceFilesMd5
	for _, f := range files {
		sourceFile := SourceFile{
			Name: f.Name(),
		}
		sourceFilePath := SourceFilePath{
			Path: filepath.Join(parentPath, f.Name()),
		}
		file, err := os.Open(filepath.Join(parentPath, f.Name()))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)
		sourceFileSize := SourceFileSize{
			Size: fileSize,
		}

		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fileMd5 := hex.EncodeToString(h.Sum(nil))
		sourceFileMd5 := SourceFileMd5{
			Md5: fileMd5,
		}
		sourceFiles.SourceFilesNames = append(sourceFiles.SourceFilesNames, sourceFile)
		sourceFilesPath.SourceFilesPaths = append(sourceFilesPath.SourceFilesPaths, sourceFilePath)
		sourceFilesSize.SourceFilesSizes = append(sourceFilesSize.SourceFilesSizes, sourceFileSize)
		sourceFilesMd5.SourceFilesMd5s = append(sourceFilesMd5.SourceFilesMd5s, sourceFileMd5)
	}

	sourceFileName, err := json.Marshal(sourceFiles.SourceFilesNames)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFilesName := string(sourceFileName)

	sourceFilePaths, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFilePath := string(sourceFilePaths)

	sourceFileSizes, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFileSize := string(sourceFileSizes)

	sourceFileMd5s, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFileMd5 := string(sourceFileMd5s)

	expandedDir, err := LevelDbPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	newCarCsv := CarCsv{
		CarFileName:    carFileName,
		CarFilePath:    carFilePath,
		PieceCid:       manifestRecord[2],
		DataCid:        manifestRecord[0],
		CarFileSize:    carFileSize,
		CarFileMd5:     carFileMd5,
		SourceFileName: sourceFilesName,
		SourceFilePath: sourceFilePath,
		SourceFileSize: sourceFileSize,
		SourceFileMd5:  sourceFileMd5,
		CarFileUrl:     "",
	}
	dataBytes, err := json.Marshal(newCarCsv)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(bucket+"_deals_car_csv"), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func generateCarCsv(carDir string, parentPath string) error {
	manifestPath := filepath.Join(carDir, "manifest.csv")
	carPath := filepath.Join(carDir, "car.csv")
	manifestCSV, err := os.Open(manifestPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer manifestCSV.Close()

	if _, err := os.Stat(carPath); err == nil {
		os.Remove(carPath)
	}

	carCSV, err := os.OpenFile(carPath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer carCSV.Close()

	reader := csv.NewReader(manifestCSV)
	reader.LazyQuotes = true
	reader.Comma = ','

	//ignore values in detail
	reader.FieldsPerRecord = -1

	csvLines, err := reader.ReadAll()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	manifestRecord := csvLines[len(csvLines)-1]
	carFileName := manifestRecord[0] + ".car"
	carFilePath := filepath.Join(carDir, carFileName)
	carFile, err := os.Open(carFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer carFile.Close()

	stat, err := carFile.Stat()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	carFileSize := strconv.FormatInt(stat.Size(), 10)

	h := md5.New()
	if _, err := io.Copy(h, carFile); err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	carFileMd5 := hex.EncodeToString(h.Sum(nil))

	files, err := ioioutil.ReadDir(parentPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	var sourceFiles SourceFiles
	var sourceFilesPath SourceFilesPath
	var sourceFilesSize SourceFilesSize
	var sourceFilesMd5 SourceFilesMd5
	for _, f := range files {
		sourceFile := SourceFile{
			Name: f.Name(),
		}
		sourceFilePath := SourceFilePath{
			Path: filepath.Join(parentPath, f.Name()),
		}
		file, err := os.Open(filepath.Join(parentPath, f.Name()))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)
		sourceFileSize := SourceFileSize{
			Size: fileSize,
		}

		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fileMd5 := hex.EncodeToString(h.Sum(nil))
		sourceFileMd5 := SourceFileMd5{
			Md5: fileMd5,
		}
		sourceFiles.SourceFilesNames = append(sourceFiles.SourceFilesNames, sourceFile)
		sourceFilesPath.SourceFilesPaths = append(sourceFilesPath.SourceFilesPaths, sourceFilePath)
		sourceFilesSize.SourceFilesSizes = append(sourceFilesSize.SourceFilesSizes, sourceFileSize)
		sourceFilesMd5.SourceFilesMd5s = append(sourceFilesMd5.SourceFilesMd5s, sourceFileMd5)
	}

	sourceFileName, err := json.Marshal(sourceFiles.SourceFilesNames)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFilesName := string(sourceFileName)

	sourceFilePaths, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFilePath := string(sourceFilePaths)

	sourceFileSizes, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFileSize := string(sourceFileSizes)

	sourceFileMd5s, err := json.Marshal(sourceFilesPath.SourceFilesPaths)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	sourceFileMd5 := string(sourceFileMd5s)

	w := csv.NewWriter(carCSV)
	err = w.Write([]string{"car_file_name", "car_file_path", "piece_cid", "data_cid", "car_file_size", "car_file_md5", "source_file_name", "source_file_path", "source_file_size", "source_file_md5"})
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	err = w.Write([]string{carFileName, carFilePath, manifestRecord[2], manifestRecord[0], carFileSize, carFileMd5, sourceFilesName, sourceFilePath, sourceFileSize, sourceFileMd5})
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	w.Flush()
	err = w.Error()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func readCsvInDb(bucket string) (CarCsv, error) {
	expandedDir, err := LevelDbPath()
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return CarCsv{}, err
	}
	defer db.Close()

	carCsv, err := db.Get([]byte(bucket+"_deals_car_csv"), nil)
	if err != nil || carCsv == nil {
		logs.GetLogger().Error(err)
		return CarCsv{}, err
	}
	csvRecord := CarCsv{}
	err = json.Unmarshal(carCsv, &csvRecord)
	if err != nil {
		logs.GetLogger().Error(err)
		return CarCsv{}, err
	}
	return csvRecord, err
}

func readCsv(_filepath string) ([][]string, error) {
	csvFile, err := os.Open(_filepath)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.LazyQuotes = true
	reader.Comma = ','

	//ignore values in detail
	reader.FieldsPerRecord = -1
	csvLines, err := reader.ReadAll()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	csvRecord := csvLines[len(csvLines)-1]
	csvColumeLen := len(csvLines[0])
	var records [][]string
	i := 0
	for i < len(csvRecord) {
		record := csvRecord[i : i+csvColumeLen]
		records = append(records, record)
		i = i + csvColumeLen
	}
	return records, err
}

func writeCarCsv(_filepath string, records [][]string) error {
	csvFile, err := os.OpenFile(_filepath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	err = w.WriteAll(records)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	w.Flush()
	err = w.Error()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func createTask(bucket string, offlineDeals CarCsv, outputDir string, request TaskInfo) ([]byte, error) {
	taskUuid := uuid.New().String()
	err := generateMetadataCsvToDb(bucket, offlineDeals, taskUuid, outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = generateTaskCsvToDb(bucket, offlineDeals, outputDir, taskUuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = generateTaskCsv(offlineDeals, outputDir, taskUuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	response, err := createSwanTask(outputDir, taskUuid, request)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return response, err

}

type TaskInfo struct {
	TaskName       string `json:"task_name"`
	CuratedDataset string `json:"curated_dataset"`
	Description    string `json:"description"`
	IsPublic       string `json:"is_public"`
	Type           string `json:"type"`
	MinerId        string `json:"miner_id"`
	MinPrice       string `json:"min_price"`
	MaxPrice       string `json:"max_price"`
	Tags           string `json:"tags"`
	ExpireDays     string `json:"expire_days"`
}

func createSwanTask(outputDir string, taskUuid string, request TaskInfo) ([]byte, error) {
	taskCsv, err := os.Open(path.Join(outputDir, taskUuid+".csv"))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	client := http.Client{Timeout: time.Minute}
	method := "POST"
	swanUrl := config.SwanAddress + "/tasks"
	token := config.SwanToken

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormField("task_name")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.TaskName))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("curated_dataset")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.CuratedDataset))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("description")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.Description))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("is_public")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if request.IsPublic == "0" {
	}
	_, err = io.Copy(fw, strings.NewReader(request.IsPublic))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("verified_type")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.Type))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("miner_id")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.MinerId))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("min_price")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.MinPrice))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("max_price")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.MaxPrice))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("tags")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.Tags))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormField("expired_days")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(request.ExpireDays))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fw, err = writer.CreateFormFile("file", taskUuid+".csv")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	_, err = io.Copy(fw, taskCsv)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest(method, swanUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := client.Do(req)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := ioioutil.ReadAll(res.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return bodyBytes, err
}

type TaskCsv struct {
	Uuid          string
	MinerId       string
	DealCid       string
	PayloadCid    string
	FileSourceUrl string
	Md5           string
	StartEpoch    string
}

func generateTaskCsvToDb(bucket string, records CarCsv, outputDir string, taskUuid string) error {
	expandedDir, err := LevelDbPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	newRecord := TaskCsv{
		Uuid:          taskUuid,
		MinerId:       "",
		DealCid:       "",
		PayloadCid:    records.DataCid,
		FileSourceUrl: records.CarFileUrl,
		Md5:           records.CarFileMd5,
		StartEpoch:    "",
	}
	dataBytes, err := json.Marshal(newRecord)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(bucket+"_deals_task_csv"), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func generateTaskCsv(records CarCsv, outputDir string, taskUuid string) error {
	taskCsvPath := filepath.Join(outputDir, taskUuid+".csv")
	csvFile, err := os.OpenFile(taskCsvPath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer csvFile.Close()

	var taskRecords [][]string
	taskRecords = append(taskRecords, []string{"uuid", "miner_id", "deal_cid", "payload_cid", "file_source_url", "md5", "start_epoch"})
	taskRecord := []string{taskUuid, "", "", records.DataCid, records.CarFileUrl, records.CarFileMd5, ""}
	taskRecords = append(taskRecords, taskRecord)

	w := csv.NewWriter(csvFile)
	err = w.WriteAll(taskRecords)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	w.Flush()
	err = w.Error()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

type MetaDataCsv struct {
	CarFileName    string
	CarFilePath    string
	PieceCid       string
	DataCid        string
	CarFileSize    string
	CarFileMd5     string
	SourceFileName string
	SourceFilePath string
	SourceFileSize string
	SourceFileMd5  string
	CarFileUrl     string
	Uuid           string
	SourceFileUrl  string
	DealCid        string
	MinerId        string
	StartEpoch     string
}

func generateMetadataCsvToDb(bucket string, records CarCsv, taskUuid string, outputDir string) error {
	expandedDir, err := LevelDbPath()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	db, err := leveldb.OpenFile(expandedDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer db.Close()

	newRecord := MetaDataCsv{
		CarFileName:    records.CarFileName,
		CarFilePath:    records.CarFilePath,
		PieceCid:       records.PieceCid,
		DataCid:        records.DataCid,
		CarFileSize:    records.DataCid,
		CarFileMd5:     records.CarFileMd5,
		SourceFileName: records.SourceFileName,
		SourceFilePath: records.SourceFilePath,
		SourceFileSize: records.SourceFileSize,
		SourceFileMd5:  records.SourceFileMd5,
		CarFileUrl:     records.CarFileUrl,
		Uuid:           taskUuid,
		SourceFileUrl:  "",
		DealCid:        "",
		MinerId:        "",
		StartEpoch:     "",
	}
	dataBytes, err := json.Marshal(newRecord)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	err = db.Put([]byte(bucket+"_deals_metadata_csv"), []byte(dataBytes), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func generateMetadataCsv(records [][]string, taskUuid string, outputDir string) error {
	var newRecords [][]string
	newRecords = append(newRecords, []string{"car_file_name", "car_file_path", "piece_cid", "data_cid", "car_file_size", "car_file_md5", "source_file_name", "source_file_path", "source_file_size", "source_file_md5", "car_file_url", "uuid", "source_file_url", "deal_cid", "miner_id", "start_epoch"})
	for _, record := range records {
		record = append(record, taskUuid, "", "", "", "")
		newRecords = append(newRecords, record)
	}

	metaCsvPath := filepath.Join(outputDir, taskUuid+"-metadata.csv")
	csvFile, err := os.OpenFile(metaCsvPath, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	err = w.WriteAll(newRecords)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	w.Flush()
	err = w.Error()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return err
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
