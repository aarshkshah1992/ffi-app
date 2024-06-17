This document aims to create a detailed work plan for shipping the `pre-built-ffi` workstream.
The high level approach for this has already been well documented at https://hackmd.io/@mvdan/Hy7iK0TEY.

### Changes to https://github.com/filecoin-project/filecoin-ffi

Going ahead, except for cases where users want to build `filecoin-ffi` from source, `filecoin-ffi` will essentially act as a "wrapper" that forwards all API calls to the corresponding `prebuilt-ffi` module. Since a `prebuilt-ffi` module contains platform-dependent assets (such as static C libraries), APIs in `filecoin-ffi` will need to delegate calls to the "platform-specific" `prebuilt-ffi` module.

Currently, `filecoin-ffi` supports multiple (GOOS + GOARCH) combinations. Therefore, we must modify every public API in `filecoin-ffi` to have a variant for each (GOOS + GOARCH) combination. These variants will delegate the call to the corresponding platform-specific `prebuilt-ffi` module instead of calling into the CGO bindings as they do today (except when users build `filecoin-ffi` from source).

Furthermore, it is crucial to ensure complete backward compatibility for current users of `filecoin-ffi`, including those who opt to compile `filecoin-ffi` from the source.

This objective can be achieved using Go build tags. 
For illustration, consider the existing public `ffi.Hash()` API in the `bls.go` file of `filecoin-ffi`. Below is an outline of the necessary variants for this API:

```go
prebuilt_bls_darwin_arm64.go
//go:build cgo && darwin && arm64 && !ffi_source
// +build cgo,darwin,arm64
// +build !ffi_source
package ffi
import (
	prebuilt "module.fil.org/prebuilt-ffi-darwin-arm64"
    // When building for darwin/arm64, Go tooling automatically selects this file due to the specified build 
    // tags. It then fetches the "prebuilt-ffi-darwin-arm64" module from the module proxy, using the version 
    // specified in the `go.mod` file.
)

// Hash computes the digest of a message
func Hash(message Message) Digest {
	return prebuilt.Hash(message)
}
```

```go
prebuilt_bls_linux_amd64.go
//go:build cgo && linux && amd64 && !ffi_source
// +build cgo,linux,amd64
// +build !ffi_source
package ffi
import (
	prebuilt "module.fil.org/prebuilt-ffi-linux-amd64"
)

// Hash computes the digest of a message
func Hash(message Message) Digest {
	return prebuilt.Hash(message)
}
```

```go
// Build from source
bls__source.go
//go:build ffi_source
// +build ffi_source
// Same code as we have today
```

And so on and so forth for each (GOOS + GOARCH) combination for each of those files containing public APIs that currently call into the CGO bindings.

The logical next question to ask is how are the `prebuilt-ffi-{GOOS}-{GOARCH}` modules mentioned here created and where are they hosted?

### Building and publishing the prebuit-ffi modules + CI changes

For each release of `filecoin-ffi`, CI will build and publish the corresponding `prebuilt-ffi-{GOOS}-{GOARCH}` modules as "go mod compatible zip files" to the Github release assets page for `filecoin-ffi` for each supported combination of (GOOS + GOARCH).

In addition to the pre-built zip modules, we will also need to publish the corresponding `go.mod` and "info" meta for each pre-built module (go tooling needs these -> more details in the Go Module Proxy section below) . These can be created synthetically and can be persisted in the Github release assets as well.


We already have some flavour of this today. See the `Assets` section [here](https://github.com/filecoin-project/filecoin-ffi/releases/tag/ed08caaf8778e1b6).


The high level steps to create these assets for each `prebuilt-ffi-{GOOS}-{GOARCH}` module are as follows:

1. Clone `filecoin-ffi` source on a machine with {GOOS}X and {GOARCH}Y.
2. Remove all the `prebuilt_*` files
3. Build it from source to create the prebuilt assets
4. Remove all the transient build assets in `rust/target` dir
5. Create the `prebuilt_bls_{GOOS}_{GOARCH}.info` and `prebuilt_bls_{GOOS}_{GOARCH}.mod` files (the latter can be created by removing the existing `go.mod` file and running `go mod tidy` to generate a new one)
5. Zip it up using something like `https://github.com/aarshkshah1992/prebuilt-ffi-zipper` (the directories inside the zip just need to follow a specific hierarchy)
5. Publish the `prebuilt_bls_{GOOS}_{GOARCH}.zip`, `prebuilt_bls_{GOOS}_{GOARCH}.mod` and `prebuilt_bls_{GOOS}_{GOARCH}.info` files to the Github release assets page for the `filecoin-ffi` repo

### Building and hosting a lightweight custom Go Module Proxy at the module DNS/namespace

We need a lightweight minimal HTTPs server running at `https://prebuilt-ffi.io` that can serve the `prebuilt-ffi-{GOOS}-{GOARCH}` modules as zip files for each supported combination of (GOOS + GOARCH).

### Running a light weight custom module proxy
A minimal HTTPS server(referred to as a "module proxy" in the Go world) must be established at `https://fil.org` to serve all the `fil.org/prebuilt-ffi-{GOOS}-{GOARCH}` Go modules to go tooling. This server could be implemented using a Cloudflare Worker or a custom-managed HTTPS server.

This server is essential because the Google Go Module Proxy does not accommodate custom domains, and GOPROXY in `direct` mode cannot retrieve modules directly from `https://fil.org` since the modules and pre-built assets will be hosted/persisted on GitHub. To address this, we need a redirection mechanism from `https://fil.org` to the appropriate GitHub URLs/assets. Fortunately, Go tooling is capable of handling 3XX redirects, allowing all module requests from `https://fil.org` to be redirected to the respective GitHub URLs/assets. This redirection ensures that the module server incurs minimal ingress/egress/compute costs, functioning primarily as a redirecting proxy.

**This server will have to serve the following APIs:**
1. GET https://fil.org/prebuilt-ffi-{GOOS}-{GOARCH}?go-get=1

This API must respond with the following HTML which specifies the URL go tooling should fetch the module assets from

```html
<meta name="go-import" content="fil.org/prebuilt-ffi-{GOOS}-{GOARCH} mod https://fil.org">
```
Go tooling will now use the URL specified in the above response and send the following API requests:

2. GET https://fil.org/fil.org/prebuilt-ffi-{GOOS}-{GOARCH}/@v/{$version}.info

Here `{$version}` refers to the go module semver. More details about what this and the the APIs enlisted below
need to return can be found at https://go.dev/ref/mod#goproxy-protocol. The important point here is that
this API can be implemented by doing a redirect to the Google Module proxy.

This redirects to the `prebuilt-ffi-{GOOS}-{GOARCH}.info` file in release assets

3. GET https://module.fil.org/fil.org/prebuilt-ffi-{GOOS}-{GOARCH}/@v/{$version}.mod

This redirects to the `prebuilt-ffi-{GOOS}-{GOARCH}.mod` file in release assets

4. GET https://module.fil.org/fil.org/prebuilt-ffi-{GOOS}-{GOARCH}/@v/{$version}.zip

This redirects to the `prebuilt-ffi-{GOOS}-{GOARCH}.zip` file in release assets

Note that one limitation of the above approach is that users will not be able to depend on unqualified/`latest` versions of prebuit-ffi.