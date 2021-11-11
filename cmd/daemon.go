
package cmd

import (
	"context"
	"time"

	//"flag"
	"fmt"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/examples/binapi/interfaces"
	"git.fd.io/govpp.git/examples/binapi/vpe"

	//logging "github.com/ipfs/go-log/v2"
	"github.com/srene/vpp-controller/common"
	"os"
	"os/signal"
	"log"

	"github.com/spf13/cobra"

	"git.fd.io/govpp.git"
	//"git.fd.io/govpp.git/adapter/socketclient"
)
var errors []error

//var log = logging.Logger("cmd")

/*var (
	sockAddr = flag.String("sock", socketclient.DefaultSocketName, "Path to VPP binary API socket file")
)*/

// getCmd represents the get command
func InitDaemonCmd(comm *common.Common) *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Run controller daemon",
		Long: ``,
		Run: func(cmd *cobra.Command, args []string) {
			datahopCli := `
 		 _     _____ _____ _   _    _____            _             _ _
		| |   |_   _/ ____| \ | |  / ____|          | |           | | |          
		| |__   | || |    |  \| | | |     ___  _ __ | |_ _ __ ___ | | | ___ _ __ 
		| '_ \  | || |    | .   | | |    / _ \| '_ \| __| '__/ _ \| | |/ _ \ '__|
		| | | |_| || |____| |\  | | |___| (_) | | | | |_| | | (_) | | |  __/ |
		|_| |_|_____\_____|_| \_|  \_____\___/|_| |_|\__|_|  \___/|_|_|\___|_|
	`
			conn, connEv, err := govpp.AsyncConnect("/run/vpp/api2.sock", core.DefaultMaxReconnectAttempts, core.DefaultReconnectInterval)
			if err != nil {
				log.Fatalln("ERROR:", err)
			}
			defer conn.Disconnect()

			// wait for Connected event
			select {
			case e := <-connEv:
				if e.State != core.Connected {
					log.Fatalln("ERROR: connecting to VPP failed:", e.Error)
				}
			}


			// check compatibility of used messages
			ch, err := conn.NewAPIChannel()

			//comm.Connection = conn
			comm.Channel = ch
			fmt.Println(datahopCli)
			fmt.Println("VPP controller daemon running")

			if err != nil {
				log.Fatalln("ERROR: creating channel failed:", err)
			}
			defer ch.Close()
			if err := ch.CheckCompatiblity(vpe.AllMessages()...); err != nil {
				log.Fatal(err)
			}
			if err := ch.CheckCompatiblity(interfaces.AllMessages()...); err != nil {
				log.Fatal(err)
			}

			// process errors encountered during the example
			defer func() {
				if len(errors) > 0 {
					fmt.Printf("finished with %d errors\n", len(errors))
					os.Exit(1)
				} else {
					fmt.Println("finished successfully")
				}
			}()

			// send and receive messages using stream (low-low level API)
			stream, err := conn.NewStream(context.Background(),
				core.WithRequestSize(50),
				core.WithReplySize(50),
				core.WithReplyTimeout(2*time.Second))
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := stream.Close(); err != nil {
					logError(err, "closing the stream")
				}
			}()

			var sigChan chan os.Signal
			sigChan = make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			for {
				select {
				case <-sigChan:
					//fmt.Println("cancel")
					comm.Cancel()
					return
				case <-comm.Context.Done():
					return
				}
			}
		},
	}
}
