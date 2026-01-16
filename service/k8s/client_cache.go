package k8s

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
)

type cachedClient struct {
	client     *kubernetes.Clientset
	configHash string
	expiresAt  time.Time
}

const defaultClientCacheTTL = 10 * time.Minute

var clientCache sync.Map

func hashKubeConfig(kubeconfig string) string {
	sum := sha256.Sum256([]byte(kubeconfig))
	return hex.EncodeToString(sum[:])
}

func getCachedClient(clusterID uint, configHash string) (*kubernetes.Clientset, bool) {
	value, ok := clientCache.Load(clusterID)
	if !ok {
		return nil, false
	}

	entry, ok := value.(cachedClient)
	if !ok {
		clientCache.Delete(clusterID)
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		clientCache.Delete(clusterID)
		return nil, false
	}

	if entry.configHash != configHash {
		clientCache.Delete(clusterID)
		return nil, false
	}

	return entry.client, true
}

func setCachedClient(clusterID uint, configHash string, client *kubernetes.Clientset) {
	clientCache.Store(clusterID, cachedClient{
		client:     client,
		configHash: configHash,
		expiresAt:  time.Now().Add(defaultClientCacheTTL),
	})
}
