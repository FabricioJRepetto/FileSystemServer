package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type FileRequest struct {
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
	DeleteFile string `json:"deleteFile"`
	MoveFile   bool   `json:"moveFile"`
}

// Copia, renombra, mueve y elimina archivos de imagenes de cheques.
func manageCheckFilesHandler(w http.ResponseWriter, r *http.Request) {
	ORIGIN_PATH := "C:\\ncr-cc\\temp\\ipm\\"
	FILES_PATH := "C:\\ncr-cc\\temp\\checks-images\\"

	createDirs(w, FILES_PATH)

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		logMensaje("ERROR", "Método no permitido: "+r.Method)
		return
	}

	// Decodificar el JSON del body
	var actions []FileRequest
	err := json.NewDecoder(r.Body).Decode(&actions)
	if err != nil {
		http.Error(w, "Error en el body", http.StatusBadRequest)
		logMensaje("ERROR", "Error en el body: "+err.Error())
		return
	}

	for _, action := range actions {
		// Mover el archivo .tif renombrado
		if action.OldName != "" && action.NewName != "" {
			err := os.Rename(ORIGIN_PATH+action.OldName, FILES_PATH+action.NewName)
			if err != nil {
				http.Error(w, "Error al renombrar: "+err.Error(), http.StatusInternalServerError)
				logMensaje("ERROR", "Error al renombrar: "+err.Error())
				return
			} else {
				logMensaje("STATUS", "Archivo renombrado: "+action.OldName+" a "+action.NewName)
			}
		}

		// Eliminar archivo .jpg
		if action.DeleteFile != "" {
			err := os.Remove(ORIGIN_PATH + action.DeleteFile)
			if err != nil {
				http.Error(w, "Error al eliminar: "+err.Error(), http.StatusInternalServerError)
				logMensaje("ERROR", "Error al eliminar: "+err.Error())
				return
			} else {
				logMensaje("STATUS", "Archivo eliminado: "+action.DeleteFile)
			}
		}

		// Mover archivo a recurso compartido
		if action.MoveFile == true && action.NewName != "" {
			moveToSharedFolder(FILES_PATH + action.NewName)
		}
	}

	// Desmontar carpeta de red al finalizar proceso
	desmontarCarpetaRed()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"Archivos procesados correctamente"}`)
	logMensaje("OK", "✅ Archivos procesados correctamente.")
}
