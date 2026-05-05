# Router Package

Sistema de enrutamiento HTTP con soporte para groups y middlewares.

## Estructura

```go
type Router struct {
    routes      *RouteCollection
    middlewares []interface{}
}
```

## Métodos de Ruta

### `Get(path string, handler interface{}) *Router`
### `Post(path string, handler interface{}) *Router`
### `Put(path string, handler interface{}) *Router`
### `Delete(path string, handler interface{}) *Router`
### `Patch(path string, handler interface{}) *Router`

Registran rutas para sus respectivos métodos HTTP. Retornan el Router para encadenamiento.

## Middlewares

### `Middleware(middleware interface{}) *Router`
Añade un middleware global al router.

### `Group(callback func(*Router), middlewares []interface{}) *Router`
Crea un grupo de rutas con middlewares específicos. Combina middlewares del router padre con los del grupo.

## Resolución

### `Resolve(method, uri string) *Route`
Busca una ruta que coincida con el método y URI dados.

### `Routes() *RouteCollection`
Retorna la colección completa de rutas registradas.

## Uso

```go
router := router.New()
router.Get("/users", handler).
    Post("/users", createHandler).
    Middleware(loggingMiddleware)

router.Group(func(r *router.Router) {
    r.Get("/admin", adminHandler)
}, []interface{}{authMiddleware})
```
