package main

import "github.com/tgbot-upnp/tgbot-upnp/app"

func main() {
	a := app.New()
	defer a.Wait()
	select {}
}
