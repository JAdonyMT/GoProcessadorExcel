import numpy as np
import psutil
import os
import sys
import pandas as pd
import json
from datetime import datetime
import csv

def convert_nan_to_none(value):
    if isinstance(value, (float, np.float64)):
        return None if np.isnan(value) else value
    elif isinstance(value, (list, tuple, np.ndarray)):
        return [convert_nan_to_none(v) for v in value]
    elif isinstance(value, dict):
        return {k: convert_nan_to_none(v) for k, v in value.items()}
    else:
        return value

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
        "NombreTercero": None,
        "Rechazado": False
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
        "Rechazado": bool
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
    message = [['IDDTE', 'ERROR', 'FECHA', 'STATUS']]
    
    archivo_excel = sys.argv[1]
    tipo_dte = sys.argv[2]
    try:
        hojas = pd.read_excel(archivo_excel, sheet_name=None)
    except Exception as e:
        message.append(['', f"Error al cargar el archivo Excel: {e}", '', "Error"])
        print(f"Error al cargar el archivo Excel: {e}")
        sys.exit(1)
    hojas_a_procesar = list(hojas.keys())  # Obtener automáticamente los nombres de las hojas

    map_dte = {
        "01": fc_map,
        "03": ccf_map,
        "11": fex_map,
        "05": nc_map
    }

    map_selected = map_dte.get(tipo_dte)
    map_datatype_selected = type_map
    map_columns_selected = map_columns

    detalles_por_id = {}
    
    dtype_dict_per_sheet = {}
    for hoja_nombre in hojas_a_procesar:
        hoja = hojas.get(hoja_nombre)
        if hoja is not None:
            if hoja_nombre in map_datatype_selected:
                datatype_map = map_datatype_selected[hoja_nombre]
                columnas_texto = [col for col, dtype in datatype_map.items() if dtype == str]
            else:
                columnas_texto = []
            dtype_dict_per_sheet[hoja_nombre] = {col: str if col in columnas_texto else None for col in hoja.columns}        

    # Procesamiento de todas las hojas
    for hoja_nombre in hojas_a_procesar:
        try:
            hoja = hojas.get(hoja_nombre)
            if hoja is None:
                message.append(['', f"La hoja '{hoja_nombre}' no existe en el archivo Excel.", '', "Error"])
                print(f"La hoja '{hoja_nombre}' no existe en el archivo Excel.")
                continue
            
            # Verificar si la hoja está vacía
            if hoja.empty:
                message.append(['', f"La hoja '{hoja_nombre}' está vacia.", '', "Error"])
                print(f"La hoja '{hoja_nombre}' está vacia.")
                continue

            # Especificar columnas que deben ser tratadas como cadenas de texto al cargar el Excel para esta hoja según el mapa de tipos de datos
            if hoja_nombre in map_datatype_selected:
                datatype_map = map_datatype_selected[hoja_nombre]
                columnas_texto = [col for col, dtype in datatype_map.items() if dtype == str]
            else:
                columnas_texto = []

            # Aplicar los tipos de datos al cargar la hoja
            hoja = pd.read_excel(archivo_excel, sheet_name=hoja_nombre, dtype=dtype_dict_per_sheet[hoja_nombre])

            try:
                hoja = hoja.rename(columns=map_columns_selected.get(hoja_nombre, {}))  # Renombrar las columnas según el mapa de la hoja
            except Exception as e:
                message.append(['', f"Error al renombrar las columnas de la hoja '{hoja_nombre}': {e}", '', "Error"])
                print(f"Error al renombrar las columnas de la hoja '{hoja_nombre}': {e}")
                continue
            
            for index, row in hoja.iterrows():  
                try:
                    idte = row['IDDTE']
                    if pd.isnull(idte) or idte == '':
                        print(f"La columna 'IDDTE' no puede estar vacia. Hoja: {hoja_nombre}, fila {index + 2}")
                        raise ValueError("La columna 'IDDTE' no puede estar vacia.")
                    
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
                            # Verificar si hay múltiples tributos separados por comas
                            if "," in str(tributos_value):
                                # Dividir los tributos por comas y eliminar espacios en blanco
                                tributos_list = [t.strip() for t in tributos_value.split(",")]
                            else:
                                # Si solo hay un tributo, colocarlo en una lista
                                tributos_list = [str(tributos_value)] if not pd.isna(tributos_value) else []
                            row["Tributos"] = tributos_list
                            

                        detalles_por_id[idte][hoja_nombre].append(row.drop(labels=['IDDTE']).to_dict())

                except Exception as e:
                    columna_error = hoja.columns[index]
                    error_info = f"Error en la columna '{columna_error}': {e}"
                    ahora = datetime.now()
                    fecha_hora = ahora.strftime("%Y-%m-%d %H:%M:%S")
                    message.append([hoja.iloc[index]['IDDTE'], error_info, fecha_hora, "Error"])

        except Exception as e:
            message.append(['', f"Error al procesar la hoja '{hoja_nombre}': {e}", '', "Error"])
            print(f"Error al procesar la hoja '{hoja_nombre}' en la fila {index + 2}")
            continue
    
    # Integrar map_selected en detalles_por_id
    for idte, detalle in detalles_por_id.items():
        try:
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
        except Exception as e:
            message.append([idte, f"Error al integrar map_selected en detalles_por_id para el IDDTE '{idte}' y la hoja '{hoja_nombre}': {e}", '', "Error"])
            print(f"Error al integrar map_selected en detalles_por_id para el IDDTE '{idte}' y la hoja '{hoja_nombre}': {e}")
                                
    # Convertir la hoja en objeto si tiene solo una fila asociada
    for idte, detalle in detalles_por_id.items():
        try:
            for hoja_nombre, data in detalle.items():
                if hoja_nombre in ["Detalles", "DocumentosRelacionados"]:  # Verificar si es "Detalles" o "DocumentosRelacionados"
                    # Asegurar que la hoja siempre sea una lista
                    if not isinstance(data, list):
                        detalles_por_id[idte][hoja_nombre] = [data]
                elif isinstance(data, list) and len(data) == 1:  # Convertir a lista si solo hay un objeto
                    detalles_por_id[idte][hoja_nombre] = data[0]
        except Exception as e:
            message.append([idte, f"Error al convertir la hoja '{hoja_nombre}' en objeto para el IDDTE '{idte}': {e}", '', "Error"])
            print(f"Error al convertir la hoja '{hoja_nombre}' en objeto para el IDDTE '{idte}': {e}")


    for idte in detalles_por_id:
        try:
            message.append([idte, '', '', 'SUCCESS'])
        except Exception as e:
            message.append([idte, f"Error al agregar el mensaje de exito para el IDDTE '{idte}': {e}", '', "Error"])
            print(f"Error al agregar el mensaje de exito para el IDDTE '{idte}': {e}")

        # Convertir las claves numpy.int64 a str
    try:      
        detalles_por_id_str_keys = {str(key): value for key, value in detalles_por_id.items()}
    except Exception as e:
        message.append(['', f"Error al convertir las claves a cadena de texto: {e}", '', "Error"])
        print(f"Error al convertir las claves a cadena de texto: {e}")

    # Convertir NaN a None para representarlos como null en el JSON
    for idte, detalle in detalles_por_id_str_keys.items():
        try:
            detalles_por_id_str_keys[idte] = convert_nan_to_none(detalle)
        except Exception as e:
            message.append([idte, f"Error al convertir valores a null para el IDDTE '{idte}': {e}", '', "Error"])
            print(f"Error al convertir valores a null para el IDDTE '{idte}': {e}")

                
    # Agregar datos del objeto "dte" directamente en la raíz
    for idte, detalle in detalles_por_id_str_keys.items():
        try:
            if "dte" in map_selected:
                dte_data = map_selected["dte"]
                for key, value in dte_data.items():
                    if key not in detalle:
                        detalles_por_id_str_keys[idte][key] = value
        except Exception as e:
            message.append([idte, f"Error al agregar datos del objeto 'dte' a la raiz para el IDDTE '{idte}': {e}", '', "Error"])
            print(f"Error al agregar datos del objeto 'dte' a la raiz para el IDDTE '{idte}': {e}")
            
        # Convertir tipos de datos según el mapa map_datatype_selected
        try:
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
        except Exception as e:
            message.append([idte, f"Error al convertir tipos de datos para el IDDTE '{idte}': {e}", '', "Error"])
            print(f"Error al convertir tipos de datos para el IDDTE '{idte}': {e}")


    nombre_archivo = os.path.splitext(os.path.basename(archivo_excel))[0]
    
    # Generar un único archivo JSON al final del procesamiento
    nombre_json = nombre_archivo + '.json'
    json_data = json.dumps(detalles_por_id_str_keys, default=lambda x: x if x is not pd.NA else None)
    with open(nombre_json, "w") as json_file:
        json_file.write(json_data)

    ahora = datetime.now()
    fecha_hora = ahora.strftime("%Y%m%d%H%M%S")
    with open(nombre_archivo + fecha_hora + '.csv', mode='a', newline='') as file:
        writer = csv.writer(file)
        for e in message:
            writer.writerow(e)

if __name__ == "__main__":
    main()
