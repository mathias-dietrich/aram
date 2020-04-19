package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
        "github.com/rakyll/portmidi"
)

var usageStr = `
Usage: stan-pub [options] <subject> <message>
Options:
	-s,  --server   <url>            NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>   NATS Streaming cluster name
	-id, --clientid <client ID>      NATS Streaming client ID
	-a,  --async                     Asynchronous publish mode
	-cr, --creds    <credentials>    NATS 2.0 Credentials
`

// NOTE: Use tls scheme for TLS, e.g. stan-pub -s tls://demo.nats.io:4443 foo hello
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func main() {

	var (
		clusterID string
		clientID  string
		URL       string
		async     bool
		userCreds string
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clientID, "id", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&clientID, "clientid", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.BoolVar(&async, "a", false, "Publish asynchronously")
	flag.BoolVar(&async, "async", false, "Publish asynchronously")
	flag.StringVar(&userCreds, "cr", "", "Credentials File")
	flag.StringVar(&userCreds, "creds", "", "Credentials File")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Example Publisher")}
	// Use UserCredentials
	if userCreds != "" {
		opts = append(opts, nats.UserCredentials(userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()


	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	defer sc.Close()


    portmidi.Initialize()
	println("MIDI Devices")
	println(portmidi.CountDevices() )
	var deviceID = portmidi.DefaultInputDeviceID()
    println(portmidi.Info(deviceID))
    println("================")


	in, err := portmidi.NewInputStream(portmidi.DefaultInputDeviceID(), 1024)
		if err != nil {
			log.Fatal(err)
		}


for {
		result, err := in.Poll()

		if (err != nil) {
			log.Fatal(err)
		}


		if result {
			msgMidi, err := in.Read(1024)
			if err != nil {
				log.Fatal(err)
			}

			println(msgMidi[0].Status)
			println(msgMidi[0].Data1)
			println(msgMidi[0].Data2)

			var sendMsg = []byte { byte(msgMidi[0].Status), byte(msgMidi[0].Data1), byte(msgMidi[0].Data2)}

	
			fmt.Println("Send")



			//subj, msg := args[0], []byte(msgMidi.Data1)

			ch := make(chan bool)
			var glock sync.Mutex
			var guid string
			acb := func(lguid string, err error) {
				glock.Lock()
				log.Printf("Received ACK for guid %s\n", lguid)
				defer glock.Unlock()

				if err != nil {
					log.Fatalf("Error in server ack for guid %s: %v\n", lguid, err)
				}

				if lguid != guid {
					log.Fatalf("Expected a matching guid in ack callback, got %s vs %s\n", lguid, guid)
				}
			ch <- true
			}

		
			if !async {
				err = sc.Publish(args[0], sendMsg)
				if err != nil {
					log.Fatalf("Error during publish: %v\n", err)
				}
				
			} else {
				glock.Lock()
				guid, err = sc.PublishAsync(args[0], sendMsg, acb)
				if err != nil {
					log.Fatalf("Error during async publish: %v\n", err)
				}
				glock.Unlock()
				if guid == "" {
					log.Fatal("Expected non-empty guid to be returned.")
				}

				select {
					case <-ch:
						break
					case <-time.After(5 * time.Second):
						log.Fatal("timeout")
					}
				}

		} else {
			//fmt.Println("No new messages in the queue :(")
		}
	
	}
}


