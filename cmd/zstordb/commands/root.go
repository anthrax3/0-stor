package commands

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	badgerkv "github.com/dgraph-io/badger"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/zero-os/0-stor/cmd"
	"github.com/zero-os/0-stor/server/api"
	"github.com/zero-os/0-stor/server/api/grpc"
	"github.com/zero-os/0-stor/server/db/badger"
	"github.com/zero-os/0-stor/server/jwt"
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "zstordb",
	Short: "A generic object store server.",
	Long:  `A generic object store server used by zero-os.`,
	RunE:  rootFunc,
	PreRun: func(*cobra.Command, []string) {
		if rootCfg.DebugLog {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug logging enabled")
		}
		if rootCfg.AuthDisabled {
			log.Warning("!! Authentification disabled, don't use this mode for production!!!")
		}
	},
}

func rootFunc(*cobra.Command, []string) error {
	cmd.LogVersion()

	dbOpts := badgerkv.DefaultOptions
	dbOpts.SyncWrites = !rootCfg.AsyncWrite

	db, err := badger.NewWithOpts(rootCfg.DB.Dirs.Data, rootCfg.DB.Dirs.Meta, dbOpts)
	if err != nil {
		log.Errorf("error while opening database files: %v", err)
		return err
	}

	var storServer api.Server
	if rootCfg.AuthDisabled {
		storServer, err = grpc.New(db, nil, rootCfg.MaxMsgSize, rootCfg.JobCount)
	} else {
		storServer, err = grpc.New(db, jwt.DefaultVerifier(), rootCfg.MaxMsgSize, rootCfg.JobCount)
	}
	if err != nil {
		log.Errorf("error while creating database layer: %v", err)
		return err
	}
	defer func() {
		log.Println("Gracefully closing zstordb")
		storServer.Close()
	}()

	if rootCfg.ProfileAddress != "" {
		go func() {
			log.Infof("profiling enabled on %v", rootCfg.ProfileAddress)
			if err := http.ListenAndServe(rootCfg.ProfileAddress, http.DefaultServeMux); err != nil {
				log.Panicf("Failed to enable profiling on %v, err:%v", rootCfg.ProfileAddress, err)
			}
		}()
	}

	if rootCfg.ProfileMode.String() != "" {
		stat, err := os.Stat(rootCfg.ProfileOutput)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err := os.MkdirAll(rootCfg.ProfileOutput, 0660); err != nil {
				return fmt.Errorf("fail to create profile output directory: %v", err)
			}
		}
		if !stat.IsDir() {
			return fmt.Errorf("profile-output (%s) is not a directory", rootCfg.ProfileOutput)
		}

		switch rootCfg.ProfileMode {
		case cmd.ProfileModeCPU:
			defer profile.Start(
				profile.NoShutdownHook,
				profile.ProfilePath(rootCfg.ProfileOutput),
				profile.CPUProfile).Stop()
		case cmd.ProfileModeMem:
			defer profile.Start(
				profile.NoShutdownHook,
				profile.ProfilePath(rootCfg.ProfileOutput),
				profile.MemProfile).Stop()
		case cmd.ProfileModeTrace:
			defer profile.Start(
				profile.NoShutdownHook,
				profile.ProfilePath(rootCfg.ProfileOutput),
				profile.TraceProfile).Stop()
		case cmd.ProfileModeBlock:
			defer profile.Start(
				profile.NoShutdownHook,
				profile.ProfilePath(rootCfg.ProfileOutput),
				profile.BlockProfile).Stop()
		}
	}

	// create a TCP listener for our server
	listener, err := net.Listen("tcp", rootCfg.ListenAddress.String())
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	errChan := make(chan error, 1)

	go func() {
		err := storServer.Serve(listener)
		errChan <- err
	}()

	log.Infof("Server interface: grpc")
	log.Infof("Server listening on %s", listener.Addr().String())

	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		return nil
	}
}

var rootCfg struct {
	DebugLog       bool
	ListenAddress  cmd.ListenAddress
	ProfileAddress string
	ProfileMode    cmd.ProfileMode
	ProfileOutput  string
	AuthDisabled   bool
	MaxMsgSize     int
	AsyncWrite     bool
	JobCount       int

	DB struct {
		Dirs struct {
			Meta string
			Data string
		}
	}
}

func init() {
	rootCmd.AddCommand(cmd.VersionCmd)

	rootCmd.Flags().BoolVarP(
		&rootCfg.DebugLog, "debug", "D", false, "Enable debug logging.")
	rootCmd.Flags().VarP(
		&rootCfg.ListenAddress, "listen", "L", rootCfg.ListenAddress.Description())
	rootCmd.Flags().StringVar(
		&rootCfg.DB.Dirs.Data, "data-dir", ".db/data", "Directory path used to store the data.")
	rootCmd.Flags().StringVar(
		&rootCfg.DB.Dirs.Meta, "meta-dir", ".db/meta", "Directory path used to store the meta data.")
	rootCmd.Flags().StringVar(
		&rootCfg.ProfileAddress, "profile-addr", "", "Enables profiling of this server as an http service.")
	rootCmd.Flags().VarP(
		&rootCfg.ProfileMode, "profile-mode", "", rootCfg.ProfileMode.Description())
	rootCmd.Flags().StringVar(
		&rootCfg.ProfileOutput, "profile-output", ".", "Path of the directory where profiling files are written")
	rootCmd.Flags().BoolVar(
		&rootCfg.AuthDisabled, "no-auth", false, "Disable JWT authentication.")
	rootCmd.Flags().IntVar(
		&rootCfg.MaxMsgSize, "max-msg-size", 32, "Configure the maximum size of the message GRPC server can receive, in MiB")
	rootCmd.Flags().BoolVar(
		&rootCfg.AsyncWrite, "async-write", false, "Enable asynchronous writes in BadgerDB.")
	rootCmd.Flags().IntVarP(
		&rootCfg.JobCount, "jobs", "j", grpc.DefaultJobCount,
		"amount of async jobs to run for heavy GRPC server commands")
}
