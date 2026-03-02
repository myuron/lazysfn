// Package main is the entry point for lazysfn.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	sfnclient "github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/jroimartin/gocui"
	"github.com/myuron/lazysfn/internal/aws"
	"github.com/myuron/lazysfn/internal/config"
	"github.com/myuron/lazysfn/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}

	profiles, err := config.ParseProfiles(filepath.Join(home, ".aws", "config"))
	if err != nil {
		return fmt.Errorf("parsing aws config: %w", err)
	}

	app := ui.NewApp(profiles)
	spinnerDone := make(chan struct{})

	app.OnProfileSelected = func(g *gocui.Gui) error {
		selected := app.GetSelectedProfile()

		// Load AWS config with selected profile and region.
		opts := []func(*awsconfig.LoadOptions) error{
			awsconfig.WithSharedConfigProfile(selected.Name),
		}
		if selected.Region != "" {
			opts = append(opts, awsconfig.WithRegion(selected.Region))
		}
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			return app.ShowErrorModal(g, fmt.Sprintf("loading aws config: %v", err))
		}

		svc := &aws.Service{
			Client:         sfnclient.NewFromConfig(cfg),
			MaxConcurrency: 10,
		}

		// Transition to the main view with empty machines (spinner shows in right panel).
		app.SetLoading(true)
		if err := app.SetupMainView(g, nil); err != nil {
			return fmt.Errorf("setting up main view: %w", err)
		}

		// Wire up SM selection handler.
		app.OnSMSelect = func(arn string) {
			go func() {
				app.SetLoading(true)
				executions, err := svc.FetchExecutionHistory(context.Background(), arn)
				app.SetLoading(false)
				g.Update(func(g *gocui.Gui) error {
					if err != nil {
						return app.ShowErrorModal(g, fmt.Sprintf("loading executions: %v", err))
					}
					return app.RenderRightPanel(g, executions)
				})
			}()
		}

		// Start spinner ticker (runs until spinnerDone is closed).
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-spinnerDone:
					return
				case <-ticker.C:
					if app.IsLoading() {
						g.Update(app.AdvanceSpinner)
					}
				}
			}
		}()

		// Load state machines in the background.
		go func() {
			ctx := context.Background()

			machines, err := svc.FetchStateMachines(ctx)
			if err != nil {
				app.SetLoading(false)
				g.Update(func(g *gocui.Gui) error {
					return app.ShowErrorModal(g, fmt.Sprintf("loading state machines: %v", err))
				})
				return
			}

			// Show machine names in left panel (status not yet loaded).
			g.Update(func(g *gocui.Gui) error {
				app.SetMachines(machines)
				return app.RenderLeftPanel(g)
			})

			// Fetch latest execution status for each machine.
			machines, err = svc.FetchLatestStatus(ctx, machines)
			if err != nil {
				app.SetLoading(false)
				g.Update(func(g *gocui.Gui) error {
					return app.ShowErrorModal(g, fmt.Sprintf("loading status: %v", err))
				})
				return
			}

			app.SetLoading(false)
			g.Update(func(g *gocui.Gui) error {
				app.SetMachines(machines)
				if err := app.RenderLeftPanel(g); err != nil {
					return err
				}
				// Trigger initial execution history load for the selected SM.
				if arn := app.CurrentSMARN(); arn != "" {
					if fn := app.OnSMSelect; fn != nil {
						fn(arn)
					}
				}
				return nil
			})
		}()

		return nil
	}

	err = app.Run()
	close(spinnerDone)
	return err
}
