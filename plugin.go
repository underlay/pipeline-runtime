package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	core "github.com/ipfs/go-ipfs/core"
	plugin "github.com/ipfs/go-ipfs/plugin"
	// coreiface "github.com/ipfs/interface-go-ipfs-core"
)

type pipelinePlugin struct {
	moduleDirectory string
	ipfs            *core.IpfsNode
	// api             coreiface.CoreAPI
}

// Compile-time type check
var _ plugin.PluginDaemonInternal = (*pipelinePlugin)(nil)

// Name returns the plugin's name, satisfying the plugin.Plugin interface.
func (*pipelinePlugin) Name() string {
	return "pipeline-workflow"
}

// Version returns the plugin's version, satisfying the plugin.Plugin interface.
func (*pipelinePlugin) Version() string {
	return "0.1.0"
}

// Init initializes plugin, satisfying the plugin.Plugin interface. Put any
// initialization logic here.
func (p *pipelinePlugin) Init(env *plugin.Environment) error {
	moduleDirectory, err := filepath.Abs("node_modules")
	if err != nil {
		return err
	}

	p.moduleDirectory = moduleDirectory

	return nil
}

func (p *pipelinePlugin) Start(node *core.IpfsNode) (err error) {
	fmt.Println("Hello!")
	log.Println("http://localhost:8086")

	p.ipfs = node
	// p.api, err = coreapi.NewCoreAPI(p.ipfs)
	// if err != nil {
	// 	return
	// }

	// streamInfo := &p2p.Stream{}
	// p.ipfs.P2P.Streams.Register(streamInfo)

	go log.Fatal(http.ListenAndServe(":8086", p))

	return
}

func (*pipelinePlugin) Close() error {
	fmt.Println("Goodbye!")
	return nil
}

// Plugins is the exported global variable that the IPFS daemon looks for
var Plugins = []plugin.Plugin{&pipelinePlugin{}}
