// Package sec wraps the SEC EDGAR APIs used by company-goat-pp-cli.
//
// Two surfaces are used:
//   - https://efts.sec.gov/LATEST/search-index — full-text search across all
//     EDGAR filings; we filter forms=D to get private fundraising filings.
//   - https://www.sec.gov/Archives/edgar/data/<CIK>/<accession>/primary_doc.xml —
//     raw Form D XML for a specific filing, parsed for issuer + offering
//     fields per Form D XML Technical Specification v9.
//
// SEC EDGAR requires a descriptive User-Agent under fair-access policy. The
// constructor takes a contact email so users can identify themselves.
package sec

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client talks to SEC EDGAR. Construct with NewClient; pass a contact email
// so the User-Agent line identifies you as the SEC fair-access policy
// requires.
type Client struct {
	HTTP      *http.Client
	UserAgent string
}

// NewClient returns a Client with sensible defaults.
//
// SEC's fair-access policy requires the User-Agent to identify the user.
// Their accepted format is a plain "Name Email" string — embedding the
// URL or email inside parentheses gets the request blocked. We follow
// the simple two-token form they document.
//
// When contactEmail is empty, a placeholder is used. Most data.sec.gov
// endpoints will still serve under it, but EFTS (efts.sec.gov) is
// stricter and will return 403 — set COMPANY_PP_CONTACT_EMAIL.
func NewClient(contactEmail string) *Client {
	ua := "company-goat-pp-cli example@example.com"
	if contactEmail != "" {
		ua = "company-goat-pp-cli " + contactEmail
	}
	return &Client{
		HTTP:      &http.Client{Timeout: 15 * time.Second},
		UserAgent: ua,
	}
}

// SearchHit is one row from the EFTS full-text search response.
type SearchHit struct {
	Accession    string   `json:"accession"`     // e.g. 0002000934-23-000001
	CIKs         []string `json:"ciks"`          // 10-digit zero-padded CIKs
	DisplayNames []string `json:"display_names"` // e.g. ["Stripe Milton LLC  (CIK 0002000934)"]
	Form         string   `json:"form"`          // "D"
	FileDate     string   `json:"file_date"`     // YYYY-MM-DD
	BizStates    []string `json:"biz_states"`
	IncStates    []string `json:"inc_states"`
	Items        []string `json:"items"` // e.g. ["06B"] = exemption claimed
	PrimaryDoc   string   `json:"primary_doc"`
}

// SearchResponse is the EFTS JSON envelope.
type SearchResponse struct {
	Hits []SearchHit `json:"hits"`
	// Total hit count across all pages (Total > len(Hits) means paginated).
	Total int `json:"total"`
}

// SearchFormD runs a full-text search filtered to Form D filings.
//
// query is matched against issuer name and filing text. Results are sorted
// by relevance (best matches first). hitsPerPage caps each page (max 100
// per SEC, default 10). Returns at most hitsPerPage results.
func (c *Client) SearchFormD(ctx context.Context, query string, hitsPerPage int) (*SearchResponse, error) {
	if query == "" {
		return nil, errors.New("query is empty")
	}
	if hitsPerPage <= 0 {
		hitsPerPage = 10
	}
	if hitsPerPage > 100 {
		hitsPerPage = 100
	}

	u, _ := url.Parse("https://efts.sec.gov/LATEST/search-index")
	q := u.Query()
	q.Set("q", `"`+query+`"`)
	q.Set("forms", "D")
	q.Set("hits", strconv.Itoa(hitsPerPage))
	u.RawQuery = q.Encode()

	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	var raw struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string `json:"_id"` // accession:primary_doc_filename
				Source struct {
					CIKs         []string `json:"ciks"`
					DisplayNames []string `json:"display_names"`
					Form         string   `json:"form"`
					FileDate     string   `json:"file_date"`
					BizStates    []string `json:"biz_states"`
					IncStates    []string `json:"inc_states"`
					Items        []string `json:"items"`
					ADSH         string   `json:"adsh"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode efts response: %w", err)
	}

	out := &SearchResponse{Total: raw.Hits.Total.Value, Hits: make([]SearchHit, 0, len(raw.Hits.Hits))}
	for _, h := range raw.Hits.Hits {
		// _id is "0002000934-23-000001:primary_doc.xml" — split on colon.
		accession := h.Source.ADSH
		primaryDoc := ""
		if idx := strings.Index(h.ID, ":"); idx > 0 {
			primaryDoc = h.ID[idx+1:]
		}
		out.Hits = append(out.Hits, SearchHit{
			Accession:    accession,
			CIKs:         h.Source.CIKs,
			DisplayNames: h.Source.DisplayNames,
			Form:         h.Source.Form,
			FileDate:     h.Source.FileDate,
			BizStates:    h.Source.BizStates,
			IncStates:    h.Source.IncStates,
			Items:        h.Source.Items,
			PrimaryDoc:   primaryDoc,
		})
	}
	return out, nil
}

// FormD is the parsed shape of a Form D primary_doc.xml. Only fields we
// surface are populated — see Form D XML Technical Specification v9 for
// the full schema.
type FormD struct {
	CIK               string        `json:"cik"`
	EntityName        string        `json:"entity_name"`
	EntityType        string        `json:"entity_type"`
	State             string        `json:"state_of_inc"`
	YearOfInc         string        `json:"year_of_inc,omitempty"`
	IndustryGroup     string        `json:"industry_group"`
	OfferingAmount    int64         `json:"offering_amount"` // dollars; -1 if "Indefinite"
	AmountSold        int64         `json:"amount_sold"`     // dollars
	ExemptionsClaimed []string      `json:"exemptions_claimed,omitempty"`
	RelatedPersons    []FormDPerson `json:"related_persons,omitempty"`
	FilingDate        string        `json:"filing_date"` // populated from the search hit when fetched via FetchFormD
	Accession         string        `json:"accession"`
}

// FormDPerson is a related-person entry from Form D.
type FormDPerson struct {
	Name          string   `json:"name"`
	Relationships []string `json:"relationships"` // "Executive Officer", "Director", "Promoter"
}

// FetchFormD downloads and parses the Form D primary_doc.xml for the given
// CIK and accession (with dashes, e.g. 0002000934-23-000001).
func (c *Client) FetchFormD(ctx context.Context, cik, accession string) (*FormD, error) {
	if cik == "" || accession == "" {
		return nil, errors.New("cik and accession required")
	}
	cikInt, err := strconv.Atoi(strings.TrimLeft(cik, "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid cik %q: %w", cik, err)
	}
	dashless := strings.ReplaceAll(accession, "-", "")
	u := fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%d/%s/primary_doc.xml", cikInt, dashless)

	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var raw struct {
		XMLName       xml.Name `xml:"edgarSubmission"`
		PrimaryIssuer struct {
			CIK           string `xml:"cik"`
			EntityName    string `xml:"entityName"`
			EntityType    string `xml:"entityType"`
			IssuerAddress struct {
				StateOrCountry string `xml:"stateOrCountry"`
			} `xml:"issuerAddress"`
			JurisdictionOfInc string `xml:"jurisdictionOfInc"`
			YearOfInc         struct {
				YearOfIncOver5 string `xml:"yearOfIncOver5Years"`
				Year           string `xml:"value"`
			} `xml:"yearOfInc"`
		} `xml:"primaryIssuer"`
		OfferingData struct {
			IndustryGroup struct {
				IndustryGroupType string `xml:"industryGroupType"`
			} `xml:"industryGroup"`
			OfferingSalesAmounts struct {
				TotalOfferingAmount string `xml:"totalOfferingAmount"`
				TotalAmountSold     string `xml:"totalAmountSold"`
			} `xml:"offeringSalesAmounts"`
			TypesOfSecurities struct {
				IsEquityType bool `xml:"isEquityType"`
				IsDebtType   bool `xml:"isDebtType"`
			} `xml:"typesOfSecurities"`
			FederalExemptionsExclusions struct {
				ItemList []string `xml:"item"`
			} `xml:"federalExemptionsExclusions"`
		} `xml:"offeringData"`
		RelatedPersonsList struct {
			RelatedPersonInfo []struct {
				FirstName     string `xml:"relatedPersonName>firstName"`
				LastName      string `xml:"relatedPersonName>lastName"`
				Relationships struct {
					Relationship []string `xml:"relationship"`
				} `xml:"relatedPersonRelationshipList"`
			} `xml:"relatedPersonInfo"`
		} `xml:"relatedPersonsList"`
	}
	if err := xml.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode form D xml: %w", err)
	}

	fd := &FormD{
		CIK:           raw.PrimaryIssuer.CIK,
		EntityName:    strings.TrimSpace(raw.PrimaryIssuer.EntityName),
		EntityType:    raw.PrimaryIssuer.EntityType,
		State:         raw.PrimaryIssuer.JurisdictionOfInc,
		YearOfInc:     raw.PrimaryIssuer.YearOfInc.Year,
		IndustryGroup: raw.OfferingData.IndustryGroup.IndustryGroupType,
		Accession:     accession,
	}
	if v := strings.TrimSpace(raw.OfferingData.OfferingSalesAmounts.TotalOfferingAmount); v != "" {
		if v == "Indefinite" {
			fd.OfferingAmount = -1
		} else if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			fd.OfferingAmount = n
		}
	}
	if v := strings.TrimSpace(raw.OfferingData.OfferingSalesAmounts.TotalAmountSold); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			fd.AmountSold = n
		}
	}
	fd.ExemptionsClaimed = raw.OfferingData.FederalExemptionsExclusions.ItemList
	for _, rp := range raw.RelatedPersonsList.RelatedPersonInfo {
		name := strings.TrimSpace(rp.FirstName + " " + rp.LastName)
		if name == "" {
			continue
		}
		fd.RelatedPersons = append(fd.RelatedPersons, FormDPerson{
			Name:          name,
			Relationships: rp.Relationships.Relationship,
		})
	}
	return fd, nil
}

// SearchAndFetchAll searches Form D filings for query and fetches the
// parsed XML for up to maxFilings results. Each FormD has FilingDate set
// from the search hit. Errors fetching individual filings are skipped
// silently — callers see only successfully parsed filings.
func (c *Client) SearchAndFetchAll(ctx context.Context, query string, maxFilings int) ([]FormD, error) {
	if maxFilings <= 0 {
		maxFilings = 10
	}
	results, err := c.SearchFormD(ctx, query, maxFilings)
	if err != nil {
		return nil, err
	}
	out := make([]FormD, 0, len(results.Hits))
	for _, hit := range results.Hits {
		if len(hit.CIKs) == 0 {
			continue
		}
		fd, err := c.FetchFormD(ctx, hit.CIKs[0], hit.Accession)
		if err != nil {
			continue
		}
		fd.FilingDate = hit.FileDate
		out = append(out, *fd)
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json,text/xml,*/*")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("sec edgar %d: %s", resp.StatusCode, summary(body))
	}
	return body, nil
}

func summary(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
