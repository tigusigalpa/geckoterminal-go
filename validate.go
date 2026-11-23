package geckoterminal

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	// maxPage is the maximum page number accepted by the API.
	maxPage = 10
	// maxMultiAddresses is the maximum number of addresses accepted by
	// the "multi" pool/token endpoints and by the simple price endpoint.
	maxMultiAddresses = 30
	// maxOHLCVLimit is the maximum number of candles per OHLCV request.
	maxOHLCVLimit = 1000
	// maxTradesPerPage is the maximum per_page value for trades.
	maxTradesPerPage = 300
	// maxSearchQuery is the maximum search query length.
	maxSearchQuery = 256
)

// validateSegment validates a single path segment such as a network ID
// or a contract address.
func validateSegment(field, v string, maxLen int) error {
	if v == "" {
		return &ValidationError{Field: field, Value: v, Message: "must not be empty"}
	}
	if len(v) > maxLen {
		return &ValidationError{Field: field, Value: v, Message: fmt.Sprintf("must be at most %d characters", maxLen)}
	}
	if strings.ContainsAny(v, "/?# \t\n\r") {
		return &ValidationError{Field: field, Value: v, Message: "must not contain '/', '?', '#' or whitespace"}
	}
	return nil
}

// validateNetwork validates the network path parameter.
func validateNetwork(network string) error {
	return validateSegment("network", network, 64)
}

// validateAddress validates a pool or token address path parameter.
func validateAddress(field, addr string) error {
	return validateSegment(field, addr, 128)
}

// validateAddresses validates the address list of a multi endpoint and
// returns it as a comma-joined path fragment with each address
// individually escaped.
func validateAddresses(field string, addrs []string) (string, error) {
	if len(addrs) == 0 {
		return "", &ValidationError{Field: field, Value: "", Message: "must contain at least one address"}
	}
	if len(addrs) > maxMultiAddresses {
		return "", &ValidationError{
			Field:   field,
			Value:   strconv.Itoa(len(addrs)),
			Message: fmt.Sprintf("must contain at most %d addresses", maxMultiAddresses),
		}
	}
	escaped := make([]string, len(addrs))
	for i, a := range addrs {
		if err := validateAddress(field, a); err != nil {
			return "", err
		}
		escaped[i] = url.PathEscape(a)
	}
	return strings.Join(escaped, ","), nil
}

// setPage validates and applies a 1..10 page parameter.
func setPage(q url.Values, page *int) error {
	if page == nil {
		return nil
	}
	if *page < 1 || *page > maxPage {
		return &ValidationError{Field: "page", Value: strconv.Itoa(*page), Message: "must be between 1 and 10"}
	}
	q.Set("page", strconv.Itoa(*page))
	return nil
}

// setLimit validates and applies a bounded integer parameter.
func setLimit(q url.Values, field string, v *int, lo, hi int) error {
	if v == nil {
		return nil
	}
	if *v < lo || *v > hi {
		return &ValidationError{
			Field:   field,
			Value:   strconv.Itoa(*v),
			Message: fmt.Sprintf("must be between %d and %d", lo, hi),
		}
	}
	q.Set(field, strconv.Itoa(*v))
	return nil
}

// setEnum validates and applies a string enum parameter. The allowed set
// is endpoint specific.
func setEnum[T ~string](q url.Values, field string, v *T, allowed map[T]bool) error {
	if v == nil {
		return nil
	}
	if !allowed[*v] {
		vals := make([]string, 0, len(allowed))
		for a := range allowed {
			vals = append(vals, string(a))
		}
		sort.Strings(vals)
		return &ValidationError{
			Field:   field,
			Value:   string(*v),
			Message: "must be one of: " + strings.Join(vals, ", "),
		}
	}
	q.Set(field, string(*v))
	return nil
}

// setBool applies an optional boolean parameter as "true"/"false".
func setBool(q url.Values, field string, v *bool) {
	if v != nil {
		q.Set(field, strconv.FormatBool(*v))
	}
}

// setInclude validates an include list against the endpoint-specific
// allowed set, deduplicates it preserving order, and applies it as CSV.
func setInclude(q url.Values, include []Include, allowed map[Include]bool) error {
	if len(include) == 0 {
		return nil
	}
	seen := make(map[Include]bool, len(include))
	out := make([]string, 0, len(include))
	for _, inc := range include {
		if !allowed[inc] {
			return &ValidationError{
				Field:   "include",
				Value:   string(inc),
				Message: "not supported by this endpoint",
			}
		}
		if !seen[inc] {
			seen[inc] = true
			out = append(out, string(inc))
		}
	}
	q.Set("include", strings.Join(out, ","))
	return nil
}

// setTokenParam validates and applies the "token" inversion parameter,
// which is either "base", "quote" or a token address.
func setTokenParam(q url.Values, v *string) error {
	if v == nil {
		return nil
	}
	if *v != "base" && *v != "quote" {
		if err := validateAddress("token", *v); err != nil {
			return err
		}
	}
	q.Set("token", *v)
	return nil
}

// setString applies an optional non-empty string parameter.
func setString(q url.Values, field string, v *string) {
	if v != nil && *v != "" {
		q.Set(field, *v)
	}
}
