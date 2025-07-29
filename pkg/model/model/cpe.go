package model

import (
	"fmt"
	"strings"

	"github.com/knqyf263/go-cpe/common"
	"github.com/knqyf263/go-cpe/naming"
	"github.com/praetorian-inc/tabularium/pkg/registry"
)

type CPE struct {
	registry.BaseModel
	Part      string `neo4j:"part" json:"part" desc:"CPE part (h=hardware, o=os, a=application)." example:"a"`
	Vendor    string `neo4j:"vendor" json:"vendor" desc:"Vendor name." example:"microsoft"`
	Product   string `neo4j:"product" json:"product" desc:"Product name." example:"windows_10"`
	Version   string `neo4j:"version" json:"version" desc:"Version identifier." example:"10.0.19042"`
	Update    string `neo4j:"update" json:"update" desc:"Update or service pack." example:"sp1"`
	Edition   string `neo4j:"edition" json:"edition" desc:"Edition information." example:"professional"`
	Language  string `neo4j:"language" json:"language" desc:"Language tag." example:"en-us"`
	SwEdition string `neo4j:"swEdition" json:"swEdition" desc:"Software edition." example:"home"`
	TargetSw  string `neo4j:"targetSw" json:"targetSw" desc:"Target software environment." example:"windows"`
	TargetHw  string `neo4j:"targetHw" json:"targetHw" desc:"Target hardware environment." example:"x64"`
	Other     string `neo4j:"other" json:"other" desc:"Other relevant information." example:"oem"`
}

func init() {
	registry.Registry.MustRegisterModel(&CPE{})
}

func NewCPE(cpe string) (CPE, error) {
	parsed, err := naming.UnbindFS(cpe)
	if err != nil {
		// The CPE string has too many segments, trying once again with extra segments removed
		if strings.Contains(err.Error(), "Found") && strings.Contains(err.Error(), "components in") {
			splitted := strings.Split(cpe, ":")
			merged := strings.Join(splitted[:13], ":")
			return NewCPE(merged)
		}
		return CPE{}, err
	}
	return newCPEFromWellFormedName(parsed), nil
}

func NewCPEFromURI(cpe string) (CPE, error) {
	parsed, err := naming.UnbindURI(cpe)
	if err != nil {
		return CPE{}, err
	}
	return newCPEFromWellFormedName(parsed), nil
}

func newCPEFromWellFormedName(wellFormedName common.WellFormedName) CPE {
	return CPE{
		Part:      wellFormedName.GetString("part"),
		Vendor:    wellFormedName.GetString("vendor"),
		Product:   wellFormedName.GetString("product"),
		Version:   wellFormedName.GetString("version"),
		Update:    wellFormedName.GetString("update"),
		Edition:   wellFormedName.GetString("edition"),
		Language:  wellFormedName.GetString("language"),
		SwEdition: wellFormedName.GetString("sw_edition"),
		TargetSw:  wellFormedName.GetString("target_sw"),
		TargetHw:  wellFormedName.GetString("target_hw"),
		Other:     wellFormedName.GetString("other"),
	}
}

func (c *CPE) handleLogicalValue(arg string) interface{} {
	lv, err := common.NewLogicalValue(arg)
	if err != nil {
		return arg
	}
	return lv
}

func (c *CPE) WellFormedName() common.WellFormedName {
	return common.WellFormedName{
		"part":       c.handleLogicalValue(c.Part),
		"vendor":     c.handleLogicalValue(c.Vendor),
		"product":    c.handleLogicalValue(c.Product),
		"version":    c.handleLogicalValue(c.Version),
		"update":     c.handleLogicalValue(c.Update),
		"edition":    c.handleLogicalValue(c.Edition),
		"language":   c.handleLogicalValue(c.Language),
		"sw_edition": c.handleLogicalValue(c.SwEdition),
		"target_sw":  c.handleLogicalValue(c.TargetSw),
		"target_hw":  c.handleLogicalValue(c.TargetHw),
		"other":      c.handleLogicalValue(c.Other),
	}
}

func (c *CPE) String() string {
	return naming.BindToFS(c.WellFormedName())
}

func (c *CPE) SearchQuery() string {
	// HACK: Setting version to - (or NA in NVD terms) should match to CPE without a version
	// for the technology, which will have JUST a product name in title and no version, etc.
	// saving us from doing some sketchy title string manipulation
	return fmt.Sprintf("cpe:2.3:%s:%s:%s:-", c.Part, c.Vendor, c.Product)
}

// GetDescription returns a description for the CPE model.
func (c *CPE) GetDescription() string {
	return "Represents a Common Platform Enumeration (CPE) identifier, used for naming hardware, software, and operating systems."
}
