// Services > UPNP
// This module provides UPNP port mapping functionality for routers, so that a node that is behind a router can still be accessed by other nodes.

package upnp

import (
	"aether-core/services/globals"
	"aether-core/services/logging"
	"fmt"
	extUpnp "github.com/NebulousLabs/go-upnp"
)

// var router *extUpnp.IGD
// var err error

func MapPort() {
	router, err := extUpnp.Discover()
	if err != nil {
		// Either could not be found, or connected to the internet directly.
		logging.Log(1, fmt.Sprintf("A router to port map could not be found. This computer could be directly connected to the Internet without a router. Error: %s", err.Error()))
		return
	}
	extIp, err2 := router.ExternalIP()
	if err2 != nil {
		// External IP finding failed.
		logging.Log(1, fmt.Sprintf("External IP of this machine could not be determined. Error: %s", err2.Error()))
	} else {
		globals.ExternalIp = extIp
		logging.Log(1, fmt.Sprintf("This computer's external IP is %s", globals.ExternalIp))
	}
	err3 := router.Forward(globals.AddressPort, "Aether")
	if err3 != nil {
		// Router is there, but port mapping failed.
		logging.Log(1, fmt.Sprintf("In an attempt to port map, the router was found, but the port mapping failed. Error: %s", err3.Error()))
	}
	logging.Log(1, fmt.Sprintf("Port mapping was successful. We mapped port %d to this computer.", globals.AddressPort))
}

// func UnmapPort() {
// 	// Attempt to remove the port mapping on quit.
// 	_ = router.Clear(globals.AddressPort)
// }
