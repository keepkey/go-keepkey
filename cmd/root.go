package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/keepkey/go-keepkey/pkg/keepkey"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var kk *keepkey.Keepkey

// debug level logging
var debug bool

// automatic button presses in debug mode
var debugButtonPress bool

// Button, Pin, and passphrase protection
var buttonProtection, pinProtection, passphraseProtection bool

// Pin cache / screensaver timeout
var autoLockDelayMs uint32

// setting label, pin, language
var label, pin, language string

// initialization vector for encryptKeyValue and decryptKeyValue
var initVector string

// BIP44 node path
var nodePath string

// Coin type i.e (Bitcoin, Ethereum)
var coinType string

var logger = log.New(ioutil.Discard, "", 0)

// Which device to send messages to if more than one is detected
var targetDevice string

// Root cobra CLI command
var rootCmd = &cobra.Command{
	Use:   "go-keepkey",
	Short: "go-keepkey is a CLI for interacting with keepkey devices",
	Long:  "go-keepkey is a CLI for interacting with keepkey devices",
}

func init() {
	// TODO: init on each subcommand instead
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Debug level logging")
	rootCmd.PersistentFlags().BoolVarP(&debugButtonPress, "autoButton", "", true, "Automatic button pressing if debug link is enabled")
	rootCmd.PersistentFlags().StringVarP(&targetDevice, "target", "", "", "Device label or HID serial to connect to if more than one device is connected")
	cobra.OnInitialize(connectDevice)
}

// Execute is the entry point to run the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Cleanup
	// TODO: should probably clean up other devices even if they weren't used
	if kk != nil {
		kk.Close()
	}
}

func connectDevice() {

	// TODO: add way to specify files as output not just Stdout
	if debug {
		logger = log.New(os.Stdout, "Log: ", 0)
	}

	kks, err := keepkey.GetDevicesWithConfig(&keepkey.Config{Logger: logger, AutoButton: debugButtonPress})
	if err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
	fmt.Println(color.GreenString(fmt.Sprintf("Connected to %d devices", len(kks))))
	for _, d := range kks {
		str := "  --  "
		if d.Serial() != "" {
			str += " Serial: " + d.Serial()
		}
		if d.Label() != "" {
			str += " Label: " + d.Label()
		}
		if d.ID() != "" {
			str += " ID: " + d.ID()
		}
		fmt.Println(str)
	}

	// Connect to the specified keepkey or if none is specified connect to the first device found
	if len(kks) > 1 && targetDevice == "" {
		yellow := color.New(color.FgYellow).Add(color.Bold).SprintFunc()
		fmt.Println(yellow("Multiple devices connected but none specified with -target flag, connecting to first device..."))
		kk = kks[0]
	} else if targetDevice != "" {
		for _, k := range kks {
			if k.Serial() == targetDevice || k.Label() == targetDevice {
				kk = k
				break
			}
			log.Fatal(color.RedString("No keepkey found with given label or serial"))
		}
	} else {
		// Connect to the first found keepkey
		kk = kks[0]
	}

	if debug {
		kk.SetLogger(log.New(os.Stdout, "DEBUG: ", log.Ltime))
	}
}
