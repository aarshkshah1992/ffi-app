# Prebuilt Filecoin FFI POC

This Proof of Concept (POC) demonstrates the utilization of prebuilt-ffi modules. Currently, it incorporates a [prebuilt ffi module](https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64) specifically for the darwin + arm64 configuration. Future extensions will support all GOOS + GOARCH combinations.

### Running the Application

1) Clone the [custom-mod-proxy-server repository](https://github.com/aarshkshah1992/custom-mod-proxy-server). This sets up a local Go Module Proxy Server on port `8080` that:
    - Redirects requests for all Go modules to the Google Go Proxy server, except for the module `github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64`.
    - Serves the `github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64` module directly as a zip file. This zip file is included with the repository. Alternatively, you can create it manually by following these steps:
        - Clone the repository from `https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64`.
            - Build the source as described in the [README](https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64).
            - Completely remove the `rust/target` directory.
            - Execute the script from [prebuilt-ffi-zipper](https://github.com/aarshkshah1992/prebuilt-ffi-zipper), ensuring to update the path and version as necessary.

2) Run `export GOPROXY=http://localhost:8080` as we need go tooling to use the local go module proxy instead of the default Google go module proxy. This wont be needed once we use DNS.

**Run the app -> it should just work.**

### Forked `https://github.com/filecoin-project/filecoin-ffi`
Note that the app depends on a forked `filecoin-ffi` module at https://github.com/aarshkshah1992/filecoin-ffi which in turns depends on the `prebuilt-ffi-darwin-arm64` module. This is as per the design discussed in https://hackmd.io/@mvdan/Hy7iK0TEY.

### Note on checksumming for the prebuilt-ffi module

Please note that there is a known issue of a `go.sum` mismatch when downloading the `github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64` dependency using Go modules. This occurs because the repository is hosted on GitHub, which is where the Google checksum server retrieves its checksums, but lacks the prebuilt assets. Instead, the actual zip file for the module, which includes the pre-built assets, is served from our local module server. This discrepancy is expected to be resolved once we transition to using DNS, allowing both the Google checksum server and our custom Go module server to source the pre-built ffi module from the same location.

For this proof of concept, the issue has been temporarily addressed by manually updating the checksum for the prebuilt-ffi module in the `go.sum` file of this app.

