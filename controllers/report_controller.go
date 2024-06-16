package controllers

import (
	"GoProcesadorExcel/authentication"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/tealeg/xlsx"
)

// KeyValue representa un par clave-valor
type KeyValue struct {
	Key   string
	Value string
}

// Mensaje estructura del mensaje en el valor del par clave-valor
type Mensaje struct {
	CodigoGeneracion string `json:"CodigoGeneracion"`
	SelloRecibido    string `json:"SelloRecibido"`
	Estado           string `json:"Estado"`
	DescripcionMsg   string `json:"DescripcionMsg"`
}

// Estructura del valor en el par clave-valor
type Valor struct {
	Codigo  int     `json:"Código"`
	Mensaje Mensaje `json:"Mensaje"`
}

func GetReporte(c *gin.Context, rdb *redis.Client) {
	// Obtener el token del encabezado
	token := c.GetHeader("Authorization")

	// Validar el token
	empid, err := authentication.ValidateToken(token)
	if err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obtener el correlativo de la solicitud
	correlativo := c.Param("correlativo")

	// Obtener los datos desde Redis
	data, err := getDatosRedis(rdb, empid, correlativo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	fileName := fmt.Sprintf("%s_Lote_%s.xlsx", empid, correlativo)
	// Generar el informe en Excel
	fileBytes, err := generarInformeExcel(data, fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error al generar el informe en Excel: %v", err)})
		return
	}

	// Codificar los datos en base64
	encodedFile := base64.StdEncoding.EncodeToString(fileBytes)

	// Enviar los datos codificados en base64 como respuesta
	c.JSON(http.StatusOK, gin.H{"ReporteExcel": encodedFile})
}

// obtenerDatosDesdeRedis obtiene los datos desde Redis y los devuelve como una lista de KeyValue
func getDatosRedis(rdb *redis.Client, empid string, correlativo string) ([]KeyValue, error) {
	// Construir el nombre del lote usando el correlativo
	nombreLote := fmt.Sprintf("%s_Lote_%s", empid, correlativo)

	// Obtener todos los pares clave-valor del hash en Redis
	data, err := rdb.HGetAll(context.Background(), nombreLote).Result()
	if err != nil {
		return nil, err
	}

	// Verificar si no se encontraron estados para el correlativo dado
	if len(data) == 0 {
		return nil, fmt.Errorf("no existe un lote con el correlativo %s", correlativo)
	}

	// Crear una lista de KeyValue ordenada
	var result []KeyValue
	for key, value := range data {
		result = append(result, KeyValue{Key: key, Value: value})
	}

	// Ordenar las claves basadas en sus valores numéricos
	sort.Slice(result, func(i, j int) bool {
		return getNumero(result[i].Key) < getNumero(result[j].Key)
	})

	return result, nil
}

// Función para obtener el valor numérico de una clave
func getNumero(clave string) int {
	numeroStr := strings.TrimLeftFunc(clave, func(r rune) bool {
		return !unicode.IsDigit(r) // Eliminar los caracteres no numéricos al principio
	})
	numero, _ := strconv.Atoi(numeroStr)
	return numero
}

// Función para generar un informe en Excel
func generarInformeExcel(data []KeyValue, lote string) ([]byte, error) {
	// Construir el nombre del archivo de Excel del lote
	originalFilePath := filepath.Join("data", "archivos_excel", lote)

	// Verificar si el archivo existe
	_, err := os.Stat(originalFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("el archivo %s no existe", originalFilePath)
	}

	// Abrir el archivo de Excel existente
	originalFile, err := xlsx.OpenFile(originalFilePath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo de Excel: %v", err)
	}

	// Crear un nuevo archivo de Excel
	newFile := xlsx.NewFile()

	// Crear una nueva hoja de informe en el nuevo archivo
	informeSheet, err := newFile.AddSheet("Informe")
	if err != nil {
		return nil, fmt.Errorf("error al añadir la hoja 'Informe' al nuevo archivo de Excel: %v", err)
	}

	// Escribir los encabezados de las columnas en la hoja 'Informe'
	headerRow := informeSheet.AddRow()
	headerRow.AddCell().SetValue("IDDTE")
	headerRow.AddCell().SetValue("CodigoGeneracion")
	headerRow.AddCell().SetValue("SelloRecibido")
	headerRow.AddCell().SetValue("Estado")
	headerRow.AddCell().SetValue("Mensaje")

	// Escribir los datos en la hoja 'Informe'
	for _, kv := range data {
		// Encontrar el índice de "Código"
		startIndex := strings.Index(kv.Value, `Código: `)
		if startIndex == -1 {
			fmt.Printf("Formato inválido para clave %s\n", kv.Key)
			continue
		}

		// Encontrar el índice de "Mensaje"
		mensajeIndex := strings.Index(kv.Value, `Mensaje: {`)
		if mensajeIndex == -1 {
			fmt.Printf("Formato inválido para clave %s\n", kv.Key)
			continue
		}

		// Extraer y deserializar el Código
		codigoStr := kv.Value[startIndex+8 : mensajeIndex-2]
		var v Valor
		fmt.Sscanf(codigoStr, "%d", &v.Codigo)

		// Extraer y deserializar el Mensaje
		jsonData := kv.Value[mensajeIndex+9:]        // +9 para saltar "Mensaje: {"
		endIndex := strings.LastIndex(jsonData, "}") // Encontrar el último cierre de objeto
		if endIndex != -1 {
			jsonData = jsonData[:endIndex+1] // Mantener hasta el cierre de "}"
		}

		// Añadir una nueva fila al archivo Excel en la hoja 'Informe'
		row := informeSheet.AddRow()
		row.AddCell().SetValue(kv.Key)

		// Verificar si el mensaje contiene solo un campo "Message"
		if strings.Contains(jsonData, `"Message"`) {
			var mensajeSimple map[string]string
			if err := json.Unmarshal([]byte(jsonData), &mensajeSimple); err != nil {
				fmt.Printf("Error al parsear JSON simple para clave %s: %v\n", kv.Key, err)
				continue
			}

			// Escribir el contenido de "Message" en la celda de Mensaje
			row.AddCell().SetValue("N/A")
			row.AddCell().SetValue("N/A")
			row.AddCell().SetValue("N/A")
			row.AddCell().SetValue(mensajeSimple["Message"])

		} else {
			// Deserializar el mensaje completo
			if err := json.Unmarshal([]byte(jsonData), &v.Mensaje); err != nil {
				fmt.Printf("Error al parsear JSON para clave %s: %v\n", kv.Key, err)
				continue
			}

			// Escribir los valores en las celdas correspondientes
			if v.Mensaje.CodigoGeneracion == "" {
				row.AddCell().SetValue("N/A")
			} else {
				row.AddCell().SetValue(v.Mensaje.CodigoGeneracion)
			}
			if v.Mensaje.SelloRecibido == "" {
				row.AddCell().SetValue("N/A")
			} else {
				row.AddCell().SetValue(v.Mensaje.SelloRecibido)
			}
			if v.Mensaje.Estado == "" {
				row.AddCell().SetValue("N/A")
			} else {
				row.AddCell().SetValue(v.Mensaje.Estado)
			}
			if v.Mensaje.DescripcionMsg == "" {
				row.AddCell().SetValue("N/A")
			} else {
				row.AddCell().SetValue(v.Mensaje.DescripcionMsg)
			}
		}

		// Determinar el color de la fila según si el valor indica un acierto o un error
		var color string
		if esAcierto(v) {
			color = "C6EFCE" // verde
		} else {
			color = "FFC7CE" // rojo
		}

		// Aplicar el color a la fila
		for i, cell := range row.Cells {
			style := cell.GetStyle()
			if i == 0 { // Solo aplicar color a la primera celda de cada fila de datos
				style.Fill = *xlsx.NewFill("solid", color, color)
				cell.SetStyle(style)
			}
			style.Font.Name = "Calibri"
			style.Font.Size = 11
		}
	}

	// Copiar todas las hojas del archivo original al nuevo archivo
	for _, sheet := range originalFile.Sheets {
		newSheet, err := newFile.AddSheet(sheet.Name)
		if err != nil {
			return nil, fmt.Errorf("error al añadir la hoja '%s' al nuevo archivo de Excel: %v", sheet.Name, err)
		}
		for _, row := range sheet.Rows {
			newRow := newSheet.AddRow()
			for _, cell := range row.Cells {
				newCell := newRow.AddCell()
				newCell.Value = cell.Value
				style := cell.GetStyle()
				newCell.SetStyle(style)
			}
		}
	}

	// Guardar el nuevo archivo de Excel en un buffer
	var buffer bytes.Buffer
	err = newFile.Write(&buffer)
	if err != nil {
		return nil, fmt.Errorf("error al guardar el archivo de Excel: %v", err)
	}

	return buffer.Bytes(), nil
}

// esAcierto determina si el valor indica un acierto o un error
func esAcierto(v Valor) bool {
	return v.Codigo == 200 && v.Mensaje.CodigoGeneracion != "" && v.Mensaje.SelloRecibido != "" && v.Mensaje.Estado == "PROCESADO"
}
