A POC app that uses `filecoin-ffi` as yet another go module. For this POC, we use
a prebuilt ffi module for [darwin + arm64](https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64).

### How to Run

1) Clone the [custom-mod-proxy-server repository](https://github.com/aarshkshah1992/custom-mod-proxy-server). It starts 
   a local Go Module Proxy Server on port `8080` that:
    a) Serves all go modules by redirecting to the Google Go Proxy server EXCEPT for `github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64`
    b) For `github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64`, it serves the prebuilt ffi directly as a zip file. The zip file comes along with the repo but you also can generate it manually by:
        - Cloning `https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64`
        - Building it from source (see `README` at https://github.com/aarshkshah1992/prebuilt-ffi-darwin-arm64)
        - Removing the `rust/target` directory completely 
        - Running https://github.com/aarshkshah1992/prebuilt-ffi-zipper on it. You'll just have to
        update the path/version there.

2) Run `export GOPROXY=http://localhost:8080` as we need go tooling to use the local go module proxy instead of the default Google go module proxy. This wont be needed once we use DNS.

Run the app -> it should just work

Note that the app depends on a forked `filecoin-ffi` module at https://github.com/aarshkshah1992/filecoin-ffi which in turns depends on the `prebuilt-ffi-darwin-arm64` module. This is as per the design discussed in https://hackmd.io/@mvdan/Hy7iK0TEY.

