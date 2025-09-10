# Workaround:

Se desarrolla una aplicación adicional que se encarga de levantar un servidor local. Se ejecuta al levantar el resto de la app en la dirección http://localhost:8081// y escucha los siguientes endpoints:

-   /manageCheckFiles [POST]
-   /depositCanceled [DELETE]
-   /windowFocus [POST]

## 📁 Manejo de imagenes de cheques

> Soluciona el problema de que Terminal no recibe configuración para designar un nombre a las imagenes de cheques, ni para moverlas o eliminarlas de los directorios destino.

**/manageCheckFiles** [POST]

Recibe el siguiente body:

```
[
    {
        oldName: string
        newName: string
        deleteFile: string
        moveFile?: boolean
    }
]
```

oldName: archivo a renombrar.
newName: nuevo nombre para el archivo.
deleteFile: indica un archivo a eliminar.
moveFile: indica si mover el archivo renombrado a la carpeta compartida.

Se encarga de eliminar los .jpg temporales (solo se utilizan como previsualización en la app), renombrar los archivos .tif (las imágenes renombradas quedan en /ncr-cc/temp/checks-images) y moverlos a la carpeta compartida (al día de la fecha: \\10.241.162.33\tfrfile\clearing\imagenesTAS).

> ⚠️ Las credenciales de la carpeta compartida se setean como variables de entorno (_TSM_RemoteDirectory_, _TSM_RemoteUser_, _TSM_RemotePassword_).
> **Hay un script que facilita el proceso en el repositorio.**

**/depositCanceled** [DELETE]

Elimina todos los archivos de la carpeta /ncr-cc/temp/ipm en caso de que la operación se interrumpa por cualquier motivo.

## 🔍 Fix focus de pantalla cliente

> Soluciona el problema de que los inputs no recibian los eventos del teclado, originado a partir de que la pantalla de cliente no estaba en foco, ya que el supervisor se ponía en foco por defecto al levantar las apps y al cerrar sesión de cliente.

**/windowFocus** [POST]

Recibe el campo _WindowTitle_ en la request para buscar ese nombre entre las pantallas activas actualmente y le da foco.

Recibe el siguiente body:

```
{
    WindowTitle: string
}
```
