package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
)

type NacosRegistration struct {
	client naming_client.INamingClient
	param  vo.RegisterInstanceParam
}

func parseNacosServerConfigs(v *viper.Viper) ([]constant.ServerConfig, error) {
	serverAddrs := strings.TrimSpace(v.GetString("nacos.server_addrs"))
	defaultPort := v.GetUint64("nacos.port")
	if defaultPort == 0 {
		defaultPort = 8848
	}
	if serverAddrs == "" {
		return []constant.ServerConfig{
			*constant.NewServerConfig(v.GetString("nacos.host"), defaultPort, constant.WithContextPath("/nacos")),
		}, nil
	}

	items := strings.Split(serverAddrs, ",")
	serverConfigs := make([]constant.ServerConfig, 0, len(items))
	for _, item := range items {
		addr := strings.TrimSpace(item)
		if addr == "" {
			continue
		}
		host, portText, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
			portText = strconv.FormatUint(defaultPort, 10)
		}
		port, err := strconv.ParseUint(portText, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid nacos server addr %q: %w", addr, err)
		}
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(host, port, constant.WithContextPath("/nacos")))
	}
	if len(serverConfigs) == 0 {
		return nil, fmt.Errorf("nacos server_addrs is empty")
	}
	return serverConfigs, nil
}

func buildNacosClientConfig(v *viper.Viper) constant.ClientConfig {
	timeout := v.GetUint64("nacos.timeout_ms")
	if timeout == 0 {
		timeout = 5000
	}
	options := []constant.ClientOption{
		constant.WithNamespaceId(v.GetString("nacos.namespace")),
		constant.WithTimeoutMs(timeout),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("error"),
	}
	username := strings.TrimSpace(v.GetString("nacos.username"))
	password := strings.TrimSpace(v.GetString("nacos.password"))
	if username != "" {
		options = append(options, constant.WithUsername(username))
	}
	if password != "" {
		options = append(options, constant.WithPassword(password))
	}
	return *constant.NewClientConfig(options...)
}

func detectLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if ip != nil {
			return ip.String()
		}
	}
	return "127.0.0.1"
}

func RegisterToNacos(v *viper.Viper) (*NacosRegistration, error) {
	if !v.GetBool("nacos.enable") || !v.GetBool("nacos.register_enable") {
		return nil, nil
	}

	serverConfigs, err := parseNacosServerConfigs(v)
	if err != nil {
		return nil, err
	}
	clientConfig := buildNacosClientConfig(v)

	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("create nacos naming client failed: %w", err)
	}

	serviceName := strings.TrimSpace(v.GetString("nacos.service_name"))
	if serviceName == "" {
		serviceName = "devops-backend"
	}
	ip := strings.TrimSpace(v.GetString("nacos.service_ip"))
	if ip == "" {
		ip = detectLocalIP()
	}
	port := v.GetUint64("nacos.service_port")
	if port == 0 {
		port = uint64(v.GetInt("server.port"))
	}
	if port == 0 {
		port = 8000
	}

	groupName := strings.TrimSpace(v.GetString("nacos.service_group"))
	if groupName == "" {
		groupName = strings.TrimSpace(v.GetString("nacos.group"))
		if groupName == "" {
			groupName = "DEFAULT_GROUP"
		}
	}
	clusterName := strings.TrimSpace(v.GetString("nacos.service_cluster"))
	if clusterName == "" {
		clusterName = "DEFAULT"
	}
	weight := v.GetFloat64("nacos.service_weight")
	if weight <= 0 {
		weight = 1
	}

	metadata := v.GetStringMapString("nacos.service_metadata")
	if metadata == nil {
		metadata = map[string]string{}
	}
	if _, ok := metadata["protocol"]; !ok {
		metadata["protocol"] = "http"
	}
	if _, ok := metadata["version"]; !ok {
		metadata["version"] = "v1"
	}

	param := vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		Weight:      weight,
		Enable:      true,
		Healthy:     true,
		Metadata:    metadata,
		ClusterName: clusterName,
		ServiceName: serviceName,
		GroupName:   groupName,
		Ephemeral:   v.GetBool("nacos.service_ephemeral"),
	}

	ok, err := client.RegisterInstance(param)
	if err != nil {
		return nil, fmt.Errorf("register nacos instance failed: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("register nacos instance failed: response not ok")
	}

	return &NacosRegistration{client: client, param: param}, nil
}

func (r *NacosRegistration) Deregister() error {
	if r == nil || r.client == nil {
		return nil
	}
	ok, err := r.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          r.param.Ip,
		Port:        r.param.Port,
		Cluster:     r.param.ClusterName,
		ServiceName: r.param.ServiceName,
		GroupName:   r.param.GroupName,
		Ephemeral:   r.param.Ephemeral,
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("deregister nacos instance failed: response not ok")
	}
	return nil
}
