package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aca/go-kubectx/config"
	"github.com/aca/go-kubectx/fzfutil"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var cfg *config.Config
var rules *clientcmd.ClientConfigLoadingRules
var clientcmdCfg *clientcmdapi.Config
var kubeCfgBytes []byte

func main() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg, err = config.ReadCfg()
	if err != nil {
		log.Fatal(err)
	}

	fs := pflag.NewFlagSet("kubens", pflag.ContinueOnError)
	current := fs.BoolP("current", "c", false, "show the current context")
	help := fs.BoolP("help", "h", false, "help")
	fs.Parse(os.Args[1:])

	if *help {
		fmt.Printf(`USAGE:
  kubens                : switch namespace(interactive)
  kubens <QUERY>        : switch namespace which closest match with QUERY(if failed to match, interactive mode)
  kubens -              : switch to previous namespace (~/.config/kubectx/config.json)
  kubens -c, --current  : show current namespace name
  kubens -h,--help      : help
`)
		return
	}

	rules = clientcmd.NewDefaultClientConfigLoadingRules()
	kubeCfgBytes, err = ioutil.ReadFile(rules.GetDefaultFilename())
	if err != nil {
		log.Fatal(err)
	}

	clientcmdCfg, err = clientcmd.Load(kubeCfgBytes)
	if err != nil {
		log.Fatal(err)
	}

	var query string

	if *current {
		fmt.Println(clientcmdCfg.Contexts[clientcmdCfg.CurrentContext].Namespace)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "-" {
		query = cfg.LastNamespace
	} else if len(os.Args) > 1 {
		query = os.Args[1]
	}

	err = kubens(query)
	if err != nil {
		fmt.Printf("Err: %s", err)
		os.Exit(1)
	}
}

func kubens(query string) (err error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeCfgBytes)
	if err != nil {
		return err
	}
	restConfig.Timeout = time.Second * 5

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range namespaces.Items{
		if item.ObjectMeta.Name == query {
			return modifyConfig(query)
		}
	}

	fzfopt := []string{"--select-1", "--query", query}

	result, err := fzfutil.FZF(func(in io.WriteCloser) {
		for _, item := range namespaces.Items {
			fmt.Fprintln(in, item.ObjectMeta.Name)
		}
	}, fzfopt...)

	if err != nil {
		return err
	}

	if len(result) == 0 {
		return errors.New("no namespace selected")
	}

	return modifyConfig(result[0])
}

func modifyConfig(to string) (err error) {
	curNS := clientcmdCfg.Contexts[clientcmdCfg.CurrentContext].Namespace
	defer func() {
		if err == nil {
			if curNS != to {
				cfg.LastNamespace = curNS
				config.WriteCfg(cfg)
			}
			fmt.Printf("Switched to namespace \"%s\"\n", to)
		}
	}()
	clientcmdCfg.Contexts[clientcmdCfg.CurrentContext].Namespace = to
	return clientcmd.ModifyConfig(rules, *clientcmdCfg, false)
}
