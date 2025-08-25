package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Handler para la ruta raíz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "File System Server 🎈")
	})

	http.HandleFunc("/manageCheckFiles", manageCheckFilesHandler)

	// Puerto y dirección donde va a escuchar
	port := ":8081"
	logMensaje("STATUS", "🎈 Server listening on http://localhost"+port)

	// Levantar el servidor
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logMensaje("ERROR", "Error al iniciar el servidor:"+err.Error())
	}
}
