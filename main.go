package main

import (
	"github.com/mager/caffy-beans/bundlefx"
	"github.com/mager/caffy-beans/httphandler"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		bundlefx.Module,
		fx.Invoke(httphandler.New),
	).Run()
}
