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

def correct_names(s): 
    translator = str.maketrans("áéíóúÁÉÍÓÚñÑ", "aeiouAEIOUnN")
    return s.translate(translator).lower().capitalize()

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

def main():
    before_memory_usage = get_memory_usage()
    start_cpu_usage = psutil.cpu_percent(interval=1)
    start_time = time.time()

    # hojas_a_procesar = ["dte", "Identificacion", "Receptor", 
    #                     "Documentos Relacionados", "Extension", 
    #                     "Resumen", "Detalles", "Apendices"]

    archivo_excel = sys.argv[1]
    hojas = pd.read_excel(archivo_excel, sheet_name=None)
    hojas_a_procesar = list(hojas.keys())  # Obtener automáticamente los nombres de las hojas


    detalles_por_id = {}
    errores = [['IDDTE', 'ERROR', 'FECHA']]

    for hoja_nombre in hojas_a_procesar:
        hoja = hojas.get(hoja_nombre)
        if hoja is None:
            print(f"La hoja '{hoja_nombre}' no existe en el archivo Excel.")
            continue

        hoja = hoja.rename(columns=correct_names)

        for index, row in hoja.iterrows():
            try:
                idte = row['Iddte']
                detalle = detalles_por_id.get(idte, {})
                if not detalle:
                    detalles_por_id[idte] = {}

                if hoja_nombre == hojas_a_procesar[0]:  # Procesar la primera hoja
                    for col, val in row.items():
                        if col != 'Iddte':
                            detalles_por_id[idte][col] = val if not pd.isna(val) else None
                else:
                    if hoja_nombre not in detalles_por_id[idte]:
                        detalles_por_id[idte][hoja_nombre] = []
                    detalles_por_id[idte][hoja_nombre].append(row.drop(labels=['Iddte']).to_dict())
            except Exception as e:
                error_line = sys.exc_info()[2].tb_lineno
                print(f"Hubo un error: {e}, en la linea: {error_line}")
                ahora = datetime.now()
                fecha_hora = ahora.strftime("%Y-%m-%d %H:%M:%S")
                errores.append([hoja.iloc[index]['Iddte'], f"Hubo un error: {e}, en la linea: {error_line}", fecha_hora])

    # Convertir NaN a None para representarlos como null en el JSON
    for idte, detalle in detalles_por_id.items():
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

        nombre_json = generar_nombre_aleatorio(10) + '.json'  # Generar un nombre aleatorio para el archivo JSON
        detalles_por_id = {key: {inner_key: inner_value[0] if isinstance(inner_value, list) and len(inner_value) == 1 else inner_value for inner_key, inner_value in value.items()} for key, value in detalles_por_id.items()}
        json_data = json.dumps(detalles_por_id, default=lambda x: x if x is not pd.NA else None)
        with open(nombre_json, "w") as json_file:
            json_file.write(json_data)

    ahora = datetime.now()
    fecha_hora = ahora.strftime("%Y%m%d%H%M%S")
    with open('errores' + fecha_hora + '.csv', mode='a', newline='') as file:
        writer = csv.writer(file)
        for e in errores:
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
