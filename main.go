package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type FileRequest struct {
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
	DeleteFile string `json:"deleteFile"`
	MoveFile   bool   `json:"moveFile"`
}

func main() {
	// Handler para la ruta raíz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "File System Server 🎈")
	})

	http.HandleFunc("/manageCheckFiles", manageCheckFilesHandler)

	// Puerto y dirección donde va a escuchar
	port := ":8081"
	fmt.Println("🎈 Server listening on http://localhost" + port)

	// Levantar el servidor
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

func manageCheckFilesHandler(w http.ResponseWriter, r *http.Request) {
	ORIGIN_PATH := "C:\\ncr-cc\\temp\\ipm\\"
	FILES_PATH := "C:\\ncr-cc\\temp\\checks-images\\"
	SHARED_PATH := "C:\\ncr-cc\\recurso-compartido\\"

	createDirs(w, FILES_PATH, SHARED_PATH)

	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar el JSON del body
	var actions []FileRequest
	err := json.NewDecoder(r.Body).Decode(&actions)
	if err != nil {
		http.Error(w, "Error en el body", http.StatusBadRequest)
		return
	}

	for _, action := range actions {
		// Mover el archivo .tif renombrado
		if action.OldName != "" && action.NewName != "" {
			err := os.Rename(ORIGIN_PATH+action.OldName, FILES_PATH+action.NewName)
			if err != nil {
				http.Error(w, "Error al renombrar: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Eliminar archivo .jpg
		if action.DeleteFile != "" {
			err := os.Remove(ORIGIN_PATH + action.DeleteFile)
			if err != nil {
				http.Error(w, "Error al eliminar: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Mover archivo a recurso compartido
		if action.MoveFile == true && action.NewName != "" {
			err := copyFile(FILES_PATH+action.NewName, SHARED_PATH+action.NewName)
			if err != nil {
				http.Error(w, "Error al mover: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"Archivos procesados correctamente"}`)
}

// Función para copiar archivo
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// Crea los directorios si no existen
func createDirs(w http.ResponseWriter, paths ...string) {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				fmt.Println("Error al crear directorio:", path, err)
				http.Error(w, "Error al crear directorio: "+err.Error(), http.StatusInternalServerError)
			} else {
				fmt.Println("Directorio creado:", path)
			}
		}
	}
}
