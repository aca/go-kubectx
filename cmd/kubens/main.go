package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aca/kubectx/fzfutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var query string
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], "")
	}

	err := kubens(query)
	if err != nil {
		fmt.Printf("Err: %s", err)
		os.Exit(1)
	}
}

// kubens fuzzy finds & switches namespaces
func kubens(query string) error {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeconfig, err := ioutil.ReadFile(rules.GetDefaultFilename())
	if err != nil {
		return err
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	restConfig.Timeout = time.Second * 5

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	fzfopt := []string{"--select-1", "--query", query}

	output, err := fzfutil.FZF(func(in io.WriteCloser) {
		namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		inputChan := make(chan string, 100)
		for _, item := range namespaces.Items {
			fmt.Fprintln(in, item.ObjectMeta.Name)
		}
		close(inputChan)

	}, fzfopt...)

	if err != nil {
		return err
	}

	if len(output) == 0 {
		return fmt.Errorf("namespace not selected")
	}

	clientcmdConfig, err := rules.Load()
	if err != nil {
		return err
	}

	clientcmdConfig.Contexts[clientcmdConfig.CurrentContext].Namespace = output[0]
	return clientcmd.ModifyConfig(rules, *clientcmdConfig, false)
}
