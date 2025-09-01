package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var ORIGIN_PATH string = "C:\\ncr-cc\\temp\\ipm\\"
var FILES_PATH string = "C:\\ncr-cc\\temp\\checks-images\\"
var LOG_DIR string = "C:\\ncr-cc\\logs\\"
var LOG_PATH string = LOG_DIR + "TerminalSystemManager-" + time.Now().Format("2006-01-02") + ".log"

// jaja boludoo
func main() {
	createDirs(FILES_PATH, FILES_PATH, LOG_DIR)

	// Handler para la ruta raíz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Terminal System Manager Runnig")
	})

	http.HandleFunc("/manageCheckFiles", manageCheckFilesHandler)
	http.HandleFunc("/depositCanceled", handleCanceledDeposit)

	// Puerto y dirección donde va a escuchar
	port := ":8081"
	logMensaje("STATUS", "✓ Terminal System Manager listening on http://localhost"+port)

	// Levantar el servidor
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logMensaje("ERROR", "Error al iniciar Terminal System Manager: "+err.Error())
	}
}

type FileRequest struct {
	OldName    string `json:"oldName"`
	NewName    string `json:"newName"`
	DeleteFile string `json:"deleteFile"`
	MoveFile   bool   `json:"moveFile"`
}

// Copia, renombra, mueve y elimina archivos de imagenes de cheques.
func manageCheckFilesHandler(w http.ResponseWriter, r *http.Request) {

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

	logMensaje("STATUS", "Renombrando archivos...")
	for k, action := range actions {
		// Mover el archivo .tif renombrado
		logMensaje("STATUS", "Archivo número ["+fmt.Sprint(k+1)+"] de ["+fmt.Sprint(len(actions))+"]")
		if action.OldName != "" && action.NewName != "" {
			err := os.Rename(ORIGIN_PATH+action.OldName, FILES_PATH+action.NewName)
			if err != nil {
				logMensaje("ERROR", "Error al renombrar: "+err.Error())
				http.Error(w, "Error al renombrar: "+err.Error(), http.StatusInternalServerError)
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
	logMensaje("OK", "✓ Archivos procesados correctamente.")
}

// Elimina archivos de imagenes de cheques actual ante una cancelación de la operación.
func handleCanceledDeposit(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		logMensaje("ERROR", "Método no permitido: "+r.Method)
		return
	}

	logMensaje("STATUS", "Limpiando directorio "+ORIGIN_PATH+" ...")
	err := os.RemoveAll(ORIGIN_PATH)
	if err != nil {
		http.Error(w, "Error al eliminar: "+err.Error(), http.StatusInternalServerError)
		logMensaje("ERROR", "Error al eliminar: "+err.Error())
		return
	} else {
		logMensaje("STATUS", "Directorio "+ORIGIN_PATH+" eliminado")
	}

	errMk := os.MkdirAll(ORIGIN_PATH, 0755)
	if errMk != nil {
		logMensaje("[ERROR]", "Error al crear directorio: "+ORIGIN_PATH+" - "+errMk.Error())
	} else {
		logMensaje("[STATUS]", "✓ Directorio creado: "+ORIGIN_PATH)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"Archivos eliminados correctamente"}`)
	logMensaje("OK", "✓ Archivos eliminados correctamente.")
}

// Logger
func logMensaje(logType string, mensaje string) {
	f, err := os.OpenFile(LOG_PATH, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error al abrir el archivo de log:", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s [%s] %s\n", timestamp, logType, mensaje)
	fmt.Println(logEntry)
	f.WriteString(logEntry)
}

// Crea los directorios si no existen
func createDirs(paths ...string) {
	fmt.Println("******* ****** [STATUS] Creando directorios")
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				fmt.Println("[ERROR] Error al crear directorio: " + path + " - " + err.Error())
			} else {
				fmt.Println("[STATUS] ✓ Directorio creado: " + path)
			}
		}
	}

	// Configuración carpeta de red
	remoteDirectory := os.Getenv("TSM_RemoteDirectory")
	user := os.Getenv("TSM_RemoteUser")
	password := os.Getenv("TSM_RemotePassword")

	if user == "" || password == "" || remoteDirectory == "" {
		logMensaje("ERROR", "Faltan variables de entorno para la configuración de la carpeta de red.")
	} else {
		// Montar carpeta de red
		cmdMap := exec.Command("cmd", "/C", "net", "use", "Z:", remoteDirectory, "/user:"+user, password)
		logMensaje("STATUS", "Montando carpeta de red...")
		errMount := cmdMap.Run()
		if errMount != nil {
			logMensaje("ERROR", "Error al montar la carpeta de red: "+errMount.Error())
			return
		} else {
			logMensaje("STATUS", "Carpeta de red montada correctamente.")
		}
	}

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

// Función para mover archivo a recurso compartido
func moveToSharedFolder(filePath string) {
	// Mover archivo
	cdmMove := exec.Command("cmd", "/C", "move", filePath, "Z:\\")
	err := cdmMove.Run()
	if err != nil {
		logMensaje("ERROR", "Error al mover: "+err.Error())
	} else {
		logMensaje("STATUS", "Archivo movido a recurso compartido")
	}
}

// Desmontar carpeta de red al finalizar proceso
func desmontarCarpetaRed() {
	// Desconectar carpeta de red
	cmdDel := exec.Command("cmd", "/C", "net", "use", "Z:", "/delete", "/y")
	errDel := cmdDel.Run()
	if errDel != nil {
		logMensaje("ERROR", "Error al desmontar la unidad: "+errDel.Error())
	} else {
		logMensaje("STATUS", "Unidad desmontada correctamente.")
	}
}
