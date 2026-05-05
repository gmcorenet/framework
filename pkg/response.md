# Response Package

Proporciona una estructura inmutable para construir respuestas HTTP con body, status code y headers encadenables.

## Estructura

```go
type Response struct {
    body       string
    statusCode int
    headers    map[string]string
}
```

## Métodos

### `NewResponse(body string, statusCode int) *Response`
Constructor que inicializa una respuesta con body y código de estado.

### `Body() string`
Retorna el cuerpo de la respuesta.

### `StatusCode() int`
Retorna el código de estado HTTP.

### `Headers() map[string]string`
Retorna todos los headers establecidos.

### `WithHeader(key, value string) *Response`
Añade un header a la respuesta (encadenable).

### `JSON(data interface{}) *Response`
Convierte el data a JSON, establece el Content-Type header y retorna la respuesta modificada.

## Uso

```go
resp := pkg.NewResponse("", 200).
    WithHeader("X-Custom", "value").
    JSON(map[string]string{"status": "ok"})
```
