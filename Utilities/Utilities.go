package Utilities

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"proyecto1/Structs"
	"strings"
)

// Funcion para crear un archivo binario
func CreateFile(name string) error {
	//Se asegura que el archivo existe
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Err CreateFile dir==", err)
		return err
	}

	// Crear archivo
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Err CreateFile create==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

// Funcion para abrir un archivo binario ead/write mode
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

// Funcion para escribir un objecto en un archivo binario
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

// Funcion para leer un objeto de un archivo binario
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

func DeleteFile (name string) error {
	err := os.Remove(name)
	if err != nil {
		fmt.Println("Err DeleteFile==", err)
		return err
	}
	return nil
}

func GenerateReportMBR(mbr Structs.MRB, ebrs []Structs.EBR, outputPath string, file *os.File)error{
	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	// Crear el archivo .dot donde se generará el reporte
	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	// Iniciar el contenido del archivo en formato Graphviz (.dot)
	content := "digraph G {\n"
	content += "\tnode [shape=none, margin=0]\n"

	//Primera tabla
	content += "tabla1 [label=<\n"
	content += "<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"10\" bgcolor=\"#f7f7f7\">\n"
	content += "<tr>\n"
	content += "<td bgcolor=\"#003366\" colspan=\"2\" align=\"center\">"
	content += "<font color=\"white\"><b>Reporte del MBR</b></font>"
	content += "</td>\n"
	content += "</tr>\n"

	// Subgrafo del MBR Comenzamos con la informacion del MBR
	content += "<tr>\n"
	content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
	content += "<font color=\"white\"><b>MBR Tamaño</b></font>"
	content += "</td>\n"
	content += "<td bgcolor=\"#87cefa\" align=\"left\">"
	content +=	"<font color=\"black\">"
	content += fmt.Sprintf("%d", mbr.MbrSize)
	content += "</font>"
	content += "</td>\n"
	content += "</tr>\n"

	content += "<tr>\n"
	content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
	content += "<font color=\"white\"><b>MBR Fecha de Creación</b></font>"
	content += "</td>\n"
	content += "<td bgcolor=\"#87cefa\" align=\"left\">"
	content +=	"<font color=\"black\">"
	content += string(mbr.CreationDate[:])
	content += "</font>"
	content += "</td>\n"
	content += "</tr>\n"

	content += "<tr>\n"
	content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
	content += "<font color=\"white\"><b>MBR Signature</b></font>"
	content += "</td>\n"
	content += "<td bgcolor=\"#87cefa\" align=\"left\">"
	content +=	"<font color=\"black\">"
	content += fmt.Sprintf("%d", mbr.Signature)
	content += "</font>"
	content += "</td>\n"
	content += "</tr>\n"
	
	// Recorrer las particiones del MBR en orden

	for i := 0; i < 4; i++ {
		part := mbr.Partitions[i]
		if part.Size > 0 { // Si la partición tiene un tamaño válido
			partName := strings.TrimRight(string(part.Name[:]), "\x00") // Limpiar el nombre de la partición

			content += "<tr>\n"
			content += "<td bgcolor=\"#003366\" colspan=\"2\" align=\"center\">"
			content += fmt.Sprintf("<font color=\"white\"><b>Partición %s</b></font>", partName)
			content += "</td>\n"
			content += "</tr>\n"

			//Status
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Status</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += string(part.Status[:])
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"

			//Type
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Type</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += string(part.Type[:])
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"

			//Fit
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Fit</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += string(part.Fit[:])
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"

			//Start
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Start</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += fmt.Sprintf("%d", part.Start)
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"

			//Size
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Size</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += fmt.Sprintf("%d", part.Size)
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"

			//Name
			content += "<tr>\n"
			content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
			content += "<font color=\"white\"><b>Name</b></font>"
			content += "</td>\n"
			content += "<td bgcolor=\"#87cefa\" align=\"left\">"
			content +=	"<font color=\"black\">"
			content += partName
			content += "</font>"
			content += "</td>\n"
			content += "</tr>\n"
			

			// Si la partición es extendida, leer los EBRs
			if string(part.Type[:]) == "e" {
				// Recolectamos todos los EBRs en orden
				ebrPos := part.Start
				var ebrList []Structs.EBR
				for {
					var ebr Structs.EBR
					err := ReadObject(file, &ebr, int64(ebrPos)) // Asegúrate de que la función ReadObject proviene de Utilities
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					ebrList = append(ebrList, ebr)

					// Si no hay más EBRs, salir del bucle
					if ebr.PartNext == -1 {
						break
					}

					// Mover a la siguiente posición de EBR
					ebrPos = ebr.PartNext
				}

				// Ahora agregamos los EBRs en orden correcto

				for j, ebr := range ebrList {
					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00") // Limpiar el nombre del EBR

					content += "<tr>\n"
					content += "<td bgcolor=\"#003366\" colspan=\"2\" align=\"center\">"
					content += fmt.Sprintf("<font color=\"white\"><b>Particion logica %d</b></font>", j+1)
					content += "</td>\n"
					content += "</tr>\n"

					//part_status
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_status</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += "0"
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"

					//part_next
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_next</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += fmt.Sprintf("%d", ebr.PartNext)
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"

					//part_fit
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_fit</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += string(ebr.PartFit)
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"

					//part_start
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_start</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += fmt.Sprintf("%d", ebr.PartStart)
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"

					//part_size
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_size</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += fmt.Sprintf("%d", ebr.PartSize)
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"

					//part_name
					content += "<tr>\n"
					content += "<td bgcolor=\"#1e90ff\" align=\"left\">"
					content += "<font color=\"white\"><b>Part_name</b></font>"
					content += "</td>\n"
					content += "<td bgcolor=\"#87cefa\" align=\"left\">"
					content +=	"<font color=\"black\">"
					content += ebrName
					content += "</font>"
					content += "</td>\n"
					content += "</tr>\n"
					
				}
			}
		}
	}

	content += "</table>\n"

	// Cerrar la tabla 1
	content += ">];\n"

	// Cerrar el archivo .dot
	content += "}\n"

	// Escribir el contenido en el archivo .dot
	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte MBR generado exitosamente en:", dotFilePath)
	return nil
}