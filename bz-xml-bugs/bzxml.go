// Bzxml reads and parses XML-exported bugzilla data, then "does something" with it, see options.
// In the absence of options, prints every bug ID and title.
//
// Usage:
//
//	bzxml [options] filename.xml
//
// Options:
//
//		-sel n
//		  Print bug n
//
//	 -json
//	   Print out a JSON representation (a prototype)
//
// Bzxml handles the Bugzilla XML structure as created by an older bugzilla, as used by
// nordugrid/ARC.  The top-level structure is:
//
//	<bugzilla>
//	  <bug>
//	    <bug_id>...</bug_id>
//	    <short_desc>...</short_desc>
//	    <long_desc>
//	      <who>...</who>
//	      <thetext>...</thetext>
//	      ...
//	    </long_desc>
//	    ...
//	    <attachment>
//	      <desc>...</desc>
//	      ...
//	    </attachment>
//	    ...
//	    </bug>
//	    ...
//	</bugzilla>
//
// There are many fields.  See the code for the exact structure supported.
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// A number of string fields below could be custom types, notably Priority, Severity, and Keywords,
// but it hasn't been worth the bother yet.

type Bool bool

func (b *Bool) UnmarshalText(text []byte) error {
	*b = string(text) != "0"
	return nil
}

type Timestamp time.Time

func (t *Timestamp) UnmarshalText(text []byte) error {
	x, err := time.Parse("2006-01-02 15:04:05 -0700", string(text))
	if err != nil {
		return err
	}
	*t = Timestamp(x)
	return nil
}

func (t *Timestamp) Format(fmt string) string {
	return time.Time(*t).Format(fmt)
}

type Email struct {
	Name string `xml:"name,attr"`
	Addr string `xml:",chardata"`
}

type Bugzilla struct {
	Bugs []BzBug `xml:"bug"`
}

type BzBug struct {
	BugId              uint           `xml:"bug_id"`
	CreationTs         Timestamp      `xml:"creation_ts"`
	ShortDesc          string         `xml:"short_desc"`
	ReporterAccessible Bool           `xml:"reporter_accessible"`
	CCListAccessible   Bool           `xml:"cclist_accessible"`
	Classification     string         `xml:"classification"`
	Product            string         `xml:"product"`
	Component          string         `xml:"component"`
	Version            string         `xml:"version"`
	RepPlatform        string         `xml:"rep_platform"`
	OpSys              string         `xml:"op_sys"`
	BugStatus          string         `xml:"bug_status"`
	Resolution         string         `xml:"resolution"`
	BugFileLoc         string         `xml:"bug_file_loc"`
	StatusWhiteboard   string         `xml:"status_whiteboard"`
	Keywords           string         `xml:"keywords"`
	Priority           string         `xml:"priority"`
	BugSeverity        string         `xml:"bug_severity"`
	TargetMilestone    string         `xml:"target_milestone"`
	EverConfirmed      Bool           `xml:"everconfirmed"`
	Reporter           Email          `xml:"reporter"`
	AssignedTo         Email          `xml:"assigned_to"`
	CC                 []string       `xml:"cc"`
	EstimatedTime      float64        `xml:"estimated_time"`
	RemainingTime      float64        `xml:"remaining_time"`
	ActualTime         float64        `xml:"actual_time"`
	CfSvnRevision      string         `xml:"cf_svn_revision"`
	CfFixedIn          string         `xml:"cf_fixed_in"`
	Token              string         `xml:"token"`
	LongDescs          []BzLongDesc   `xml:"long_desc"`
	Attachments        []BzAttachment `xml:"attachment"`
}

type BzLongDesc struct {
	IsPrivate    Bool      `xml:"isprivate,attr"`
	CommentId    uint      `xml:"commentid"`
	CommentCount uint      `xml:"comment_count"`
	AttachId     uint      `xml:"attachid"`
	Who          Email     `xml:"who"`
	BugWhen      Timestamp `xml:"bug_when"`
	TheText      string    `xml:"thetext"`
}

type BzAttachment struct {
	AttachId uint      `xml:"attachid"`
	Date     Timestamp `xml:"date"`
	Desc     string    `xml:"desc"`
	Filename string    `xml:"filename"`
	Type     string    `xml:"type"`
	Size     uint      `xml:"size"`
	Attacher Email     `xml:"attacher"`
	Token    string    `xml:"token"`
}

var (
	selFlag  = flag.Uint("sel", 0, "Select a specific bug")
	jsonFlag = flag.Bool("json", false, "Print JSON (a prototype)")
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
				if *jsonFlag {
					jsonLongForm(b)
				} else {
					textLongForm(b)
				}
				break
			}
		}
	} else {
		for _, b := range bz.Bugs {
			if *jsonFlag {
				jsonShortForm(b)
			} else {
				textShortForm(b)
			}
		}
	}
}

func putJson(v any) {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stdout.Write(bs)
	if err != nil {
		log.Fatal(err)
	}
}

// The forms here are just meant as examples.
//
// Notably for the long form, we could add json tags to all fields on the BzBug and just print that,
// but that's unlikely to be right for any consumer.  So for now, just follow textLongForm and
// create a custom structure, this can then be adapted for other consumers.
//
// For the short form, it's just for debugging anyway so just print id and short description.

func textLongForm(b BzBug) {
	fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
	fmt.Printf("  %s / %s @ %s\n", b.Product, b.Component, b.CreationTs.Format(time.RFC3339))
	for _, c := range b.LongDescs {
		fmt.Printf("-------\n%s <%s> @ %s\n%s\n",
			c.Who.Name, c.Who.Addr, c.BugWhen.Format(time.RFC3339), c.TheText)
	}
	fmt.Printf("-------\n")
}

type jsonLong struct {
	Id         uint           `json:"id"`
	ShortDesc  string         `json:"short_desc"`
	Product    string         `json:"product"`
	Component  string         `json:"component"`
	CreationTs Timestamp      `json:"creation_time"`
	LongDescs  []jsonLongDesc `json:"long_descs,omitempty"`
}

type jsonLongDesc struct {
	Name string    `json:"Name"`
	Addr string    `json:"Addr"`
	When Timestamp `json:"when"`
	Text string    `json:"text"`
}

func jsonLongForm(b BzBug) {
	var v jsonLong
	v.Id = b.BugId
	v.ShortDesc = b.ShortDesc
	v.Product = b.Product
	v.Component = b.Component
	v.CreationTs = b.CreationTs
	for _, c := range b.LongDescs {
		var ld jsonLongDesc
		ld.Name = c.Who.Name
		ld.Addr = c.Who.Addr
		ld.When = c.BugWhen
		ld.Text = c.TheText
		v.LongDescs = append(v.LongDescs, ld)
	}
	putJson(&v)
}

func textShortForm(b BzBug) {
	fmt.Printf("%d  %s\n", b.BugId, b.ShortDesc)
}

type jsonShort struct {
	Id        uint   `json:"id"`
	ShortDesc string `json:"desc"`
}

func jsonShortForm(b BzBug) {
	var v jsonShort
	v.Id = b.BugId
	v.ShortDesc = b.ShortDesc
	putJson(&v)
}
