package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
	func imprimirArgumentos(args []string) {
		fmt.Println("Argumentos:")
		for i, arg := range args {
			fmt.Printf("%d: %s\n", i+1, arg)
		}
	}
*/
func main() {
	args := os.Args[1:]
	if len(args) != 8 {
		fmt.Println("Uso: go run main.go -m [valor_m] -p [valor_p] -orden [archivo_orden_creacion_procesos] -salida [archivo_salida]")
		return
	}

	// Llamar a la función para imprimir los argumentos
	//	imprimirArgumentos(args)

	// Obtener los valores de m y p de los argumentos
	m, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("El valor de m debe ser un número válido.")
		return
	}

	pStr := args[3]
	p, err := strconv.ParseFloat(pStr, 64)
	if err != nil {
		fmt.Println("El valor de p debe ser un número válido.")
		return
	}

	ordenEjecucion := args[5]
	archivoSalida := args[7]

	// Verificar que los archivos de procesos existan
	if err := verificarArchivosProcesos(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Archivos de procesos verificados correctamente (OK)")

	// Verificar que el archivo de orden de ejecución exista
	if _, err := os.Stat(ordenEjecucion); os.IsNotExist(err) {
		fmt.Println("El archivo de orden de ejecución no existe.")
		return
	}

	fmt.Println("Archivo de orden de ejecución verificado correctamente (OK)")

	// Verificar y crear el archivo de salida si está vacío
	if err := verificarCrearArchivoSalida(archivoSalida, m, p); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Archivo de salida verificado y/o creado correctamente (OK)")

	// Iniciar la simulación
	simular(m, p, ordenEjecucion, archivoSalida)
}

type OrdenEjecucion struct {
	TiempoCreacion int
	NombreProceso  string
}

func cargarOrdenEjecucion(archivoOrden string) ([]OrdenEjecucion, error) {
	contenido, err := ioutil.ReadFile(archivoOrden)
	if err != nil {
		return nil, fmt.Errorf("Error al leer el archivo de orden de ejecución: %v", err)
	}

	lineas := strings.Split(string(contenido), "\n")
	orden := []OrdenEjecucion{}

	for _, linea := range lineas {
		if strings.TrimSpace(linea) == "" || strings.HasPrefix(linea, "#") {
			continue
		}

		// Divide la línea en dos partes separadas por el carácter '|'
		parts := strings.Split(linea, "|")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Formato incorrecto en el archivo de orden de ejecución")
		}

		// Elimina los espacios en blanco alrededor de las partes
		tiempoCreacion, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("Error al convertir el tiempo de creación: %v", err)
		}

		nombreProceso := strings.TrimSpace(parts[1])

		orden = append(orden, OrdenEjecucion{
			TiempoCreacion: tiempoCreacion,
			NombreProceso:  nombreProceso,
		})
	}

	/* Imprime la lista ordenada con los valores filtrados
	fmt.Println("Lista ordenada con valores filtrados:")
	for _, o := range orden {
		fmt.Printf("Tiempo de Creación: %d, Nombre de Proceso: %s\n", o.TiempoCreacion, o.NombreProceso)
	}
	*/
	return orden, nil
}

func cargarInstruccionesProceso(nombreProceso string) ([]string, error) {
	archivoProceso := filepath.Join("procesos", nombreProceso)
	contenido, err := ioutil.ReadFile(archivoProceso)
	if err != nil {
		return nil, fmt.Errorf("Error al cargar las instrucciones del proceso %s: %v", nombreProceso, err)
	}

	lineas := strings.Split(string(contenido), "\n")
	instrucciones := []string{}

	for _, linea := range lineas {
		if strings.TrimSpace(linea) == "" || strings.HasPrefix(linea, "#") {
			continue
		}

		instrucciones = append(instrucciones, linea)
	}

	// Imprimir las instrucciones filtradas
	fmt.Println("Instrucciones del proceso", nombreProceso+":")
	for _, instruccion := range instrucciones {
		fmt.Println(instruccion)
	}

	return instrucciones, nil
}

func obtenerProcesoListo(estadoProceso map[string]string) string {
	for proceso, estado := range estadoProceso {
		if estado == "Listo" {
			return proceso
		}
	}
	return ""
}

func escribirTraza(traza *os.File, tiempoCPU int, nombreProceso, instruccion string) {
	traza.WriteString(fmt.Sprintf("%d\t%s\t%s\n", tiempoCPU, nombreProceso, instruccion))
}

func todosTerminados(estadoProceso map[string]string) bool {
	for _, estado := range estadoProceso {
		if estado != "Terminado" {
			return false
		}
	}
	return true
}

func simular(m int, p float64, ordenEjecucion, archivoSalida string) {
	// Inicializar el generador de números aleatorios
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Cargando orden de ejecución de los procesos...")
	orden, err := cargarOrdenEjecucion(ordenEjecucion)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Abriendo archivo de salida para escribir la traza...")
	traza, err := os.Create(archivoSalida)
	if err != nil {
		fmt.Println("Error al abrir el archivo de salida:", err)
		return
	}
	defer traza.Close()

	fmt.Println("Iniciando simulación...")
	procesos := make(map[string][]string)
	estadoProceso := make(map[string]string)
	contadorInstrucciones := make(map[string]int)
	indiceOrden := 0
	procesoEnEjecucion := ""
	cicloCPU := 1

	for {
		//fmt.Println(orden[indiceOrden].NombreProceso)
		//fmt.Println(orden[indiceOrden].TiempoCreacion)
		if indiceOrden < len(orden) && orden[indiceOrden].TiempoCreacion == cicloCPU {
			proceso := orden[indiceOrden].NombreProceso
			fmt.Printf("Cargando instrucciones para el proceso: %s\n", proceso)
			instrucciones, err := cargarInstruccionesProceso(proceso)
			if err != nil {
				fmt.Println("Error al cargar instrucciones del proceso:", err)
				return
			}
			procesos[proceso] = instrucciones
			estadoProceso[proceso] = "Listo"
			contadorInstrucciones[proceso] = 0
			fmt.Printf("Proceso %s creado y listo para ejecución.\n", proceso)
			indiceOrden++
		}

		if procesoEnEjecucion == "" {
			procesoEnEjecucion = obtenerProcesoListo(estadoProceso)
			if procesoEnEjecucion != "" {
				estadoProceso[procesoEnEjecucion] = "Ejecutando"
				fmt.Printf("Proceso %s está ahora ejecutando.\n", procesoEnEjecucion)
			}
		}

		if procesoEnEjecucion != "" {
			instrucciones := procesos[procesoEnEjecucion]
			contador := contadorInstrucciones[procesoEnEjecucion]

			if strings.HasPrefix(instrucciones[contador], "ES") {
				cantidadES, _ := strconv.Atoi(strings.TrimPrefix(instrucciones[contador], "ES"))
				fmt.Printf("Proceso %s ejecutando instrucción E/S, bloqueado por %d ciclos.\n", procesoEnEjecucion, cantidadES)
				estadoProceso[procesoEnEjecucion] = "Bloqueado"
				contadorInstrucciones[procesoEnEjecucion]++
				procesoEnEjecucion = ""
				continue
			}

			fmt.Printf("Proceso %s ejecutando instrucción: %s\n", procesoEnEjecucion, instrucciones[contador])
			contadorInstrucciones[procesoEnEjecucion]++

			if contador == len(instrucciones)-1 {
				fmt.Printf("Proceso %s ha finalizado.\n", procesoEnEjecucion)
				estadoProceso[procesoEnEjecucion] = "Terminado"
				escribirTraza(traza, cicloCPU, procesoEnEjecucion, "Finalizar")
				procesoEnEjecucion = ""
			} else if contador%(m+1) == 0 {
				fmt.Printf("Proceso %s ha alcanzado el límite de ejecución y se moverá de nuevo a listo.\n", procesoEnEjecucion)
				estadoProceso[procesoEnEjecucion] = "Listo"
				procesoEnEjecucion = ""
			}
		}

		for proceso, estado := range estadoProceso {
			if estado == "Ejecutando" {
				instrucciones := procesos[proceso]
				contador := contadorInstrucciones[proceso]
				if contador < len(instrucciones) {
					instruccion := instrucciones[contador]
					escribirTraza(traza, cicloCPU, proceso, instruccion)
				}
			}
		}

		if todosTerminados(estadoProceso) {
			fmt.Println("Todos los procesos han terminado. Simulación finalizada.")
			break
		}

		if procesoEnEjecucion != "" && rand.Float64() < p {
			fmt.Printf("Proceso %s terminado prematuramente debido a una probabilidad de terminación.\n", procesoEnEjecucion)
			estadoProceso[procesoEnEjecucion] = "Terminado"
			escribirTraza(traza, cicloCPU, procesoEnEjecucion, "Finalizar")
			procesoEnEjecucion = ""
		}

		cicloCPU++
		time.Sleep(time.Millisecond)
	}
}

func verificarArchivosProcesos() error {
	// Obtener la lista de archivos en la carpeta "procesos"
	archivos, err := ioutil.ReadDir("procesos")
	if err != nil {
		return fmt.Errorf("Error al leer la carpeta de procesos: %v", err)
	}

	for _, archivo := range archivos {
		// Verificar que el archivo sea un archivo de texto (.txt)
		if !strings.HasSuffix(archivo.Name(), ".txt") {
			continue
		}

		// Verificar que el archivo exista
		if _, err := os.Stat(filepath.Join("procesos", archivo.Name())); os.IsNotExist(err) {
			return fmt.Errorf("El archivo %s no existe.", archivo.Name())
		}
	}

	return nil
}

func verificarArchivoOrden(archivoOrden string) error {
	contenido, err := ioutil.ReadFile(archivoOrden)
	if err != nil {
		return fmt.Errorf("Error al leer el archivo de orden de ejecución: %v", err)
	}

	lineas := strings.Split(string(contenido), "\n")
	if len(lineas) > 0 && strings.TrimSpace(lineas[0]) != fmt.Sprintf("#%s", filepath.Base(archivoOrden)) {
		return fmt.Errorf("El archivo de orden de ejecución no está configurado correctamente.")
	}

	return nil
}

func verificarCrearArchivoSalida(archivoSalida string, m int, p float64) error {
	contenido, err := ioutil.ReadFile(archivoSalida)
	if err != nil {
		// El archivo no existe, crearlo con los valores de m y p
		if err := ioutil.WriteFile(archivoSalida, []byte(fmt.Sprintf("m=%d\np=%.2f\n", m, p)), 0644); err != nil {
			return fmt.Errorf("Error al crear el archivo de salida: %v", err)
		}
		return nil
	}

	// El archivo ya existe, verificar si está vacío
	if len(strings.TrimSpace(string(contenido))) == 0 {
		// El archivo está vacío, escribir los valores de m y p
		if err := ioutil.WriteFile(archivoSalida, []byte(fmt.Sprintf("m=%d\np=%.2f\n", m, p)), 0644); err != nil {
			return fmt.Errorf("Error al escribir en el archivo de salida: %v", err)
		}
		return nil
	}

	return nil
}
