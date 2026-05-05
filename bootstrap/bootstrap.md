# Bootstrap Package

Maneja el ciclo de vida de la aplicación: bootstrap y ejecución del kernel.

## Estructura

```go
type Application struct {
    kernel *kernel.Kernel
    ctx    context.Context
}
```

## Funciones

### `New(k *kernel.Kernel) *Application`
Crea una nueva aplicación con el kernel proporcionado.

### `(a *Application) Run() error`
Ejecuta el bootstrap del kernel y luego inicia el servidor.
Retorna error si el bootstrap falla.

### `Boot(k *kernel.Kernel) error`
Función de conveniencia que hace bootstrap del kernel sin iniciar el servidor.

## Uso

```go
app := bootstrap.New(kernel)
if err := app.Run(); err != nil {
    log.Fatal(err)
}
```
