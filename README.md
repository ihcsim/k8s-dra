![build](https://github.com/ihcsim/k8s-dra/actions/workflows/build.yaml/badge.svg)
![slsa3](https://github.com/ihcsim/k8s-dra/actions/workflows/slsa3.yaml/badge.svg)

# k8s-dra
A sample project to demonstrate K8s dynamic resource allocation derived from
https://github.com/kubernetes-sigs/dra-example-driver.

## Development

To build and test the code:

```sh
make build

make test
```

To generate and update the CRD API Go code:

```sh
make codegen

make codegen-verify
```
