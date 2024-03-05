import time
import numpy as np
import psutil
import os
import sys
import pandas as pd
import json
import random
import string
from datetime import datetime
import csv

def generar_nombre_aleatorio(longitud):
    ahora = datetime.now()
    caracteres = string.ascii_letters + string.digits
    nombre = ''.join(random.choice(caracteres) for _ in range(longitud))
    return nombre + ahora.strftime("%Y%m%d%H%M%S")

def get_memory_usage():
    pid = os.getpid()
    proc = psutil.Process(pid)
    mem_info = proc.memory_info()
    return mem_info.rss / (1024 *  1024)

fc_map = {
    "dte":{
        "CodigoGeneracionContingencia": None,
        "NumeroIntentos": 0,
        "VentaTercero": False,
        "NitTercero": None,
        "NombreTercero": None
    },
    "Identificacion":{
        "TipoDte": "01"
    },
    "Receptor":{
        "Nrc": None
    },
    "Detalles":{
        "Descuento": 0,
        "Codigo": None,
        "CodGenDocRelacionado": None,
        "CodigoTributo": None
    },
    "Resumen":{
        "DescuentoNoSujeto": 0,
        "DescuentoGravado": 0,
        "RetencionRenta": False,
        "DescuentoExento": 0
    },
    "DocumentosRelacionados":[],
    "OtrosDocumentosRelacionados":[],
    "Apendices":[]
}

ccf_map ={
    "dte":{
        "CodigoGeneracionContingencia": None,
        "NumeroIntentos": 0,
        "VentaTercero": False,
        "NitTercero": None,
        "NombreTercero": None
    },
    "Identificacion":{
        "TipoDte": "03"
    },
    "Resumen":{
        "DescuentoNoSujeto": 0,
        "DescuentoGravado":	0,
        "DescuentoExento":	0,
        "RetencionRenta": False
    },
    "DocumentosRelacionados":[],
    "OtrosDocumentosRelacionados":[],
    "Apendices":[]
}

fex_map ={
    "dte":{
        "CodigoGeneracionContingencia": None,
        "NumeroIntentos": 0,
        "VentaTercero": False,
        "NitTercero": None,
        "NombreTercero": None
    },
    "Identificacion":{
        "TipoDte": "11"
    },
    "Resumen": {
        "Seguro": 0.0,
        "Flete": 0.0,
        "CodigoIncoterm": None,
        "DescripcionIncoterm": None,
        "Observaciones": None
    },
    "OtrosDocumentosRelacionados":[],
    "Apendices": []
}

nc_map ={
    "dte":{
        "CodigoGeneracionContingencia": None,
        "NumeroIntentos": 0,
        "VentaTercero": False,
        "NitTercero": None,
        "NombreTercero": None
    },
    "Identificacion":{
        "TipoDte": "05"
    },
    "Resumen":{
        "DescuentoNoSujeto": 0,
        "DescuentoGravado":	0,
        "DescuentoExento":	0,
        "RetencionRenta": False
    },
    "Apendices":[]
}

type_map = {
    "dte": {
        "CodigoGeneracionContingencia": str,
        "NumeroIntentos": int,
        "VentaTercero": bool,
        "NitTercero": str,
        "NombreTercero": str,
        "CodigoCondicionOperacion": str,
    },
    "Identificacion": {
        "TipoDte": str,
        "CodigoEstablecimientoMH": str,
        "Moneda": str
    },
    "Receptor": {
        "TipoDocumentoIdentificacion": str,
        "NumeroDocumentoIdentificacion": str,
        "CodigoDepartamento": str,
        "CodigoMunicipio": str,
        "Direccion": str,    
        "Nrc": str,
        "CodigoActividadEconomica": str,
        "DescripcionActividadEconomica": str,
        "Correo": str,
        "Telefono": str,
        "Nit": str,
        "Nombres": str,
        "CodigoTipoPersona": int,
        "DireccionComplemento": str,
        "CodigoPais": str,
        "NombrePais": str,
    },
    "Detalles": {
        "TipoMonto": int,
        "CodigoTipoItem": int,
        "Cantidad": float,
        "Codigo": str,
        "CodGenDocRelacionado": str,
        "CodigoTributo": str,
        "CodigoUnidadMedida": str,
        "Descripcion": str,
        "Tributos": str,
        "PrecioUnitario": float,
        "IvaItem": float,
        "Descuento": float,
        "Subtotal": float,
    },
    "Resumen": {
        "DescuentoNoSujeto": float,
        "DescuentoGravado":	float,
        "DescuentoExento":	float,
        "RetencionRenta": bool,
        "CodigoRetencionIva": str,
        "PercepcionIva": bool,
        "Seguro": float,
        "Flete": float,
        "CodigoIncoterm": str,
        "DescripcionIncoterm": str,
        "Observaciones": str
    },
    "Extension": {
        "NombreEntrega": str,
        "DocumentoEntrega": str,
        "NombreRecibe": str,
        "DocumentoRecibe": str,
        "Observaciones": str,
        "PlacaVehiculo": str
    },
    "DocumentosRelacionados": {
        "TipoDte": str,
        "CodigoGeneracion": str,
        "CodigoTipoGeneracion": int,
        "FechaEmision": str
    }
    # "OtrosDocumentosRelacionados": [],
    # "Apendices": []
}

map_columns ={
    "dte": {
        "CodigoCondicionOperacion": "CodigoCondicionOperacion",
    },
    "Identificacion": {
        "CodigoEstablecimientoMH": "CodigoEstablecimientoMH",
    },
    "Receptor": {
        "TipoDocumentoIdentificacion": "TipoDocumentoIdentificacion",
        "NumeroDocumentoIdentificacion": "NumeroDocumentoIdentificacion",
        "CodigoDepartamento": "CodigoDepartamento",
        "CodigoMunicipio": "CodigoMunicipio",
        "CodigoActividadEconomica": "CodigoActividadEconomica",
        "DescripcionActividadEconomica": "DescripcionActividadEconomica",
    },
    "Detalles": {
        "TipoMonto": "TipoMonto",
        "CodigoTipoItem": "CodigoTipoItem",
        "CodGenDocRelacionado": "CodGenDocRelacionado",
        "CodigoTributo": "CodigoTributo",
        "CodigoUnidadMedida": "CodigoUnidadMedida",
        "PrecioUnitario": "PrecioUnitario",
        "IvaItem": "IvaItem",
    },
    "Resumen": {
        "CodigoRetencionIva": "CodigoRetencionIva",
    },
    "Extension": {
        "NombreEntrega": "NombreEntrega",
        "DocumentoEntrega": "DocumentoEntrega",
        "NombreRecibe": "NombreRecibe",
        "DocumentoRecibe": "DocumentoRecibe",
        "Observaciones": "Observaciones",
        "PlacaVehiculo": "PlacaVehiculo"
    },
    "DocumentosRelacionados": {
        "TipoDte": "TipoDte",
        "CodigoGeneracion": "CodigoGeneracion",
        "CodigoTipoGeneracion": "CodigoTipoGeneracion",
        "FechaEmision": "FechaEmision"
    }
}


def main():
    before_memory_usage = get_memory_usage()
    start_cpu_usage = psutil.cpu_percent(interval=1)
    start_time = time.time()

    archivo_excel = sys.argv[1]
    hojas = pd.read_excel(archivo_excel, sheet_name=None)
    hojas_a_procesar = list(hojas.keys())  # Obtener automáticamente los nombres de las hojas


    map_selected = fc_map
    map_datatype_selected = type_map
    map_columns_selected = map_columns

    detalles_por_id = {}
    message = [['IDDTE', 'ERROR', 'FECHA', 'STATUS']]
    

    # Procesamiento de todas las hojas
    for hoja_nombre in hojas_a_procesar:
        hoja = hojas.get(hoja_nombre)
        if hoja is None:
            print(f"La hoja '{hoja_nombre}' no existe en el archivo Excel.")
            continue

        # Especificar columnas que deben ser tratadas como cadenas de texto al cargar el Excel para esta hoja según el mapa de tipos de datos
        if hoja_nombre in map_datatype_selected:
            datatype_map = map_datatype_selected[hoja_nombre]
            columnas_texto = [col for col, dtype in datatype_map.items() if dtype == str]
        else:
            columnas_texto = []

        # Crear un diccionario con los tipos de datos para cada columna
        dtype_dict = {col: str if col in columnas_texto else None for col in hoja.columns}

        # Cargar el archivo Excel forzando la interpretación de ciertas columnas como cadenas de texto
        hoja = pd.read_excel(archivo_excel, sheet_name=hoja_nombre, dtype=dtype_dict)

        hoja = hoja.rename(columns=map_columns_selected.get(hoja_nombre, {}))  # Renombrar las columnas según el mapa de la hoja

        for index, row in hoja.iterrows():
            try:
                idte = row['IDDTE']
                detalle = detalles_por_id.get(idte, {})
                if not detalle:
                    detalles_por_id[idte] = {}

                if hoja_nombre == hojas_a_procesar[0]:  # Procesar la primera hoja
                    for col, val in row.items():
                        if col != 'IDDTE':
                            detalles_por_id[idte][col] = val if not pd.isna(val) else None
                else:
                    if hoja_nombre not in detalles_por_id[idte]:
                        detalles_por_id[idte][hoja_nombre] = [] 
                        
                    # Procesar la columna "Tributos" específicamente
                    if "Tributos" in row.index:
                        tributos_value = row["Tributos"]
                        # Convertir el valor de "Tributos" a una lista con un único elemento
                        tributos_list = [str(tributos_value)] if not pd.isna(tributos_value) else []
                        row["Tributos"] = tributos_list     
                        

                    detalles_por_id[idte][hoja_nombre].append(row.drop(labels=['IDDTE']).to_dict())

            except Exception as e:
                error_line = sys.exc_info()[2].tb_lineno
                print(f"Hubo un error: {e}, en la linea: {error_line}")
                ahora = datetime.now()
                fecha_hora = ahora.strftime("%Y-%m-%d %H:%M:%S")
                message.append([hoja.iloc[index]['IDDTE'], f"Hubo un error: {e}, en la linea: {error_line}", fecha_hora, "Error"])

    # Integrar fc_map en detalles_por_id
    for idte, detalle in detalles_por_id.items():
        for hoja_nombre, datos_fijos in map_selected.items():
            if hoja_nombre != "dte":
                if idte not in detalles_por_id:
                    detalles_por_id[idte] = {}

                if hoja_nombre not in detalles_por_id[idte]:
                    if isinstance(datos_fijos, dict):
                        detalles_por_id[idte][hoja_nombre] = [datos_fijos.copy()]  # Añadir datos fijos como lista
                    elif isinstance(datos_fijos, list):
                        detalles_por_id[idte][hoja_nombre] = [fijo.copy() for fijo in datos_fijos]  # Añadir datos fijos como lista
                else:
                    if isinstance(datos_fijos, dict):
                        for item in detalles_por_id[idte][hoja_nombre]:
                            item.update(datos_fijos.copy())
                    elif isinstance(datos_fijos, list):
                        for fijo in datos_fijos:
                            detalles_por_id[idte][hoja_nombre].append(fijo.copy())
                            
    # Convertir la hoja en objeto si tiene solo una fila asociada
    for idte, detalle in detalles_por_id.items():
        for hoja_nombre, data in detalle.items():
            if hoja_nombre in ["Detalles", "DocumentosRelacionados"]:  # Verificar si es "Detalles" o "DocumentosRelacionados"
                # Asegurar que la hoja siempre sea una lista
                if not isinstance(data, list):
                    detalles_por_id[idte][hoja_nombre] = [data]
            elif isinstance(data, list) and len(data) == 1:  # Convertir a lista si solo hay un objeto
                detalles_por_id[idte][hoja_nombre] = data[0]

    for idte in detalles_por_id:
        message.append([idte, '', '', 'SUCCESS'])

    # Convertir las claves numpy.int64 a str
    detalles_por_id_str_keys = {str(key): value for key, value in detalles_por_id.items()}


    # Convertir NaN a None para representarlos como null en el JSON
    for idte, detalle in detalles_por_id_str_keys.items():
        for key, value in detalle.items():
            if isinstance(value, dict):
                for k, v in value.items():
                    if pd.isna(v):
                        value[k] = None
            elif isinstance(value, list):
                for record in value:
                    for k, v in record.items():
                        if pd.isna(v):
                            record[k] = None

    # Agregar datos del objeto "dte" directamente en la raíz
    for idte, detalle in detalles_por_id_str_keys.items():
        if "dte" in map_selected:
            dte_data = map_selected["dte"]
            for key, value in dte_data.items():
                if key not in detalle:
                    detalles_por_id_str_keys[idte][key] = value

    # Convertir tipos de datos según el mapa map_datatype_selected
        for idte, detalle in detalles_por_id_str_keys.items():
            for hoja_nombre, data in detalle.items():
                if hoja_nombre in map_datatype_selected:
                    datatype_map = map_datatype_selected[hoja_nombre]
                    for key, datatype in datatype_map.items():
                        if isinstance(data, list):
                            for record in data:
                                if key in record:
                                    if datatype == str and isinstance(record[key], (int, float)):
                                        # Convertir a cadena manteniendo el formato con ceros a la izquierda
                                        record[key] = "{:02d}".format(record[key])
                        else:
                            if key in data:
                                if datatype == str and isinstance(data[key], (int, float)):
                                    # Convertir a cadena manteniendo el formato con ceros a la izquierda
                                    data[key] = "{:02d}".format(data[key])

    # Generar un único archivo JSON al final del procesamiento
    nombre_json = generar_nombre_aleatorio(10) + '.json'
    json_data = json.dumps(detalles_por_id_str_keys, default=lambda x: x if x is not pd.NA else None)
    with open(nombre_json, "w") as json_file:
        json_file.write(json_data)

    ahora = datetime.now()
    fecha_hora = ahora.strftime("%Y%m%d%H%M%S")
    with open('messages' + fecha_hora + '.csv', mode='a', newline='') as file:
        writer = csv.writer(file)
        for e in message:
            writer.writerow(e)

    elapsed_time = time.time() - start_time
    end_cpu_usage = psutil.cpu_percent(interval=1)
    cpu_usage_difference = end_cpu_usage - start_cpu_usage
    after_memory_usage = get_memory_usage()
    memory_usage_diff = after_memory_usage - before_memory_usage

    print(f"Elapsed time: {elapsed_time} seconds")
    print(f"CPU Usage Difference: {cpu_usage_difference}%")
    print(f"Memory usage difference: {memory_usage_diff:.2f} MB")

if __name__ == "__main__":
    main()
