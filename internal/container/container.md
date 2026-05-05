# Container Package

Sistema de inyección de dependencias (DI) con resolución automática de tipos.

## Estructura

```go
type Container struct {
    services  map[string]interface{}
    instances map[string]interface{}
}
```

## Métodos

### `Set(id string, concrete interface{}) *Container`
Registra un servicio en el container. Retorna el container para encadenamiento.

### `Get(id string) interface{}`
Obtiene un servicio o instancia por su identificador. Si existe una instancia, la retorna; si no, crea una nueva del servicio registrado.

### `Has(id string) bool`
Verifica si un servicio o instancia existe en el container.

### `Call(function interface{}, params map[string]interface{}) interface{}`
Invoca una función inyectando dependencias desde el container. Usa reflexión para:
- Buscar servicios por tipo (nombre del parámetro)
- Usar params proporcionados explícitamente
- Usar zero values si no encuentra nada

## Uso

```go
c := container.New()
c.Set("Logger", &log.Logger{})

result := c.Call(func(l *log.Logger) {
    l.Info("Hello")
}, nil)
```
