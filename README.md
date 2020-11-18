[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Service Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/service?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/service)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/service?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/service/)

This library provides a common way to manage multiple independent services in a single binary.

<p align="center"><strong>Note: This is a developer documentation.</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.github.io">containerssh.github.io</a>.</p>

## Using this library

This library provides two main components: the `Service` interface and the `Pool` implementation. Services are parts of code that can be started and shut down in a graceful manner. They also provide hooks when the service is ready to handle user requests and just before shutdown.

The service interface is described in [service.go](service.go), while the pool behavior is described in [pool.go](pool.go). The `Pool` itself also implements a service so multiple pools can be nested.

Together these two can be used to start multiple parallel services and shut down the pool when one of the services shuts down. For example:

```go
pool = service.NewPool()
pool.Add(myService1)
pool.Add(myService2)
go func() {
    err := pool.Run()
    // Handle errors here
}
pool.Shutdown(context.Background())
```

Ideally, the pool can be used to handle Ctrl+C and SIGTERM events:

```go
signals := make(chan os.Signal, 1)
signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
go func() {
    if _, ok := <-signals; ok {
        // ok means the channel wasn't closed
        pool.Shutdown(
            context.WithTimeout(
                context.Background(),
                20 * time.Second,
            )
        )
    }
}()
// Wait for the pool to terminate.
pool.Wait()
// We are already shutting down, ignore further signals
signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
// close signals channel so the signal handler gets terminated
close(signals)
```
