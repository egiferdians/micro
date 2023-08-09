package configetcd

import (
	"time"

	"github.com/egiferdians/micro/util/flags"

	"context"
	"log"
	"strings"

	"go.etcd.io/etcd/clientv3"
)

type ConfigFormat map[string]string
type ETCDresponder func(nodename string, updatedinfo ConfigFormat)

func ETCDConnectAndListen(etcdHost []string, servicenode string, resp ETCDresponder) (res ConfigFormat, err error) {

	cliConfig := clientv3.Config{
		Endpoints:   etcdHost,
		DialTimeout: 5 * time.Second,
	}
	cli, err := clientv3.New(cliConfig)
	if err != nil {
		log.Println(err)
	}
	defer cli.Close()
	prefixPath := servicenode + flags.ETCD_GLOBALS_CONFIG_PATH
	etcdData, err := cli.Get(context.TODO(), prefixPath)
	if err != nil {
		log.Println(err)
	}
	if etcdData.Count < 1 {
		log.Printf("node missing %s\n", prefixPath)
		return nil, nil
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	res = make(ConfigFormat)
	kv, err := cli.Get(context.Background(), prefixPath, clientv3.WithPrefix())
	for i := 0; i < len(kv.Kvs); i++ {
		key := strings.Replace(string(kv.Kvs[i].Key), prefixPath, "", -1)
		res[key] = string(kv.Kvs[i].Value)
	}

	return res, err
}
