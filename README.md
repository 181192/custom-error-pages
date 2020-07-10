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

Here's an example of a 503 `text/html` respons:

![503](images/503.gif)

And here's the `application/json` respons:

```json
{
  "code": "404",
  "title": "Not Found",
  "message": "Not Found",
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

## Getting started - Locally

```
docker build -t custom-error-pages .
docker run --rm -p 8080:8080 custom-error-pages:latest
```

## Getting started - Ingress Nginx Helm chart

1. Set the `custom-http-errors` config value
2. Enable default backend
3. Set the default backend image

```
controller:
  config:
  custom-http-errors: 404,500,501,502,503
defaultBackend:
  enabled: true
  image:
    repository: docker.pkg.github.com/stacc-as/custom-error-pages/custom-error-pages
    tag: latest
```

> NOTE: Github Packages still requires auth even on public repositories. An imagePullSecret needs to be created.
>
> 1.  Create new Github Personal Access Token with read:packages scope at https://github.com/settings/tokens/new.
> 2.  Base64 encode token with username
>
> ```
> export AUTH=$(echo -n <your-github-username>:<token> | base64)
> ```
>
> 3.  Create k8s secret
>
> ```
>  echo '{"auths":{"docker.pkg.github.com":{"auth":"$AUTH"}}}' | kubectl create secret generic dockerconfigjson-github-com --type=kubernetes.io/dockerconfigjson --from-file=.dockerconfigjson=/dev/stdin
> ```
>
> 4. Append the imagePullSecret in `values.yaml` for the ingress-nginx
>
> ```
> imagePullSecrets:
>  - name: dockerconfigjson-github-com
> ```

## Getting started - Kustomize

Kustomize manifest are provided with both ingress controller and default backend deployment

```
kubectl apply -f k8s/
```
