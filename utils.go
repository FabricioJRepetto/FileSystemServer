package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func logMensaje(logType string, mensaje string) {
	logPath := "C:\\ncr-cc\\logs\\file-system-server.log"
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
func createDirs(w http.ResponseWriter, paths ...string) {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, 0755)
			if err != nil {
				http.Error(w, "Error al crear directorio: "+err.Error(), http.StatusInternalServerError)
				logMensaje("ERROR", "Error al crear directorio: "+path+" - "+err.Error())
			} else {
				logMensaje("STATUS", "Directorio creado: "+path)
			}
		}
	}

	// Configuración carpeta de red
	remoteShare := `\\10.241.162.33\tfrfile\clearing\imagenesTAS`
	user := `GSCORP.AD\svcTASimag`
	password := `DVX4j32YGkobg0xkSNw4aA==`

	// Montar carpeta de red
	cmdMap := exec.Command("cmd", "/C", "net", "use", "Z:", remoteShare, "/user:"+user, password)
	errMount := cmdMap.Run()
	if errMount != nil {
		logMensaje("ERROR", "Error al montar la carpeta de red: "+errMount.Error())
		return
	} else {
		logMensaje("STATUS", "Carpeta de red montada correctamente.")
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
