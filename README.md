# Ram 🐏 — golang opiniated continuous testing tool

This is a very opiniated « continuous testing » tool for =Go=.
In a nutshell it does : watch a folder (gopath or not…) and execute
tests when file changes.

It supports:
- changing code in a package will only re-run tests on this package
- changing a test code, it will only re-run *that* test

```shell
$ ram -d ./pkg -d ./cmd -- -v
 INFO Watching directories: ./pkg, pkg/apis, pkg/apis/pipeline, pkg/apis/pipeline/v1alpha1, pkg/client, pkg/client/clientset, pkg/client/clientset/versioned, pkg/client/informers, pkg/client/informers/externalversions, pkg/client/listers, pkg/client/listers/pipeline, pkg/errors, pkg/foo, pkg/logging, pkg/reconciler, pkg/reconciler/testing, pkg/reconciler/v1alpha1, pkg/reconciler/v1alpha1/pipeline, pkg/reconciler/v1alpha1/pipelinerun, pkg/reconciler/v1alpha1/taskrun, pkg/system, ./cmd/, cmd/controller, cmd/controller/kodata, cmd/kubeconfigwriter, cmd/kubeconfigwriter/kodata, cmd/webhook, cmd/webhook/kodata
 INFO Run go test -v ./${dir}
┌─────────────┬──────────────────────┐
│ filewatcher │ go test -v ./pkg/foo │
└─────────────┴──────────────────────┘
=== RUN   TestBar
--- PASS: TestBar (0.00s)
=== RUN   TestFoo
--- PASS: TestFoo (0.00s)
PASS
ok      github.com/knative/build-pipeline/pkg/foo       (cached)
┌────┬────────────────┬──────────────┐
│ OK │ pkg/foo/foo.go │ 299.754067ms │
└────┴────────────────┴──────────────┘
 WARN Failed to stat pkg/foo/.foo_test.go.swx: stat pkg/foo/.foo_test.go.swx: no such file or directory
┌─────────────┬──────────────────────────────────────────┐
│ filewatcher │ go test -v ./pkg/foo -test.run ^TestFoo$ │
└─────────────┴──────────────────────────────────────────┘
=== RUN   TestFoo
--- PASS: TestFoo (0.00s)
PASS
ok      github.com/knative/build-pipeline/pkg/foo       (cached)
┌────┬─────────────────────┬──────────────┐
│ OK │ pkg/foo/foo_test.go │ 255.015843ms │
└────┴─────────────────────┴──────────────┘
┌─────────────┬──────────────────────────────────────────┐
│ filewatcher │ go test -v ./pkg/foo -test.run ^TestBar$ │
└─────────────┴──────────────────────────────────────────┘
=== RUN   TestBar
--- PASS: TestBar (0.00s)
PASS
ok      github.com/knative/build-pipeline/pkg/foo       (cached)
┌────┬─────────────────────┬──────────────┐
│ OK │ pkg/foo/bar_test.go │ 258.800903ms │
└────┴─────────────────────┴──────────────┘

```
