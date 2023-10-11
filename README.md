# chain_block

## Integrantes

- Cristian Villavicencio
- David Pazan
- Sebastian Gonzalez

## Funcionalidad

El siguiente repositorio consta de la elaboracion de un sistema de blockchain, en primera instancia solo se modela la capa de datos, para ello se utilizo golang y leveldb.

### database

En este se encuentra la conexion asi como los metodos get, put, delete y tambien la inicializacion y el termino de la conexion a la misma.

### src

En este se encuentra el core del proyecto, en donde se encuentra el bloque, las transacciones y los metodos de generacion de ellos, como el calculo de hash, guardar y cargar bloques.

### main

En este se encuentra el menu interactivo para lograr probar las funcionalidades, este mismo tiene una forma descriptiva de cada metodo.

### Ejecutar

Para ejecutar este mismo se recomienda usar docker que generar el archivo binario de golang para la posterior ejecucion del mismo, aun asi se puede ejecutar de forma casera, utilizando los comandos de go run o go build.
