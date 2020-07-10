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
