// metadata.go — XML marshalling of the SP metadata document.
//
// The crewjam/saml library exposes an EntityDescriptor with
// xml-tagged fields, so a vanilla encoding/xml.MarshalIndent is
// enough to produce valid SAML 2.0 metadata.

package saml

import (
	"encoding/xml"
	"fmt"

	cs "github.com/crewjam/saml"
)

// marshalMetadata serialises the SP's EntityDescriptor as
// indented XML with the canonical XML declaration prepended.
func marshalMetadata(md *cs.EntityDescriptor) ([]byte, error) {
	body, err := xml.MarshalIndent(md, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("saml: marshal metadata: %w", err)
	}
	out := []byte(xml.Header)
	out = append(out, body...)
	return out, nil
}
