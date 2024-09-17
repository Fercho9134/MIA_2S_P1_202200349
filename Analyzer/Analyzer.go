package Analyzer

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"proyecto1/DiskManagement"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

type CommandRequest struct {
	Commands []string `json:"commands"`
}

// Estructura para el JSON de respuesta
type CommandResponse struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

// AnalyzeHandler maneja la solicitud HTTP y ejecuta los comandos
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Recibiendo solicitud")

	var request CommandRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, "Error decodificando JSON", http.StatusBadRequest)
		log.Println("Error decodificando JSON:", err)
		return
	}

	var responses []CommandResponse
	var mensaje string

	for _, command := range request.Commands {
		//Antes de ejecutar el comando reviamos si esta linea es un comentario
		//Los comentarios tendrán un # al inicio
		if strings.HasPrefix(command, "#") {
			responses = append(responses, CommandResponse{
				Command: "Comentario",
				Message: fmt.Sprintf("> Comentario: %s", command),
			})
			log.Printf("Comentario: %s\n", command)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(responses)
			continue
		}

		commandName, params := getCommandAndParams(command)
		log.Println("Ejecutando comando:", commandName, "con parámetros:", params)
		mensaje = fmt.Sprintf("> Comando %s con parámetros: %s ejecutado exitosamente", commandName, params)
		particionesMontadasTxt := "\n> Particiones montadas:\n"
		err = AnalyzeCommnad(commandName, params)
		if err != nil {
			if commandName == "mount" {
				particionesMontadas := DiskManagement.GetMountedPartitions()
				for _, particiones := range particionesMontadas {
					for _, particion := range particiones {
						particionesMontadasTxt += fmt.Sprintf("Path: %s, Name: %s, ID: %s, Status: %d\n", particion.Path, particion.Name, particion.ID, particion.Status)

					}
				}
				//Devolvemos el mensaje de error y las particiones montadas
				responses = append(responses, CommandResponse{
					Command: commandName,
					Message: fmt.Sprintf("> %s\n%s", err.Error(), particionesMontadasTxt),
				})
			} else {
				responses = append(responses, CommandResponse{
					Command: commandName,
					Message: fmt.Sprintf("> %s", err.Error()),
				})
			}
		} else {
			if commandName == "mount" {
				particionesMontadas := DiskManagement.GetMountedPartitions()
				for _, particiones := range particionesMontadas {
					for _, particion := range particiones {
						
						particionesMontadasTxt += fmt.Sprintf("\tPath: %s, Name: %s, ID: %s, Status: %d\n", particion.Path, particion.Name, particion.ID, particion.Status)
					}
				}

				responses = append(responses, CommandResponse{
					Command: commandName,
					Message: fmt.Sprintf("%s\n%s", mensaje, particionesMontadasTxt),
				})

			} else {
				responses = append(responses, CommandResponse{
					Command: commandName,
					Message: mensaje,
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responses)
	}
}

func ImprimirHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hola mundo")
}

func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input
}

func AnalyzeCommnad(command string, params string) error {
	if strings.Contains(command, "mkdisk") {
		return fn_mkdisk(params)
	} else if strings.Contains(command, "rmdisk") {
		return fn_rmdisk(params)
	} else if strings.Contains(command, "fdisk") {
		return fn_fdisk(params)
	} else if strings.Contains(command, "mount") {
		return fn_mount(params)
	} else if strings.Contains(command, "rep") {
		return fn_rep(params)
	} else {
		return fmt.Errorf("Error: Comando %s inválido o no encontrado", command)
	}
}

func fn_mkdisk(params string) error {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Encontrar la flag en el input
	matches := re.FindAllStringSubmatch(params, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *size <= 0 {
		return fmt.Errorf("Error: La cantidad debe ser mayor a 0")
	}
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		return fmt.Errorf("Error: El fit debe ser 'bf', 'ff', o 'wf'")
	}
	if *unit != "k" && *unit != "m" {
		return fmt.Errorf("Error: Las unidades deben ser 'k' o 'm'")
	}
	if *path == "" {
		return fmt.Errorf("Error: La ruta es obligatoria")
	}

	// Llamar a la función
	DiskManagement.Mkdisk(*size, *fit, *unit, *path)
	return nil
}

func fn_rmdisk(params string) error {
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")

	matches := re.FindAllStringSubmatch(params, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" {
		return fmt.Errorf("Error: La ruta es obligatoria")
	}

	err := DiskManagement.Rmdisk(*path)
	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}
	return nil
}

func fn_fdisk(params string) error {
	// Definir flags
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "k", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "", "Ajuste")

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(params, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	// Validaciones
	if *size <= 0 {
		return fmt.Errorf("Error: Size debe ser mayor a 0")
	}
	if *path == "" {
		return fmt.Errorf("Error: Path es obligatorio")
	}
	if *fit == "" {
		*fit = "wf"
	}
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		return fmt.Errorf("Error: Fit debe ser 'bf', 'ff', o 'wf'")
	}
	if *unit != "k" && *unit != "m" && *unit != "b" {
		return fmt.Errorf("Error: Unidad debe ser 'k', 'm', o 'b'")
	}
	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		return fmt.Errorf("Error: Tipo debe ser 'p', 'e', o 'l'")
	}
	if *name == "" {
		return fmt.Errorf("Error: Name es obligatorio")
	}

	// Llamar a la función
	err := DiskManagement.Fdisk(*size, *path, *name, *unit, *type_, *fit)
	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}

	return nil
}

func fn_mount(params string) error {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" {
		return fmt.Errorf("Error: Path es obligatorio")
	}

	if *name == "" {
		return fmt.Errorf("Error: Name es obligatorio")
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	nombreMinuscula := strings.ToLower(*name)
	err := DiskManagement.Mount(*path, nombreMinuscula)
	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}
	return nil
}

func fn_rep(params string) error {
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	name := fs.String("name", "", "Nombre")
	path := fs.String("path", "", "Ruta")
	id := fs.String("id", "", "ID")
	path_file_ls := fs.String("path_file_ls", "", "Ruta del archivo")

	matches := re.FindAllStringSubmatch(params, -1)
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *name == "" {
		return fmt.Errorf("Error: Name es obligatorio")
	}
	if *path == "" {
		return fmt.Errorf("Error: Path es obligatorio")
	}
	if *id == "" {
		return fmt.Errorf("Error: ID es obligatorio")
	}

	//Verificamos si la particion con la id dada esta montada
	montada := false
	var pathDisco string
	for _, particiones := range DiskManagement.GetMountedPartitions() {
		for _, particion := range particiones {
			if particion.ID == *id {
				montada = true
				pathDisco = particion.Path
			}
		}
	}

	if !montada {
		return fmt.Errorf("Error: La partición con ID %s no está montada", *id)
	}

	reportsDir := filepath.Dir(*path)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error: %s", err.Error())
	}

	switch *name {
	case "mbr":
		file, err := Utilities.OpenFile(pathDisco)
		if err != nil {
			return fmt.Errorf("Error: %s", err.Error())
		}
		defer file.Close()

		var TempMBR Structs.MRB
		if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
			return fmt.Errorf("Error: %s", err.Error())
		}

		var ebrs []Structs.EBR
		for i := 0; i < 4; i++ {
			if string(TempMBR.Partitions[i].Type[:]) == "e" {
				log.Println("Partición extendida encontrada", string(TempMBR.Partitions[i].Name[:]))

				ebrPosition := TempMBR.Partitions[i].Start
				ebrCounter := 1

				//Leemos todos los ebrs de la partición extendida
				for ebrPosition != -1 {
					log.Println("Leyendo EBR en posicion", ebrPosition)
					var TempEBR Structs.EBR
					if err := Utilities.ReadObject(file, &TempEBR, int64(ebrPosition)); err != nil {
						return fmt.Errorf("Error: %s", err.Error())
					}

					ebrs = append(ebrs, TempEBR)
					Structs.PrintEBR(TempEBR)

					ebrPosition = TempEBR.PartNext
					ebrCounter++

					if ebrPosition == -1 {
						break
					}
				}
			}

		}

		pathReporte := *path
		if err:= Utilities.GenerateReportMBR(TempMBR, ebrs, pathReporte, file); err != nil {
			return fmt.Errorf("Error: %s", err.Error())
		}else{
			log.Println("Reporte MBR generado exitosamente")
			dotFile := strings.TrimSuffix(pathReporte, filepath.Ext(pathReporte)) + ".dot"
			outupPng := strings.TrimSuffix(pathReporte, filepath.Ext(pathReporte)) + ".png"

			cmd := exec.Command("dot", "-Tpng", dotFile, "-o", outupPng)
			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("Error: %s", err.Error())
			}else{
				log.Println("Imagen generada exitosamente")
			}
		}

	}

	log.Println(path_file_ls)

	return nil
}
