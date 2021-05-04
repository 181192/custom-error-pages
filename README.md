# Custom error pages for Nginx Ingress controller

When the `custom-http-errors` option is enabled, the Ingress controller configures NGINX so
that it passes several HTTP headers down to its `default-backend` in case of error:

| Header           | Value                                                               |
| ---------------- | ------------------------------------------------------------------- |
| `X-Code`         | HTTP status code returned by the request                            |
| `X-Format`       | Value of the `Accept` header sent by the client                     |
| `X-Original-URI` | URI that caused the error                                           |
| `X-Namespace`    | Namespace where the backend Service is located                      |
| `X-Ingress-Name` | Name of the Ingress where the backend is defined                    |
| `X-Service-Name` | Name of the Service backing the backend                             |
| `X-Service-Port` | Port number of the Service backing the backend                      |
| `X-Request-ID`   | Unique ID that identifies the request - same as for backend service |

custom-error-pages returns JSON or HTML based on the `Accept` header sent by the client.

Here's an example of a 503 `text/html` response:

![503](images/503.gif)

And here's the `application/json` response:

```json
{
  "code": "404",
  "title": "Not Found",
  "messages": [
    "The page you're looking for could not be found."
  ],
  "details": {
    "originalURI": "/status/404",
    "namespace": "default",
    "ingressName": "podinfo",
    "serviceName": "podinfo",
    "servicePort": "9898",
    "requestId": "3a35542b0611d3ee0915e196bffac546"
  }
}
```

custom-error-pages supports 404, 500, 503 and 5xx error codes.

## Configurations

| Flag                    | Evironment variable   | Default             | Description                                         |
|-------------------------|-----------------------|---------------------|-----------------------------------------------------|
| `--debug`               | `DEBUG`               | `false`             | enable debug log                                    |
| `--error-files-path`    | `ERROR_FILES_PATH`    | `./themes/knockout` | the location on disk of files served by the handler |
| `--hide-details`        | `HIDE_DETAILS`        | `false`             | hide request details in response                    |
| `--http-listen-address` | `HTTP_LISTEN_ADDRESS` | `:8080`             | http server address                                 |
| `--log-color`           | `LOG_COLOR`           | `false`             | sets log format to human-friendly, colorized output |

## Getting started - Locally `go get`

```bash
go get github.com/181192/custom-error-pages && custom-error-pages
```

## Getting started - Locally docker

```bash
docker run --rm -p 8080:8080 ghcr.io/181192/custom-error-pages
```

Or build locally

```bash
docker build -t custom-error-pages .
docker run --rm -p 8080:8080 custom-error-pages:latest
```

## Getting started - Ingress Nginx Helm chart

1. Set the `custom-http-errors` config value
2. Enable default backend
3. Set the default backend image

```yaml
controller:
  config:
    custom-http-errors: 404,500,501,502,503
defaultBackend:
  enabled: true
  image:
    repository: ghcr.io/181192/custom-error-pages
    tag: latest
  # optional: change path to theme
  extraEnvs:
  - name: ERROR_FILES_PATH
    value: ./themes/knockout
```

## Getting started - Kustomize

Kustomize manifest are provided with both ingress controller and default backend deployment

```bash
kubectl apply -f k8s/
```

## Credits

- Themes `ghost`, `l7-dark`, `l7-light`, `noise` and `shuffle` are from [tarampampam/error-pages](https://github.com/tarampampam/error-pages).
