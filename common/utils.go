package common

import (
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"
    "crypto/sha256"
    "encoding/hex"
    "strconv"
    "github.com/joho/godotenv"
)

const SeedNodesEnvVar = "LIBP2P_SEED_NODES"

// ChooseRandomNodeFromEnv selecciona un nodo aleatorio de los nodos semilla en el archivo .env, excluyendo el nodo especificado si se proporciona.
func ChooseRandomNodeFromEnv(excludeNodeID string) (string, error) {

    err := godotenv.Load(".env")
    if err != nil {
        return "", fmt.Errorf("error al cargar el archivo .env: %v", err)
    }

    seedNodes := os.Getenv(SeedNodesEnvVar)



    nodes := strings.Split(seedNodes, ",")
    if len(nodes) == 0 {
        return "", fmt.Errorf("no se encontraron nodos semilla")
    }

    // Filtrar el nodo excluido si se especifica
    if excludeNodeID != "" {
        var filteredNodes []string
        for _ , node := range nodes {
            if !strings.HasSuffix(node, excludeNodeID) {
                filteredNodes = append(filteredNodes, node)
            }
        }
        nodes = filteredNodes
    }

    if len(nodes) == 0 {
        return "", fmt.Errorf("no hay nodos semilla disponibles después de excluir el nodo especificado")
    }

    rand.Seed(time.Now().UnixNano())
    randomNode := nodes[rand.Intn(len(nodes))]
    return randomNode, nil
}

func GenerateTransactionHash(transaction Transaction) string {
    // Concatenar los campos de la transacción para formar una cadena única
    data := transaction.Sender + transaction.Recipient + fmt.Sprintf("%f", transaction.Ammount) + transaction.Signature + strconv.FormatInt(transaction.TimeStamp, 10)

    // Calcular el hash SHA-256 de la cadena
    hash := sha256.Sum256([]byte(data))

    // Convertir el hash a una cadena hexadecimal
    return hex.EncodeToString(hash[:])
}