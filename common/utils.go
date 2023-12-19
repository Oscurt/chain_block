package common

import (
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"

    "github.com/joho/godotenv"
)

const SeedNodesEnvVar = "LIBP2P_SEEDNODES"

// ChooseRandomNodeFromEnv selecciona un nodo aleatorio de los nodos semilla en el archivo .env, excluyendo el nodo especificado si se proporciona.
func ChooseRandomNodeFromEnv(excludeNodeID string) (string, error) {
    err := godotenv.Load()
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
        for , node := range nodes {
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