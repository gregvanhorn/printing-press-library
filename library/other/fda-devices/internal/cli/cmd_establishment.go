// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const pathEstablishment = "/device/registrationlisting.json"

func newEstablishmentCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "establishment",
		Short:       "Establishment registrations — search, list, get",
		Long:        "Query openFDA establishment registration & listings (/device/registrationlisting.json).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newEstablishmentSearchCmd(flags), newEstablishmentGetCmd(flags), newEstablishmentListCmd(flags))
	return cmd
}

func newEstablishmentSearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/registrationlisting.json",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, pathEstablishment, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func newEstablishmentGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <registration-number>",
		Short:       "Fetch one establishment by registration number",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, pathEstablishment, "registration.registration_number", args[0])
		},
	}
}

func newEstablishmentListCmd(flags *rootFlags) *cobra.Command {
	var (
		state   string
		country string
		q       openfda.Query
		all     bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List establishments with curated filters",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			clauses := []string{}
			if state != "" {
				clauses = append(clauses, fmt.Sprintf("registration.us_state_code:%q", state))
			}
			if country != "" {
				clauses = append(clauses, fmt.Sprintf("registration.iso_country_code:%q", country))
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			return runOpenFDAList(cmd, flags, pathEstablishment, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&state, "state", "", "US state code (e.g. CA)")
	cmd.Flags().StringVar(&country, "country", "", "ISO country code (e.g. US)")
	return cmd
}
