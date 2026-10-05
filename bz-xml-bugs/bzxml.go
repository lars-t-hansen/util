// Bzxml reads and parses XML-exported bugzilla data, then "does something" with it, see options.
// In the absence of options, prints every bug ID and title.
//
// Usage:
//
//	bzxml [options] filename.xml
//
// Options:
//
//	-sel n
//	  Print bug n
//
// Bzxml handles only the Bugzilla XML structure as created by our quite old bugzilla install.
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type timestamp string // "yyyy-mm-dd hh:mm:zz tz" it looks like

type Bugzilla struct {
	Bugs []BzBug `xml:"bug"`
}

type BzBug struct {
	BugId       uint      `xml:"bug_id"`
	Creation    timestamp `xml:"creation_ts"`
	ShortDesc   string    `xml:"short_desc"`
	Product     string    `xml:"product"`
	Component   string    `xml:"component"`
	RepPlatform string    `xml:"rep_platform"`
	BugStatus   string    `xml:"bug_status"`
	Resolution  string    `xml:"resolution"`
	Priority    string    `xml:"priority"`
	// there are many more fields
	Comments []BzLongDesc `xml:"long_desc"`
}

type BzLongDesc struct {
	CommentId    uint      `xml:"commentid"`
	CommentCount uint      `xml:"comment_count"`
	Who          Email     `xml:"who"`
	When         timestamp `xml:"bug_when"`
	TheText      string    `xml:"thetext"`
	// there are more fields
}

type Email struct {
	Name string `xml:"name,attr"`
	Addr string `xml:",chardata"`
}

var (
	selFlag = flag.Uint("sel", 0, "Select a specific bug")
)

func main() {
	rest := FlagParse("bzxml", "filename")
	infilename := rest[0]
	f, err := os.Open(infilename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	blob, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	var bz Bugzilla
	if err := xml.Unmarshal(blob, &bz); err != nil {
		log.Fatal(err)
	}

	if *selFlag > 0 {
		for _, b := range bz.Bugs {
			if b.BugId == *selFlag {
				fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
				fmt.Printf("  %s %s\n", b.Product, b.Component)
				for _, c := range b.Comments {
					fmt.Printf("-------\n%s <%s>\n%s\n", c.Who.Name, c.Who.Addr, c.TheText)
				}
				fmt.Printf("-------\n")
				break
			}
		}
	} else {
		// The default is to just print the ID and title
		for _, b := range bz.Bugs {
			fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
		}
	}
}
