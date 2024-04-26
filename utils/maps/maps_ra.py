# Description: Maps for REDABIERTA 

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


type_map = {
    "dte": {
        "CodigoGeneracionContingencia": str,
        "NumeroIntentos": int,
        "VentaTercero": bool,
        "NitTercero": str,
        "NombreTercero": str,
        "CodigoCondicionOperacion": str,
        "Rechazado": bool,
        "TipoInvalidacion": str,	
        "CodigoEstablecimientoMH": str,	
        "MotivoInvalidacion": str,
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
        "Observaciones": str,
        "TipoDocIdentResponsable": str,	
        "NumDocIdentResponsable": str,	
        "NombresResponsable": str,
        "TipoDocIdentSolicita": str,	
        "NumDocIdentSolicita": str,	
        "NombresSolicita": str,
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
    },
    "Detalle":{
        "TipoDte": str,
        "CodigoGeneracion": str,	
        "CodigoGeneracionDocumentoReemplazo": str,	
        "TipoDteReemplazo": str,
        "NombreCliente": str,
        "CorreoCliente": str,
        "TelefonoCliente": str,
    }
}