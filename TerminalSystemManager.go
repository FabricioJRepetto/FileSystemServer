package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

var ORIGIN_PATH string = "C:\\ncr-cc\\temp\\ipm\\"
var FILES_PATH string = "C:\\ncr-cc\\temp\\checks-images\\"
var LOG_DIR string = "C:\\ncr-cc\\logs\\"
var LOG_PATH string = LOG_DIR + "TerminalSystemManager-" + time.Now().Format("2006-01-02") + ".log"

type FocusRequest struct {
	WindowTitle string `json:"windowTitle"`
}

// jaja boludoo
func main() {
	port := ":8082"
	srv := &http.Server{Addr: port}

	logWelcome()
	createDirs(FILES_PATH, FILES_PATH, LOG_DIR)

	// Handler para la ruta raíz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Terminal System Manager Runnig")
	})

	http.HandleFunc("/manageCheckFiles", manageCheckFilesHandler)
	http.HandleFunc("/depositCanceled", handleCanceledDeposit)
	http.HandleFunc("/windowFocus", focusHandler)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logMensaje("STATUS", "Señal recibida cerrando servidor...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	// Puerto y dirección donde va a escuchar
	logMensaje("STATUS", "TSM listening on http://localhost"+port)
	// Levantar el servidor
	log.Fatal("[ERROR] Error al iniciar TSM - ", srv.ListenAndServe())
	logMensaje("STATUS", "Servidor cerrado.")
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

func logWelcome() {
	fmt.Println(`                                           `)
	fmt.Println("ooooooooooooo  .oooooo..o ooo        ooooo ")
	fmt.Println("8'   888   `8 d8P'    `Y8 `88.       .888' ")
	fmt.Println("     888      Y88bo.       888b     d'888  ")
	fmt.Println("     888       `'Y8888o.   8 Y88. .P  888  ")
	fmt.Println("     888           `'Y88b  8  `888'   888  ")
	fmt.Println("     888      oo     .d8P  8    Y     888  ")
	fmt.Println(`    o888o     8""88888P'  o8o        o888o `)
	fmt.Println(`                                           `)
	fmt.Println(`···········································`)
	fmt.Println(`                   Terminal System Manager `)
	fmt.Println(`···········································`)
	fmt.Println(`                                           `)
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
	logMensaje("", "··········· Creando directorios ···········")
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				logMensaje("[ERROR]", "Error al crear directorio: "+path+" - "+err.Error())
			} else {
				logMensaje("[STATUS]", "Directorio creado: "+path)
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
		logMensaje("", "········· Montando carpeta de red ·········")
		errMount := cmdMap.Run()
		if errMount != nil {
			logMensaje("ERROR", "Error al montar la carpeta de red: "+errMount.Error())
			return
		} else {
			logMensaje("STATUS", "Carpeta de red montada correctamente.")
		}
	}

}

//_____ Manejar archivos _____

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

// _____ Cambiar focus de ventanas _____
var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procFindWindowW   = user32.NewProc("FindWindowW")
	procSetForeground = user32.NewProc("SetForegroundWindow")
	// procShowWindow    = user32.NewProc("ShowWindow")
	procKeybdEvent = user32.NewProc("keybd_event")
)

const (
	SW_RESTORE      = 9
	VK_MENU         = 0x12 // ALT key
	KEYEVENTF_KEYUP = 0x0002
)

func focusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logMensaje("ERROR", "Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req FocusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logMensaje("ERROR", "Error en el JSON")
		http.Error(w, "Error en el JSON", http.StatusBadRequest)
		return
	}

	success := setFocusToWindow(req.WindowTitle)
	if success {
		logMensaje("STATUS", "Se cambió el foco a la ventana: "+req.WindowTitle)
		fmt.Fprintf(w, "Se cambió el foco a la ventana: %s", req.WindowTitle)
	} else {
		logMensaje("ERROR", "No se pudo encontrar la ventana: "+req.WindowTitle)
		http.Error(w, "No se pudo encontrar la ventana", http.StatusNotFound)
	}
}

func setFocusToWindow(title string) bool {
	ptr, _ := syscall.UTF16PtrFromString(title)

	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd == 0 {
		return false
	}

	/** Restaurar si está minimizada
	const (
		SW_HIDE            = 0
		SW_NORMAL          = 1
		SW_SHOWNORMAL      = 1
		SW_SHOWMINIMIZED   = 2
		SW_SHOWMAXIMIZED   = 3
		SW_MAXIMIZE        = 3 <- testear
		SW_SHOWNOACTIVATE  = 4
		SW_SHOW            = 5 <- testear
		SW_MINIMIZE        = 6
		SW_SHOWMINNOACTIVE = 7
		SW_SHOWNA          = 8
		SW_RESTORE         = 9 <- saca la ventana del fullscreen
		SW_SHOWDEFAULT     = 10
		SW_FORCEMINIMIZE   = 11
	) */
	// procShowWindow.Call(hwnd, SW_RESTORE)

	// Simular ALT para permitir SetForegroundWindow
	procKeybdEvent.Call(VK_MENU, 0, 0, 0)
	procKeybdEvent.Call(VK_MENU, 0, KEYEVENTF_KEYUP, 0)

	// Intentar cambiar el foco
	procSetForeground.Call(hwnd)

	return true
}
