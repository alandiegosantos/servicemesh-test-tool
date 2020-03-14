# servicemesh-test-tool

This tool was developed to quick tests in Service Mesh solutions in Kubernetes. The webserver, once it receives a request, will request from its dependency, configured in dependencies.yaml.

The deployment folder contains an deployment example.

## Compile

```
$ make webserver
```

## Generate docker image

```
$ make docker
```

## Testing how image 

You can make a HTTP request to webapp application and it will request from its dependencies while propagates all HTTP headers. This can be used to trace the request. If any dependency responds with a different code than 200, the status returned will be returned to the client.

```
$ curl http://<webapp addr>/ -v
```