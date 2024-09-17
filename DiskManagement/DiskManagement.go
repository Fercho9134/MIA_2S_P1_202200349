package DiskManagement

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"proyecto1/Structs"
	"proyecto1/Utilities"
	"strings"
	"time"
)

// Estructura para representar una partición montada
type MountedPartition struct {
	Path   string
	Name   string
	ID     string
	Status byte // 0: no montada, 1: montada
}

// Mapa para almacenar las particiones montadas, organizadas por disco
var mountedPartitions = make(map[string][]MountedPartition)

// Función para imprimir las particiones montadas
func PrintMountedPartitions() {
	fmt.Println("Particiones montadas:")

	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}

	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s, Status: %c\n",
				partition.Name, partition.ID, partition.Path, partition.Status)
		}
	}
	fmt.Println("")
}

// Funcion obtener particiones montadas, para obtener un arreglo de strings con informacion de las particiones montadas
func GetMountedPartitions() map[string][]MountedPartition{
		return mountedPartitions
}

func Mkdisk(size int, fit string, unit string, path string) {
	fmt.Println("======INICIO MKDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Fit:", fit)
	fmt.Println("Unit:", unit)
	fmt.Println("Path:", path)

	// Validar fit bf/ff/wf
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Println("Error: El Fit debe ser bf, wf or ff")
		return
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Println("Error: Size debe ser mayor a  0")
		return
	}

	// Validar unidar k - m
	if unit != "k" && unit != "m" {
		fmt.Println("Error: Las unidades validas son k o m")
		return
	}

	// Create file
	err := Utilities.CreateFile(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	/*
		Si el usuario especifica unit = "k" (Kilobytes), el tamaño se multiplica por 1024 para convertirlo a bytes.
		Si el usuario especifica unit = "m" (Megabytes), el tamaño se multiplica por 1024 * 1024 para convertirlo a MEGA bytes.
	*/
	// Asignar tamanio
	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1024 * 1024
	}

	// Open bin file
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return
	}

	// Escribir los 0 en el archivo

	// create array of byte(0)
	for i := 0; i < size; i++ {
		err := Utilities.WriteObject(file, byte(0), int64(i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// Crear MRB
	var newMRB Structs.MRB
	newMRB.MbrSize = int32(size)
	newMRB.Signature = rand.Int31() // Numero random rand.Int31() genera solo números no negativos
	copy(newMRB.Fit[:], fit)

	// Obtener la fecha del sistema en formato YYYY-MM-DD
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")
	copy(newMRB.CreationDate[:], formattedDate)

	/*
		newMRB.CreationDate[0] = '2'
		newMRB.CreationDate[1] = '0'
		newMRB.CreationDate[2] = '2'
		newMRB.CreationDate[3] = '4'
		newMRB.CreationDate[4] = '-'
		newMRB.CreationDate[5] = '0'
		newMRB.CreationDate[6] = '8'
		newMRB.CreationDate[7] = '-'
		newMRB.CreationDate[8] = '0'
		newMRB.CreationDate[9] = '8'
	*/

	// Escribir el archivo
	if err := Utilities.WriteObject(file, newMRB, 0); err != nil {
		return
	}

	var TempMBR Structs.MRB
	// Leer el archivo
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	// Print object
	Structs.PrintMBR(TempMBR)

	// Cerrar el archivo
	defer file.Close()

	fmt.Println("======FIN MKDISK======")
}

func Rmdisk(path string) error {
	fmt.Println("======INICIO RMDISK======")
	fmt.Println("Path:", path)

	// Delete file
	err := Utilities.DeleteFile(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	return nil

}

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) error {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)

	// Validar fit (b/w/f)
	if fit != "bf" && fit != "ff" && fit != "wf" {
		return fmt.Errorf("El fit debe ser 'bf', 'ff', o 'wf'")
	}

	// Validar size > 0
	if size <= 0 {
		return fmt.Errorf("El tamaño debe ser mayor a 0")
	}

	// Validar unit (b/k/m)
	if unit != "b" && unit != "k" && unit != "m" {
		return fmt.Errorf("La unidad debe ser 'b', 'k', o 'm'")
	}

	// Ajustar el tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return fmt.Errorf("No se pudo abrir el archivo en la ruta: %s", path)
	}

	var TempMBR Structs.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return fmt.Errorf("No se pudo leer el MBR desde el archivo")
	}

	// Imprimir el objeto MBR
	Structs.PrintMBR(TempMBR)

	fmt.Println("-------------")

	// Validaciones de las particiones
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			totalPartitions++
			usedSpace += TempMBR.Partitions[i].Size

			if TempMBR.Partitions[i].Type[0] == 'p' {
				primaryCount++
			} else if TempMBR.Partitions[i].Type[0] == 'e' {
				extendedCount++
			}
		}
	}

	//Verificar que no exista una partición con el mismo nombre
	for i := 0; i < 4; i++ {
		// Truncar los caracteres nulos en el nombre de la partición antes de la comparación
		partitionName := strings.TrimRight(string(TempMBR.Partitions[i].Name[:]), "\x00")
		if partitionName == name {
			return fmt.Errorf("Ya existe una partición con el nombre '%s'", name)
		}
	}

	// Validar que no se exceda el número máximo de particiones primarias y extendidas
	if totalPartitions >= 4 {
		return fmt.Errorf("No se pueden crear más de 4 particiones primarias o extendidas en total.")
	}

	// Validar que solo haya una partición extendida
	if type_ == "e" && extendedCount > 0 {
		return fmt.Errorf("Solo se permite una partición extendida por disco.")
	}

	// Validar que no se pueda crear una partición lógica sin una extendida
	if type_ == "l" && extendedCount == 0 {
		return fmt.Errorf("No se puede crear una partición lógica sin una partición extendida.")
	}

	// Validar que el tamaño de la nueva partición no exceda el tamaño del disco
	if usedSpace+int32(size) > TempMBR.MbrSize {
		return fmt.Errorf("No hay suficiente espacio en el disco para crear esta partición.")
	}

	// Determinar la posición de inicio de la nueva partición
	var gap int32 = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.Partitions[totalPartitions-1].Start + TempMBR.Partitions[totalPartitions-1].Size
	}

	// Encontrar una posición vacía para la nueva partición
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				// Crear partición primaria o extendida
				TempMBR.Partitions[i].Size = int32(size)
				TempMBR.Partitions[i].Start = gap
				copy(TempMBR.Partitions[i].Name[:], name)
				copy(TempMBR.Partitions[i].Fit[:], fit)
				copy(TempMBR.Partitions[i].Status[:], "0")
				copy(TempMBR.Partitions[i].Type[:], type_)
				TempMBR.Partitions[i].Correlative = int32(totalPartitions + 1)

				if type_ == "e" {
					// Inicializar el primer EBR en la partición extendida
					ebrStart := gap // El primer EBR se coloca al inicio de la partición extendida
					ebr := Structs.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					Utilities.WriteObject(file, ebr, int64(ebrStart))
				}

				break
			}
		}
	}

	// Manejar la creación de particiones lógicas dentro de una partición extendida
	if type_ == "l" {
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr Structs.EBR
				for {
					Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				// Calcular la posición de inicio de la nueva partición lógica
				newEBRPos := ebr.PartStart + ebr.PartSize                    // El nuevo EBR se coloca después de la partición lógica anterior
				logicalPartitionStart := newEBRPos + int32(binary.Size(ebr)) // El inicio de la partición lógica es justo después del EBR

				// Ajustar el siguiente EBR
				ebr.PartNext = newEBRPos
				Utilities.WriteObject(file, ebr, int64(ebrPos))

				// Crear y escribir el nuevo EBR
				newEBR := Structs.EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				Utilities.WriteObject(file, newEBR, int64(newEBRPos))

				// Imprimir el nuevo EBR creado
				fmt.Println("Nuevo EBR creado:")
				Structs.PrintEBR(newEBR)
				fmt.Println("")

				// Imprimir todos los EBRs en la partición extendida
				fmt.Println("Imprimiendo todos los EBRs en la partición extendida:")
				ebrPos = TempMBR.Partitions[i].Start
				for {
					err := Utilities.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					Structs.PrintEBR(ebr)
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				break
			}
		}
		fmt.Println("")
	}

	// Sobrescribir el MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		return fmt.Errorf("No se pudo escribir el MBR en el archivo")
	}

	var TempMBR2 Structs.MRB
	// Leer el objeto nuevamente para verificar
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		return fmt.Errorf("No se pudo leer el MBR desde el archivo")
	}

	// Imprimir el objeto MBR actualizado
	Structs.PrintMBR(TempMBR2)

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Println("======FIN FDISK======")
	fmt.Println("")
	return nil
}

// Función para montar particiones
func Mount(path string, name string) error {
	file, err := Utilities.OpenFile(path)
	if err != nil {
		return fmt.Errorf("No se pudo abrir el archivo en la ruta: %s", path)
	}
	defer file.Close()

	var TempMBR Structs.MRB
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return fmt.Errorf("No se pudo leer el MBR desde el archivo")
	}

	log.Println("Buscando partición con nombre:", name)

	partitionFound := false
	var partition Structs.Partition
	var partitionIndex int

	// Convertir el nombre a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(name))

	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'p' && bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
		// Si se encuentra pero no es primaria, dar mensaje de error indicando que solo las primarias se pueden montar
		if bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			return fmt.Errorf("Solo se pueden montar particiones primarias")
		}
	}

	if !partitionFound {
		return fmt.Errorf("No se encontró una partición con el nombre: '%s'", name)
	}

	// Verificar si la partición ya está montada
	if partition.Status[0] == '1' {
		return fmt.Errorf("La partición ya está montada")
	}

	//fmt.Printf("Partición encontrada: '%s' en posición %d\n", string(partition.Name[:]), partitionIndex+1)

	// Generar el ID de la partición
	diskID := generateDiskID(path)

	// Verificar si ya se ha montado alguna partición de este disco
	mountedPartitionsInDisk := mountedPartitions[diskID]
	var letter byte

	if len(mountedPartitionsInDisk) == 0 {
		// Es un nuevo disco, asignar la siguiente letra disponible
		if len(mountedPartitions) == 0 {
			letter = 'a'
		} else {
			lastDiskID := getLastDiskID()
			lastLetter := mountedPartitions[lastDiskID][0].ID[len(mountedPartitions[lastDiskID][0].ID)-1]
			letter = lastLetter + 1
		}
	} else {
		// Utilizar la misma letra que las otras particiones montadas en el mismo disco
		letter = mountedPartitionsInDisk[0].ID[len(mountedPartitionsInDisk[0].ID)-1]
	}

	// Incrementar el número para esta partición
	carnet := "202200349"
	lastTwoDigits := carnet[len(carnet)-2:]
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, partitionIndex+1, letter)

	// Actualizar el estado de la partición a montada y asignar el ID
	partition.Status[0] = '1'
	copy(partition.Id[:], partitionID)
	TempMBR.Partitions[partitionIndex] = partition
	mountedPartitions[diskID] = append(mountedPartitions[diskID], MountedPartition{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Escribir el MBR actualizado al archivo
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		return fmt.Errorf("No se pudo escribir el MBR en el archivo")
	}

	log.Println("Partición montada con ID: %s\n", partitionID)

	// Imprimir el MBR actualizado
	log.Println("MBR actualizado:")
	Structs.PrintMBR(TempMBR)

	// Imprimir las particiones montadas (solo estan mientras dure la sesion de la consola)
	PrintMountedPartitions()

	return nil

}

// Función para obtener el ID del último disco montado
func getLastDiskID() string {
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

func generateDiskID(path string) string {
	return strings.ToLower(path)
}

// Funcion Clean, desmonta todas las particiones montadas. Debe cambiar el estatus de todas las particiones a 0
func Clean() {
	log.Println("Limpiando particiones montadas...")
	for diskID, partitions := range mountedPartitions {
		file, err := Utilities.OpenFile(partitions[0].Path)
		if err != nil {
			log.Println("No se pudo abrir el archivo en la ruta:", partitions[0].Path)
			continue
		}

		var TempMBR Structs.MRB
		if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
			log.Println("No se pudo leer el MBR desde el archivo")
			continue
		}

		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Size != 0 {
				TempMBR.Partitions[i].Status[0] = '0'
				TempMBR.Partitions[i].Id = [4]byte{}
			}
		}

		if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
			log.Println("No se pudo escribir el MBR en el archivo")
			continue
		}

		log.Println("Particiones desmontadas en disco:", diskID)
		Structs.PrintMBR(TempMBR)
	}

	// Limpiar el mapa de particiones montadas
	mountedPartitions = make(map[string][]MountedPartition)
	log.Println("Particiones montadas limpiadas.")
}
