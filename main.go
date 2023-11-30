package main

import (
    "blockchain/database"
    "blockchain/core"
)

func main() {
    db := database.InitDB("blockchain.db")
    block := core.GenerateBlock(0, "", []core.Transaction{}, 0)
    
    core.SaveBlock(core.GetDB(), &block)
}

// generar network
// hacer descubrimiento de red
// hacer syncronizacion de red
// en caso de no existir nodos, se debe hacer genesis del primer bloque
// caso contrario se debe hacer syncronizacion de bloques
// en cada transaccion se debe hacer broadcast de la transaccion
// en cada bloque se debe hacer broadcast del bloque
// en cada bloque se debe hacer broadcast de la blockchain
// buscar la forma del doble gasto
// escribir en base de datos solo si se tiene el 51% de la red