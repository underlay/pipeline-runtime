# pipeline-runtime

## Build

```zsh
% go get -trimpath github.com/ipfs/go-ipfs/cmd/ipfs@v0.8.0
% go build -trimpath -buildmode=plugin -o pipeline-runtime.so .
% chmod +x pipeline-runtime.so
% cp pipeline-runtime.so ~/.ipfs/plugins/.
```

## API

The pipeline runtime has three primary internal API endpoints

### `/api/v0/evaluate`

- `POST` a JSON graph to this route. A successful `POST` returns status code `200`.

### `/api/v0/evaluate`

- `POST` a JSON graph to this route. A successful `POST` returns status code `200`.

### `/api/v0/collection`

This route takes two required query parameters `host` and `id`. `host` is a multiaddr of a remote collection server, and `id` is a URI identifying a collection hosted on that server.

- `GET /api/v0/collection?host=/foo/bar&id=http://example.com/some-collection` fetches the specified collection as a .tar.gz archive. A successful `GET` returns status code `200`.
- `POST /api/v0/collection?host=/foo/bar&id=http://example.com/some-collection` takes the .tar.gz archive in the request body and publishes it to the specified collection endpoint. A successful `POST` returns a status code `204`.
